package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lagosproject/StockTaggerIA/src/gemini"
	"github.com/lagosproject/StockTaggerIA/src/i18n"
	"github.com/lagosproject/StockTaggerIA/src/interactive"
	"github.com/lagosproject/StockTaggerIA/src/metadata"
	"github.com/lagosproject/StockTaggerIA/src/presets"
)

const AppVersion = "1.0.0"

func main() {
	// Detect default directories
	targetDir := "."
	if envDir := os.Getenv("DEFAULT_PHOTOS_DIR"); envDir != "" {
		if info, err := os.Stat(envDir); err == nil && info.IsDir() {
			targetDir = envDir
		}
	}

	// Detect default JSON path
	jsonFile := "new_tags.json"
	if _, err := os.Stat(jsonFile); err != nil {
		if exePath, err := os.Executable(); err == nil {
			cand := filepath.Join(filepath.Dir(exePath), jsonFile)
			if _, err := os.Stat(cand); err == nil {
				jsonFile = cand
			}
		}
	}

	// CLI Flags Definition
	platform := flag.String("p", "stock", "Stock platform preset: stock (Unsplash & Pexels), pixabay")
	platformLong := flag.String("platform", "", "Stock platform preset (alias for -p)")
	jsonPath := flag.String("j", jsonFile, "Path to JSON metadata file")
	jsonPathLong := flag.String("json", "", "Path to JSON metadata file (alias for -j)")
	dirPath := flag.String("d", "", "Root folder containing photos to tag")
	dirPathLong := flag.String("dir", "", "Root folder (alias for -d)")
	langFlag := flag.String("l", "", "Metadata language override: en, es, fr")
	langLong := flag.String("lang", "", "Metadata language override (alias for -l)")
	useGemini := flag.Bool("gemini", false, "Use Google Gemini Vision AI to analyze images directly")
	apiKeyFlag := flag.String("api-key", "", "Google Gemini API Key (or set GEMINI_API_KEY environment variable)")
	dryRun := flag.Bool("dry-run", false, "Simulate execution without modifying any files on disk")
	interactiveFlag := flag.Bool("i", false, "Launch interactive configuration wizard")
	interactiveLong := flag.Bool("interactive", false, "Launch interactive configuration wizard (alias for -i)")
	versionFlag := flag.Bool("v", false, "Print version and exit")
	versionLong := flag.Bool("version", false, "Print version (alias for -v)")

	flag.Usage = func() {
		fmt.Println("StockTaggerIA - Stock Photography SEO Metadata & Tagging Tool v" + AppVersion)
		fmt.Println("\nUsage:")
		fmt.Println("  stocktaggeria [options]")
		fmt.Println("  stocktaggeria -i                # Launch localized interactive wizard")
		fmt.Println("\nOptions:")
		fmt.Println("  -p, --platform <name>   Stock preset: 'stock' (Unsplash & Pexels, max 30 tags) or 'pixabay' (max 25 tags) (default: stock)")
		fmt.Println("  -l, --lang <lang>       Metadata language: 'en' (English), 'es' (Spanish), 'fr' (French)")
		fmt.Println("  -d, --dir <path>        Root directory containing photos")
		fmt.Println("  -j, --json <path>       Path to JSON metadata file (default: new_tags.json)")
		fmt.Println("      --gemini            Analyze photos directly using Gemini Vision AI")
		fmt.Println("      --api-key <key>     Gemini API Key (default: GEMINI_API_KEY env)")
		fmt.Println("      --dry-run           Simulate execution without modifying files")
		fmt.Println("  -i, --interactive       Launch interactive terminal wizard (EN/ES/FR)")
		fmt.Println("  -v, --version           Display version information")
		fmt.Println("\nExamples:")
		fmt.Println("  # Launch interactive wizard:")
		fmt.Println("  stocktaggeria")
		fmt.Println("\n  # Automated offline tagging with JSON:")
		fmt.Println("  stocktaggeria -p stock -l en -d /photos --dry-run")
		fmt.Println("\n  # Direct AI tagging with Gemini:")
		fmt.Println("  stocktaggeria --gemini -p stock -l es -d /photos")
	}

	flag.Parse()

	if *versionFlag || *versionLong {
		fmt.Printf("StockTaggerIA version %s\n", AppVersion)
		return
	}

	// Check if run without arguments or with interactive flag -> launch wizard
	if len(os.Args) == 1 || *interactiveFlag || *interactiveLong {
		initialDir := targetDir
		if *dirPath != "" {
			initialDir = *dirPath
		} else if *dirPathLong != "" {
			initialDir = *dirPathLong
		}
		if err := interactive.RunWizard(initialDir, jsonFile); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Resolve flags
	activePlatformName := *platform
	if *platformLong != "" {
		activePlatformName = *platformLong
	}

	activeJson := *jsonPath
	if *jsonPathLong != "" {
		activeJson = *jsonPathLong
	}

	activeDir := targetDir
	if *dirPath != "" {
		activeDir = *dirPath
	} else if *dirPathLong != "" {
		activeDir = *dirPathLong
	}

	activeLang := *langFlag
	if *langLong != "" {
		activeLang = *langLong
	}

	preset := presets.ResolvePreset(activePlatformName)
	finalLang := presets.ResolveLanguage(preset, activeLang)
	msg := i18n.Get(i18n.DetectSystemLanguage())

	// Execute Headless Mode
	var metadataMap map[string]metadata.MetadataRaw

	if *useGemini {
		apiKey := *apiKeyFlag
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
		if apiKey == "" {
			fmt.Println("[ERROR] Gemini API Key is required when using --gemini. Pass --api-key or set GEMINI_API_KEY.")
			os.Exit(1)
		}

		images := metadata.ScanImages(activeDir)
		if len(images) == 0 {
			fmt.Printf("[WARN] No images found in directory: %s\n", activeDir)
			return
		}

		ctx := context.Background()
		client := gemini.NewClient(apiKey, "")
		resolvedModel := client.EnsureModel(ctx)
		client.PromptFallback = func(currentModel, altModel, reason string) bool {
			fmt.Printf(msg.ModelUnavailablePrompt, currentModel, reason, altModel)
			var ans string
			fmt.Scanln(&ans)
			ans = strings.ToLower(strings.TrimSpace(ans))
			return ans == "" || ans == "s" || ans == "y" || ans == "yes" || ans == "si" || ans == "o" || ans == "oui"
		}
		fmt.Printf("--> Analyzing %d images with Gemini Vision AI (%s)...\n", len(images), resolvedModel)

		var generated []gemini.MetadataEntry
		var rawEntries []metadata.MetadataRaw

		for i, imgPath := range images {
			rel := filepath.Base(imgPath)
			fmt.Printf("[%d/%d] Analyzing %s...\n", i+1, len(images), rel)
			entry, err := client.TagImage(ctx, imgPath, finalLang)
			if err != nil {
				fmt.Printf("[WARN] Failed analyzing %s: %v\n", rel, err)
				continue
			}
			generated = append(generated, *entry)
			rawEntries = append(rawEntries, entry.ToMetadataRaw())
		}

		if len(generated) > 0 {
			backupPath, err := gemini.SaveBackupJSON(generated, activeDir)
			if err == nil {
				fmt.Printf("--> Backup metadata saved to: %s\n", backupPath)
			}
		}

		metadataMap = metadata.MetadataMapFromRawEntries(rawEntries)
	} else {
		mMap, err := metadata.LoadMetadataMapFromJSON(activeJson)
		if err != nil {
			fmt.Printf("[ERROR] Failed to load JSON metadata '%s': %v\n", activeJson, err)
			os.Exit(1)
		}
		metadataMap = mMap
	}

	metadata.UpdateImageMetadata(metadataMap, activeDir, preset, finalLang, *dryRun, msg)
}
