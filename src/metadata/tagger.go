package metadata

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lagosproject/StockTaggerIA/src/i18n"
	"github.com/lagosproject/StockTaggerIA/src/presets"
)

// MetadataRaw holds JSON items with flexible string/array fields.
type MetadataRaw struct {
	Filename      string      `json:"filename"`
	Tags          interface{} `json:"tags"`
	TagsEn        interface{} `json:"tags_en"`
	TagsEs        interface{} `json:"tags_es"`
	TagsFr        interface{} `json:"tags_fr"`
	Description   interface{} `json:"description"`
	DescriptionEn interface{} `json:"description_en"`
	DescriptionEs interface{} `json:"description_es"`
	DescriptionFr interface{} `json:"description_fr"`
}

// SupportedImageExtensions defines the image file extensions recognized by StockTaggerIA.
var SupportedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".tif":  true,
	".tiff": true,
}

// ScanImages traverses root recursively and returns paths of supported images.
func ScanImages(dir string) []string {
	var images []string
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if SupportedImageExtensions[ext] {
			images = append(images, path)
		}
		return nil
	})
	return images
}

// cleanDroppedPath normalizes a single raw dropped path, handling file:// URIs,
// URL percent-encoding, quotes, and platform-specific network path conventions.
func cleanDroppedPath(raw string) string {
	raw = strings.TrimSpace(raw)
	// Strip outer quotes if present
	if (strings.HasPrefix(raw, "'") && strings.HasSuffix(raw, "'")) ||
		(strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"")) {
		if len(raw) >= 2 {
			raw = raw[1 : len(raw)-1]
		}
	}
	raw = strings.TrimSpace(raw)

	// Handle file:// URI scheme (e.g. dropped from file managers or browsers)
	if strings.HasPrefix(strings.ToLower(raw), "file://") {
		if unescaped, err := url.PathUnescape(raw); err == nil {
			raw = unescaped
		}
		raw = strings.TrimPrefix(raw, "file://")
		raw = strings.TrimPrefix(raw, "localhost")

		// On Windows: file:///C:/path -> C:\path or file://server/share -> \\server\share
		if runtime.GOOS == "windows" {
			if strings.HasPrefix(raw, "/") && len(raw) >= 3 && raw[2] == ':' {
				raw = raw[1:]
			} else if !strings.HasPrefix(raw, `\\`) && strings.HasPrefix(raw, `//`) {
				raw = `\\` + strings.TrimPrefix(raw, `//`)
			}
		}
	}

	if raw == "" {
		return ""
	}

	return filepath.Clean(raw)
}

// ParseDroppedPaths splits terminal input string that may contain one or more
// drag-and-dropped file or folder paths, handling single/double quotes, network paths and escaped spaces.
func ParseDroppedPaths(input string) []string {
	var paths []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	isEscaped := false
	isWindows := runtime.GOOS == "windows"

	input = strings.TrimSpace(input)
	for i := 0; i < len(input); i++ {
		ch := input[i]

		if isEscaped {
			current.WriteByte(ch)
			isEscaped = false
			continue
		}

		// On non-Windows platforms, backslash escapes spaces and special chars
		if ch == '\\' && !isWindows && !inSingleQuote {
			isEscaped = true
			continue
		}

		if ch == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			continue
		}

		if ch == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			continue
		}

		if (ch == ' ' || ch == '\t') && !inSingleQuote && !inDoubleQuote {
			if current.Len() > 0 {
				if cleaned := cleanDroppedPath(current.String()); cleaned != "" {
					paths = append(paths, cleaned)
				}
				current.Reset()
			}
			continue
		}

		current.WriteByte(ch)
	}

	if current.Len() > 0 {
		if cleaned := cleanDroppedPath(current.String()); cleaned != "" {
			paths = append(paths, cleaned)
		}
	}

	return paths
}

// ResolveCandidateImages resolves input paths (files or directories) into a deduplicated list of supported images.
func ResolveCandidateImages(paths []string) []string {
	var images []string
	seen := make(map[string]bool)

	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			for _, img := range ScanImages(p) {
				cleanPath := filepath.Clean(img)
				if !seen[cleanPath] {
					seen[cleanPath] = true
					images = append(images, cleanPath)
				}
			}
		} else {
			ext := strings.ToLower(filepath.Ext(p))
			if SupportedImageExtensions[ext] {
				cleanPath := filepath.Clean(p)
				if !seen[cleanPath] {
					seen[cleanPath] = true
					images = append(images, cleanPath)
				}
			}
		}
	}
	return images
}

// Stats tracks scanning and execution metrics.
type Stats struct {
	FilesScanned  int
	MatchesFound  int
	Updated       int
	SkippedDryRun int
	Errors        int
	ErrorDetails  [][2]string
}

