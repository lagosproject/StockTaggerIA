package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultModel = "gemini-3.6-flash"
	APIBaseURL   = "https://generativelanguage.googleapis.com/v1beta/models"
	MaxDimension = 1500
)

// BuildSystemPrompt constructs the Gemini prompt. By default, metadata is generated
// in English (stock standard). If targetLang is set to another supported language (e.g. "es", "fr", or "all"),
// it instructs Gemini to generate metadata in English AND in the selected secondary language(s).
func BuildSystemPrompt(targetLang string) string {
	targetLang = strings.ToLower(strings.TrimSpace(targetLang))
	switch targetLang {
	case "", "en":
		return `Role: You are an expert stock photography metadata specialist and SEO copywriter optimized for Pexels, Unsplash, and Pixabay.

Task: Analyze the visual content of the uploaded image and output a valid JSON object containing high-ranking SEO descriptions and tag sets in English.
Use the uploaded image's exact file name to map the entry.
Output strictly valid JSON with no introductory text, surrounding commentary, or conversational filler.

Metadata Rules:
- Description (description_en):
  Concise, highly descriptive, natural-sounding title/description (1-2 sentences) in English.
  Lead with the primary subject and key action or setting to capture top search rankings across stock platforms.

- SEO Tags (tags_en):
  Generate up to 30 tags in English.
  Relevance & Search Intent: Target core subjects, specific actions, visual environment, lighting (e.g., natural light, golden hour), composition (e.g., close up, low angle, overhead view), color palette, and mood/emotion.
  Literal & Conceptual Mix: Combine direct physical descriptors with abstract search concepts (e.g., nostalgia, freedom, classic elegance, urban lifestyle).
  Platform Compliance: Avoid spammy, repetitive, or misleading keywords.
  Formatting: Format tags as a single string, with individual keywords separated by commas.

JSON Schema:
{
  "filename": "<exact_filename>",
  "description_en": "Concise English description...",
  "tags_en": "tag1, tag2, tag3..."
}`

	case "all", "trilingual", "es,fr", "fr,es":
		return `Role: You are an expert stock photography metadata specialist and SEO copywriter optimized for Pexels, Unsplash, and Pixabay.

Task: Analyze the visual content of the uploaded image and output a valid JSON object containing high-ranking SEO descriptions and tag sets.
Provide titles/descriptions and tag lists in English (primary) as well as Spanish and French (secondary).
Use the uploaded image's exact file name to map the entry.
Output strictly valid JSON with no introductory text, surrounding commentary, or conversational filler.

Metadata Rules:
- Descriptions (description_en, description_es, description_fr):
  Concise, highly descriptive, natural-sounding titles/descriptions (1-2 sentences) per language.
  Lead with the primary subject and key action or setting to capture top search rankings across stock platforms.

- SEO Tags (tags_en, tags_es, tags_fr):
  Generate up to 30 tags per language per image.
  Relevance & Search Intent: Target core subjects, specific actions, visual environment, lighting, composition, color palette, and mood/emotion.
  Literal & Conceptual Mix: Combine direct physical descriptors with abstract search concepts.
  Platform Compliance: Avoid spammy, repetitive, or misleading keywords.
  Formatting: Format tags as a single string per key, with individual keywords separated by commas.

JSON Schema:
{
  "filename": "<exact_filename>",
  "description_en": "Concise English description...",
  "description_es": "Descripción concisa en español...",
  "description_fr": "Description concise en français...",
  "tags_en": "tag1, tag2, tag3...",
  "tags_es": "etiqueta1, etiqueta2, etiqueta3...",
  "tags_fr": "balise1, balise2, balise3..."
}`

	case "fr":
		return fmt.Sprintf(`Role: You are an expert stock photography metadata specialist and SEO copywriter optimized for Pexels, Unsplash, and Pixabay.

Task: Analyze the visual content of the uploaded image and output a valid JSON object containing high-ranking SEO descriptions and tag sets.
Primary language is English, with French as secondary language.
Use the uploaded image's exact file name to map the entry.
Output strictly valid JSON with no introductory text, surrounding commentary, or conversational filler.

Metadata Rules:
- Descriptions (description_en, description_fr):
  Concise, highly descriptive, natural-sounding titles/descriptions (1-2 sentences) in English and French.
  Lead with the primary subject and key action or setting to capture top search rankings across stock platforms.

- SEO Tags (tags_en, tags_fr):
  Generate up to 30 tags per language per image.
  Relevance & Search Intent: Target core subjects, specific actions, visual environment, lighting, composition, color palette, and mood/emotion.
  Literal & Conceptual Mix: Combine direct physical descriptors with abstract search concepts.
  Platform Compliance: Avoid spammy, repetitive, or misleading keywords.
  Formatting: Format tags as a single string per key, with individual keywords separated by commas.

JSON Schema:
{
  "filename": "<exact_filename>",
  "description_en": "Concise English description...",
  "description_fr": "Description concise en français...",
  "tags_en": "tag1, tag2, tag3...",
  "tags_fr": "balise1, balise2, balise3..."
}`)

	default: // "es" or any other language defaults to Spanish secondary
		return fmt.Sprintf(`Role: You are an expert stock photography metadata specialist and SEO copywriter optimized for Pexels, Unsplash, and Pixabay.

Task: Analyze the visual content of the uploaded image and output a valid JSON object containing high-ranking SEO descriptions and tag sets.
Primary language is English, with Spanish as secondary language.
Use the uploaded image's exact file name to map the entry.
Output strictly valid JSON with no introductory text, surrounding commentary, or conversational filler.

Metadata Rules:
- Descriptions (description_en, description_es):
  Concise, highly descriptive, natural-sounding titles/descriptions (1-2 sentences) in English and Spanish.
  Lead with the primary subject and key action or setting to capture top search rankings across stock platforms.

- SEO Tags (tags_en, tags_es):
  Generate up to 30 tags per language per image.
  Relevance & Search Intent: Target core subjects, specific actions, visual environment, lighting, composition, color palette, and mood/emotion.
  Literal & Conceptual Mix: Combine direct physical descriptors with abstract search concepts.
  Platform Compliance: Avoid spammy, repetitive, or misleading keywords.
  Formatting: Format tags as a single string per key, with individual keywords separated by commas.

JSON Schema:
{
  "filename": "<exact_filename>",
  "description_en": "Concise English description...",
  "description_es": "Descripción concisa en español...",
  "tags_en": "tag1, tag2, tag3...",
  "tags_es": "etiqueta1, etiqueta2, etiqueta3..."
}`)
	}
}

