package interactive

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lagosproject/StockTaggerIA/src/gemini"
	"github.com/lagosproject/StockTaggerIA/src/i18n"
	"github.com/lagosproject/StockTaggerIA/src/metadata"
	"github.com/lagosproject/StockTaggerIA/src/presets"
)

// RunWizard launches the step-by-step interactive configuration wizard.
func RunWizard(defaultFolder, defaultJSON string) error {
	reader := bufio.NewReader(os.Stdin)

	// Step 0: Language Selection
	detectedLang := i18n.DetectSystemLanguage()
	fmt.Println("\n=======================================================")
	fmt.Println("  StockTaggerIA - Stock Photography SEO Metadata")
	fmt.Println("=======================================================")
	fmt.Println("Select Interface Language / Seleccionar idioma / Choisir la langue:")
	fmt.Printf("  1) English %s\n", defaultIndicator(detectedLang == i18n.LangEN))
	fmt.Printf("  2) Español %s\n", defaultIndicator(detectedLang == i18n.LangES))
	fmt.Printf("  3) Français %s\n", defaultIndicator(detectedLang == i18n.LangFR))
	fmt.Print("Choice [1-3, Enter for default]: ")

	activeLang := detectedLang
	langInput, _ := reader.ReadString('\n')
	langInput = strings.TrimSpace(langInput)
	switch langInput {
	case "1", "en", "EN":
		activeLang = i18n.LangEN
	case "2", "es", "ES":
		activeLang = i18n.LangES
	case "3", "fr", "FR":
		activeLang = i18n.LangFR
	}

	msg := i18n.Get(activeLang)
	fmt.Printf("-> Active language: %s\n\n", msg.LanguageName)

	// Step 1: Metadata Source Selection
	fmt.Println(msg.SelectSource)
	fmt.Printf("  %s\n", msg.SourceGemini)
	fmt.Printf("  %s\n", msg.SourceJSON)
	fmt.Print("Choice [1-2, default 1]: ")
	sourceInput, _ := reader.ReadString('\n')
	sourceInput = strings.TrimSpace(sourceInput)
	isGeminiMode := sourceInput != "2"

	// Step 2: Gemini API Key acquisition (if Gemini mode)
	var apiKey string
	if isGeminiMode {
		if envKey := os.Getenv("GEMINI_API_KEY"); envKey != "" {
			apiKey = envKey
			fmt.Println(msg.APIKeyFoundEnv)
		} else {
			secret, err := readPasswordSafe(reader, msg.APIKeyPrompt)
			if err != nil || secret == "" {
				return fmt.Errorf("%s", msg.APIKeyMissing)
			}
			apiKey = secret
		}
	}

	// Step 3: Target Folder or Dragged Images Selection
	fmt.Println("\n" + msg.EnterFolder)
	if defaultFolder != "" {
		fmt.Printf("Default: [%s]\n", defaultFolder)
	}
	var candidateImages []string
	var primaryFolder string

	for {
		fmt.Print("Path or drop files: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" && defaultFolder != "" {
			input = defaultFolder
		}

		paths := metadata.ParseDroppedPaths(input)
		candidateImages = metadata.ResolveCandidateImages(paths)
		if len(candidateImages) > 0 {
			if len(paths) == 1 {
				if info, err := os.Stat(paths[0]); err == nil && info.IsDir() {
					primaryFolder = paths[0]
				} else {
					primaryFolder = filepath.Dir(candidateImages[0])
				}
			} else {
				primaryFolder = filepath.Dir(candidateImages[0])
			}
			break
		}
		fmt.Println(msg.FolderNotFound)
	}

	fmt.Printf(msg.FoundImages+"\n", len(candidateImages))

	// Step 4: JSON Path Selection (if JSON mode)
	var jsonFilePath string
	if !isGeminiMode {
		fmt.Printf("\nEnter JSON metadata file path [%s]: ", defaultJSON)
		jInput, _ := reader.ReadString('\n')
		jInput = strings.TrimSpace(jInput)
		if jInput == "" {
			jInput = defaultJSON
		}
		jsonFilePath = jInput
	}

	// Step 5: Platform Preset Selection
	fmt.Println("\n" + msg.SelectPreset)
	fmt.Printf("  %s\n", msg.PresetStock)
	fmt.Printf("  %s\n", msg.PresetPixabay)
	fmt.Print("Choice [1-2, default 1]: ")
	presetInput, _ := reader.ReadString('\n')
	presetInput = strings.TrimSpace(presetInput)

	chosenPreset := presets.AvailablePresets["stock"]
	if presetInput == "2" {
		chosenPreset = presets.AvailablePresets["pixabay"]
	}

	// Step 6: Metadata Language Selection (English is primary stock base, secondary languages optional)
	fmt.Println("\n" + msg.SelectMetaLang)
	fmt.Printf("  %s %s\n", msg.MetaLangEN, defaultIndicator(true))
	fmt.Printf("  %s\n", msg.MetaLangENES)
	fmt.Printf("  %s\n", msg.MetaLangENFR)
	fmt.Printf("  %s\n", msg.MetaLangAll)
	defaultMetaChoice := "1"
	fmt.Print("Choice [1-4, default 1]: ")
	metaLangInput, _ := reader.ReadString('\n')
	metaLangInput = strings.TrimSpace(metaLangInput)
	if metaLangInput == "" {
		metaLangInput = defaultMetaChoice
	}

	metaLang := "en"
	langSummary := "EN (English only)"
	switch metaLangInput {
	case "1", "en", "EN":
		metaLang = "en"
		langSummary = "EN (English only)"
	case "2", "es", "ES":
		metaLang = "es"
		langSummary = "EN + ES (English + Spanish secondary)"
	case "3", "fr", "FR":
		metaLang = "fr"
		langSummary = "EN + FR (English + French secondary)"
	case "4", "all", "ALL":
		metaLang = "all"
		langSummary = "EN + ES + FR (Trilingual)"
	}

	// Step 7: Dry-Run Mode
	fmt.Println("\n" + msg.DryRunPrompt)
	fmt.Printf("  %s\n", msg.DryRunYes)
	fmt.Printf("  %s\n", msg.DryRunNo)
	fmt.Print("Choice [1-2, default 2]: ")
	modeInput, _ := reader.ReadString('\n')
	modeInput = strings.TrimSpace(modeInput)
	dryRun := modeInput == "1"

	// Step 8: Pre-flight Summary & Confirmation
	modeStr := "LIVE UPDATE (Modifying image EXIF/IPTC)"
	if dryRun {
		modeStr = "DRY RUN (Simulation only)"
	}
	sourceStr := "Gemini Vision AI (Multimodal)"
	if !isGeminiMode {
		sourceStr = fmt.Sprintf("JSON File: %s", jsonFilePath)
	}

	fmt.Println("\n" + msg.PreFlightTitle)
	fmt.Printf(msg.PreFlightFolder+"\n", primaryFolder)
	fmt.Printf(msg.PreFlightImages+"\n", len(candidateImages))
	fmt.Printf(msg.PreFlightSource+"\n", sourceStr)
	fmt.Printf(msg.PreFlightPreset+"\n", chosenPreset.Name)
	fmt.Printf(msg.PreFlightMetaLang+"\n", langSummary)
	fmt.Printf(msg.PreFlightMode+"\n", modeStr)
	fmt.Println("================================")
	fmt.Print(msg.ConfirmPrompt)

	confirmInput, _ := reader.ReadString('\n')
	confirmInput = strings.ToLower(strings.TrimSpace(confirmInput))
	if confirmInput != "" && confirmInput != "y" && confirmInput != "s" && confirmInput != "o" && confirmInput != "yes" && confirmInput != "si" && confirmInput != "oui" {
		fmt.Println(msg.Cancelled)
		return nil
	}

	// Step 9: Execution
	var metadataMap map[string]metadata.MetadataRaw
	if isGeminiMode {
		if len(candidateImages) == 0 {
			fmt.Println("No images found to analyze.")
			return nil
		}
		ctx := context.Background()
		client := gemini.NewClient(apiKey, "")
		resolvedModel := client.EnsureModel(ctx)
		client.PromptFallback = func(currentModel, altModel, reason string) bool {
			fmt.Printf(msg.ModelUnavailablePrompt, currentModel, reason, altModel)
			ans, _ := reader.ReadString('\n')
			ans = strings.ToLower(strings.TrimSpace(ans))
			return ans == "" || ans == "s" || ans == "y" || ans == "yes" || ans == "si" || ans == "o" || ans == "oui"
		}
		fmt.Printf("\n%s (%s)\n", msg.StartingTagging, resolvedModel)

		var generatedEntries []gemini.MetadataEntry
		var rawEntries []metadata.MetadataRaw

		for i, imgPath := range candidateImages {
			relName := filepath.Base(imgPath)
			fmt.Printf(msg.ProgressTagging+"\n", i+1, len(candidateImages), relName)
			entry, err := client.TagImage(ctx, imgPath, metaLang)
			if err != nil {
				fmt.Printf("[WARN] Failed analyzing %s: %v\n", relName, err)
				continue
			}
			generatedEntries = append(generatedEntries, *entry)
			rawEntries = append(rawEntries, entry.ToMetadataRaw())
		}

		if len(generatedEntries) > 0 {
			backupFile, err := gemini.SaveBackupJSON(generatedEntries, primaryFolder)
			if err == nil {
				fmt.Printf("\n"+msg.BackupSaved+"\n", backupFile)
			}
		}

		metadataMap = metadata.MetadataMapFromRawEntries(rawEntries)
	} else {
		mMap, err := metadata.LoadMetadataMapFromJSON(jsonFilePath)
		if err != nil {
			return fmt.Errorf(msg.ErrorLoadJSON, jsonFilePath, err)
		}
		metadataMap = mMap
	}

	fmt.Println("\n" + msg.StartingExifTool)
	metadata.UpdateImagesMetadata(metadataMap, candidateImages, primaryFolder, chosenPreset, metaLang, dryRun, msg)
	return nil
}

func defaultIndicator(isDefault bool) string {
	if isDefault {
		return "(default)"
	}
	return ""
}

func readPasswordSafe(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	if runtime.GOOS != "windows" {
		cmd := exec.Command("stty", "-echo")
		cmd.Stdin = os.Stdin
		_ = cmd.Run()
		defer func() {
			cmdRestore := exec.Command("stty", "echo")
			cmdRestore.Stdin = os.Stdin
			_ = cmdRestore.Run()
			fmt.Println()
		}()
	}

	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(text)
	if runtime.GOOS == "windows" {
		fmt.Println()
	}
	return trimmed, nil
}