// ParseStringOrSlice converts string (comma-separated) or array of strings into a slice.
func ParseStringOrSlice(val interface{}) []string {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case string:
		return cleanStringSlice(strings.Split(v, ","))
	case []interface{}:
		var res []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				if t := strings.TrimSpace(s); t != "" {
					res = append(res, t)
				}
			}
		}
		return res
	case []string:
		return cleanStringSlice(v)
	}
	return nil
}

func cleanStringSlice(items []string) []string {
	var res []string
	for _, s := range items {
		if t := strings.TrimSpace(s); t != "" {
			res = append(res, t)
		}
	}
	return res
}

// ParseString extracts a clean string from an interface.
func ParseString(val interface{}) string {
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprintf("%v", val))
}

func resolveFirstTags(candidates ...interface{}) []string {
	for _, c := range candidates {
		if tags := ParseStringOrSlice(c); len(tags) > 0 {
			return tags
		}
	}
	return nil
}

func resolveFirstDescription(candidates ...interface{}) string {
	for _, c := range candidates {
		if desc := ParseString(c); desc != "" {
			return desc
		}
	}
	return ""
}

// ParseMetadataEntry extracts, deduplicates, and limits tags and description for a target language.
func ParseMetadataEntry(item MetadataRaw, language string, maxTags int) ([]string, string) {
	var rawTags []string
	var description string

	switch strings.ToLower(language) {
	case "fr":
		rawTags = resolveFirstTags(item.TagsFr, item.TagsEn, item.TagsEs, item.Tags)
		description = resolveFirstDescription(item.DescriptionFr, item.DescriptionEn, item.DescriptionEs, item.Description)
	case "es":
		rawTags = resolveFirstTags(item.TagsEs, item.Tags, item.TagsEn, item.TagsFr)
		description = resolveFirstDescription(item.DescriptionEs, item.Description, item.DescriptionEn, item.DescriptionFr)
	default: // English ("en")
		rawTags = resolveFirstTags(item.TagsEn, item.Tags, item.TagsEs, item.TagsFr)
		description = resolveFirstDescription(item.DescriptionEn, item.Description, item.DescriptionEs, item.DescriptionFr)
	}

	// Case-insensitive deduplication while preserving original order
	seen := make(map[string]bool)
	var tagList []string
	for _, t := range rawTags {
		low := strings.ToLower(t)
		if !seen[low] {
			seen[low] = true
			tagList = append(tagList, t)
		}
	}

	if maxTags > 0 && len(tagList) > maxTags {
		tagList = tagList[:maxTags]
	}

	return tagList, description
}

// LoadMetadataMapFromJSON parses a JSON file into a filename-keyed map.
func LoadMetadataMapFromJSON(jsonPath string) (map[string]MetadataRaw, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}
	var rawEntries []MetadataRaw
	if err := json.Unmarshal(data, &rawEntries); err != nil {
		return nil, err
	}
	return MetadataMapFromRawEntries(rawEntries), nil
}

// MetadataMapFromRawEntries converts a slice of entries into a map keyed by lowercase base filename.
func MetadataMapFromRawEntries(entries []MetadataRaw) map[string]MetadataRaw {
	res := make(map[string]MetadataRaw, len(entries))
	for _, item := range entries {
		if item.Filename != "" {
			clean := strings.ToLower(filepath.Base(item.Filename))
			res[clean] = item
		}
	}
	return res
}

// UpdateImageMetadata writes keywords and descriptions to image files in rootFolder using ExifTool.
func UpdateImageMetadata(
	metadataMap map[string]MetadataRaw,
	rootFolder string,
	preset presets.Preset,
	lang string,
	dryRun bool,
	msg i18n.Messages,
) Stats {
	images := ScanImages(rootFolder)
	if len(images) == 0 {
		if info, err := os.Stat(rootFolder); err == nil && !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(rootFolder))
			if SupportedImageExtensions[ext] {
				images = []string{rootFolder}
			}
		}
	}
	return UpdateImagesMetadata(metadataMap, images, rootFolder, preset, lang, dryRun, msg)
}