// ModelInfo represents a model returned by the Gemini models endpoint.
type ModelInfo struct {
	Name                       string   `json:"name"`
	DisplayName                string   `json:"displayName"`
	Description                string   `json:"description"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
}

// ListModelsResponse holds the models returned by the Gemini API.
type ListModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// FallbackPromptFunc is a callback invoked when a model is unavailable (503/404)
// to ask the user whether to switch to alternativeModel. Returns true if approved by user.
type FallbackPromptFunc func(currentModel, alternativeModel, reason string) bool

// Client manages Gemini API calls with throttling and safety controls.
type Client struct {
	APIKey         string
	Model          string
	HTTPClient     *http.Client
	Delay          time.Duration
	PromptFallback FallbackPromptFunc
}

// NewClient initializes a Gemini client with reasonable safety defaults.
func NewClient(apiKey, model string) *Client {
	if model == "" {
		if envModel := os.Getenv("GEMINI_MODEL"); envModel != "" {
			model = envModel
		} else {
			model = DefaultModel
		}
	}
	return &Client{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		Delay: 4 * time.Second, // Throttling: safe for free tier 15 RPM
	}
}

// ListGenerateContentModels queries the Gemini API for all available models that support generateContent.
func (c *Client) ListGenerateContentModels(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", c.APIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list models (status %d): %s", resp.StatusCode, string(body))
	}

	var listResp ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	var valid []string
	for _, m := range listResp.Models {
		canGenerate := false
		for _, method := range m.SupportedGenerationMethods {
			if method == "generateContent" {
				canGenerate = true
				break
			}
		}
		if canGenerate {
			modelID := strings.TrimPrefix(m.Name, "models/")
			valid = append(valid, modelID)
		}
	}
	return valid, nil
}

// DiscoverBestModel finds the best available model for the user's API key.
// It queries the live Gemini models API, automatically choosing the newest supported Flash model.
func (c *Client) DiscoverBestModel(ctx context.Context) string {
	if envModel := os.Getenv("GEMINI_MODEL"); envModel != "" {
		return envModel
	}

	available, err := c.ListGenerateContentModels(ctx)
	if err != nil || len(available) == 0 {
		return DefaultModel
	}

	availableMap := make(map[string]bool, len(available))
	for _, m := range available {
		availableMap[m] = true
	}

	// Priority list of optimal stock tagging multimodal models
	priority := []string{
		"gemini-3.6-flash",
		"gemini-2.5-flash",
		"gemini-2.0-flash",
		"gemini-1.5-flash",
		"gemini-1.5-flash-latest",
	}

	for _, p := range priority {
		if availableMap[p] {
			return p
		}
	}

	// Fallback 1: Any available model with "flash" in the name
	for _, m := range available {
		if strings.Contains(strings.ToLower(m), "flash") {
			return m
		}
	}

	// Fallback 2: Any available model starting with "gemini"
	for _, m := range available {
		if strings.HasPrefix(strings.ToLower(m), "gemini") {
			return m
		}
	}

	return available[0]
}

// EnsureModel verifies and resolves an active, working model from the live API.
func (c *Client) EnsureModel(ctx context.Context) string {
	best := c.DiscoverBestModel(ctx)
	if best != "" {
		c.Model = best
	}
	return c.Model
}

// NextAlternativeModel returns the next best working model, excluding models that failed or are overloaded.
func (c *Client) NextAlternativeModel(ctx context.Context, failedModel string) string {
	available, err := c.ListGenerateContentModels(ctx)
	if err != nil || len(available) == 0 {
		if failedModel != "gemini-2.5-flash" {
			return "gemini-2.5-flash"
		}
		return "gemini-1.5-flash"
	}

	priority := []string{
		"gemini-2.5-flash",
		"gemini-1.5-flash",
		"gemini-1.5-flash-latest",
		"gemini-3.6-flash",
		"gemini-2.0-flash",
	}

	for _, p := range priority {
		if p != failedModel {
			for _, a := range available {
				if a == p {
					return p
				}
			}
		}
	}

	for _, a := range available {
		if a != failedModel && strings.Contains(strings.ToLower(a), "flash") {
			return a
		}
	}

	for _, a := range available {
		if a != failedModel {
			return a
		}
	}

	return DefaultModel
}

// PrepareImage loads and optionally resizes an image to keep payloads small and safe.
func PrepareImage(imagePath string) (string, string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(imagePath))
	mimeType := "image/jpeg"
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".webp":
		mimeType = "image/webp"
	}

	// Attempt decoding to inspect size
	imgConfig, _, err := image.DecodeConfig(file)
	if err == nil && (imgConfig.Width > MaxDimension || imgConfig.Height > MaxDimension) {
		// Rewind file to decode full image
		if _, seekErr := file.Seek(0, io.SeekStart); seekErr == nil {
			img, _, decodeErr := image.Decode(file)
			if decodeErr == nil {
				resized := downscaleImage(img, MaxDimension)
				var buf bytes.Buffer
				if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85}); err == nil {
					encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
					return encoded, "image/jpeg", nil
				}
			}
		}
	}

	// Fallback to reading raw file bytes if decoding or resizing wasn't necessary or failed
	var data []byte
	if _, seekErr := file.Seek(0, io.SeekStart); seekErr == nil {
		data, err = io.ReadAll(file)
	} else {
		data, err = os.ReadFile(imagePath)
	}
	if err != nil {
		return "", "", err
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, mimeType, nil
}

// downscaleImage performs simple proportional downscaling.
func downscaleImage(src image.Image, maxDim int) image.Image {
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= maxDim && h <= maxDim {
		return src
	}

	var newW, newH int
	if w > h {
		newW = maxDim
		newH = (h * maxDim) / w
	} else {
		newH = maxDim
		newW = (w * maxDim) / h
	}

	srcXCoords := make([]int, newW)
	for x := 0; x < newW; x++ {
		srcXCoords[x] = bounds.Min.X + (x*w)/newW
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		srcY := bounds.Min.Y + (y*h)/newH
		for x := 0; x < newW; x++ {
			dst.Set(x, y, src.At(srcXCoords[x], srcY))
		}
	}
	return dst
}

// TagImage sends a single image to Gemini Vision and parses the structured response.
func (c *Client) TagImage(ctx context.Context, imagePath string, targetLang string) (*MetadataEntry, error) {
	filename := filepath.Base(imagePath)
	b64Data, mimeType, err := PrepareImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed preparing image %s: %w", filename, err)
	}

	prompt := BuildSystemPrompt(targetLang)

	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{
						Text: fmt.Sprintf("%s\n\nImage filename: %s", prompt, filename),
					},
					{
						InlineData: &InlineData{
							MimeType: mimeType,
							Data:     b64Data,
						},
					},
				},
			},
		},
		GenerationConfig: &GenerationConfig{
			Temperature:      0.2,
			ResponseMimeType: "application/json",
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", APIBaseURL, c.Model, c.APIKey)

	// Retry loop for rate limiting (429)
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*5) * time.Second)
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBytes))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := c.HTTPClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited (HTTP 429). Retrying in %ds...", (attempt+1)*5)
			continue
		}

		if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusNotFound {
			statusReason := fmt.Sprintf("HTTP %d", resp.StatusCode)
			if resp.StatusCode == http.StatusServiceUnavailable {
				statusReason = "HTTP 503: High demand / overloaded"
				lastErr = fmt.Errorf("model high demand (HTTP 503)")
			} else {
				statusReason = "HTTP 404: Model not found / deprecated"
				lastErr = fmt.Errorf("model not found (HTTP 404)")
			}

			alt := c.NextAlternativeModel(ctx, c.Model)
			if alt != "" && alt != c.Model {
				allowSwitch := false
				if c.PromptFallback != nil {
					allowSwitch = c.PromptFallback(c.Model, alt, statusReason)
				}
				if allowSwitch {
					c.Model = alt
					url = fmt.Sprintf("%s/%s:generateContent?key=%s", APIBaseURL, c.Model, c.APIKey)
					continue
				}
			}

			return nil, fmt.Errorf("gemini API error (%s): %s", statusReason, string(body))
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
		}

		var geminiResp GeminiResponse
		if err := json.Unmarshal(body, &geminiResp); err != nil {
			return nil, fmt.Errorf("failed parsing gemini API response: %w", err)
		}

		if geminiResp.Error != nil {
			return nil, fmt.Errorf("gemini returned error: %s (code %d)", geminiResp.Error.Message, geminiResp.Error.Code)
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			return nil, fmt.Errorf("empty response candidates from Gemini")
		}

		rawText := geminiResp.Candidates[0].Content.Parts[0].Text
		cleanJSON := cleanMarkdownFence(rawText)

		var entry MetadataEntry
		if err := json.Unmarshal([]byte(cleanJSON), &entry); err != nil {
			// Check if returned as array
			var arr []MetadataEntry
			if arrErr := json.Unmarshal([]byte(cleanJSON), &arr); arrErr == nil && len(arr) > 0 {
				entry = arr[0]
			} else {
				return nil, fmt.Errorf("failed to parse entry JSON: %w (raw: %s)", err, cleanJSON)
			}
		}

		if entry.Filename == "" {
			entry.Filename = filename
		}

		return &entry, nil
	}

	return nil, fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

// cleanMarkdownFence strips ```json ... ``` code blocks if present.
func cleanMarkdownFence(s string) string {
	t := strings.TrimSpace(s)
	if strings.HasPrefix(t, "```") {
		lines := strings.Split(t, "\n")
		if len(lines) >= 2 {
			if strings.HasPrefix(lines[0], "```") {
				lines = lines[1:]
			}
			if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "```") {
				lines = lines[:len(lines)-1]
			}
			t = strings.Join(lines, "\n")
		}
	}
	return strings.TrimSpace(t)
}

// SaveBackupJSON writes the generated entries to a timestamped file for safety.
func SaveBackupJSON(entries []MetadataEntry, outputDir string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("stock_tags_%s.json", timestamp)
	targetPath := filepath.Join(outputDir, filename)

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return "", err
	}

	return targetPath, nil
}