// UpdateImagesMetadata writes keywords and descriptions to the specified image list using ExifTool.
func UpdateImagesMetadata(
	metadataMap map[string]MetadataRaw,
	images []string,
	displayTarget string,
	preset presets.Preset,
	lang string,
	dryRun bool,
	msg i18n.Messages,
) Stats {
	stats := Stats{}
	modeLabel := "[LIVE]"
	if dryRun {
		modeLabel = "[DRY RUN]"
	}

	limitStr := "All (Unlimited)"
	if preset.MaxTags > 0 {
		limitStr = fmt.Sprintf("%d tags", preset.MaxTags)
	}

	targetDesc := displayTarget
	if abs, err := filepath.Abs(displayTarget); err == nil {
		targetDesc = abs
	}

	fmt.Printf("\n--- %s %s ---\n", msg.BannerSubtitle, modeLabel)
	fmt.Printf("Platform Preset  : %s (%s)\n", preset.Name, preset.Summary)
	fmt.Printf("Metadata Language: %s\n", strings.ToUpper(lang))
	fmt.Printf("Max Tags Allowed : %s\n", limitStr)
	fmt.Printf("Target / Context : %s\n\n", targetDesc)

	var session *ExifToolSession
	var exiftoolPath string
	if !dryRun {
		etPath, err := FindExifTool()
		if err != nil {
			fmt.Printf("[ERROR] %s (%v)\n", msg.ExifToolNotFound, err)
			stats.Errors++
			return stats
		}
		exiftoolPath = etPath
		sess, err := StartExifToolSession(exiftoolPath)
		if err != nil {
			fmt.Printf("[WARN] Failed starting ExifTool in persistent stay-open mode (%v). Falling back to direct calls.\n", err)
		} else {
			session = sess
			defer session.Close()
		}
	}

	for _, path := range images {
		stats.FilesScanned++
		lookupKey := strings.ToLower(filepath.Base(path))

		itemData, exists := metadataMap[lookupKey]
		if !exists {
			continue
		}

		stats.MatchesFound++
		tagList, description := ParseMetadataEntry(itemData, lang, preset.MaxTags)

		descPreview := " | No description"
		if description != "" {
			if len(description) > 50 {
				descPreview = fmt.Sprintf(" | Desc: '%s...'", description[:50])
			} else {
				descPreview = fmt.Sprintf(" | Desc: '%s'", description)
			}
		}

		if dryRun {
			fmt.Printf("[DRY RUN] Would update: %s (%d tags%s)\n", path, len(tagList), descPreview)
			if description != "" {
				fmt.Printf("          Description: %s\n", description)
			}
			fmt.Printf("          Tags: %v\n", tagList)
			stats.SkippedDryRun++
			continue
		}

		// Build ExifTool argument list with preallocated capacity
		args := make([]string, 0, 10+2*len(tagList))
		args = append(args,
			"-overwrite_original",
			"-charset", "filename=utf8",
			"-codedcharacterset=utf8",
			"-XMP-dc:Subject=",
		)
		for _, tag := range tagList {
			args = append(args, fmt.Sprintf("-XMP-dc:Subject=%s", tag))
		}

		args = append(args, "-IPTC:Keywords=")
		for _, tag := range tagList {
			args = append(args, fmt.Sprintf("-IPTC:Keywords=%s", tag))
		}

		args = append(args, fmt.Sprintf("-EXIF:XPKeywords=%s", strings.Join(tagList, ", ")))

		if description != "" {
			args = append(args,
				fmt.Sprintf("-XMP-dc:Description=%s", description),
				fmt.Sprintf("-IPTC:Caption-Abstract=%s", description),
				fmt.Sprintf("-EXIF:ImageDescription=%s", description),
			)
		}

		args = append(args, path)

		var opErr error
		if session != nil {
			_, opErr = session.Execute(args)
		} else {
			cmd := exec.Command(exiftoolPath, args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				opErr = fmt.Errorf("%v: %s", err, string(out))
			}
		}

		if opErr != nil {
			stats.Errors++
			stats.ErrorDetails = append(stats.ErrorDetails, [2]string{path, opErr.Error()})
			fmt.Printf("[ERROR] Failed writing metadata for %s: %v\n", path, opErr)
		} else {
			stats.Updated++
			fmt.Printf("[UPDATED] (%d tags%s) -> %s\n", len(tagList), descPreview, path)
		}
	}

	fmt.Println("\n========================================")
	fmt.Printf("       %s\n", fmt.Sprintf(msg.SummaryTitle, modeLabel))
	fmt.Println("========================================")
	fmt.Printf(msg.SummaryPreset+"\n", preset.Name)
	fmt.Printf(msg.SummaryMetaLang+"\n", strings.ToUpper(lang))
	fmt.Printf(msg.SummaryScanned+"\n", stats.FilesScanned)
	fmt.Printf(msg.SummaryMatches+"\n", stats.MatchesFound)
	if dryRun {
		fmt.Printf(msg.SummaryProjected+"\n", stats.SkippedDryRun)
	} else {
		fmt.Printf(msg.SummaryUpdated+"\n", stats.Updated)
	}
	fmt.Printf(msg.SummaryErrors+"\n", stats.Errors)
	fmt.Println("========================================")

	return stats
}
