package i18n

import (
	"os"
	"strings"
)

// Language represents supported languages for CLI UI and metadata.
type Language string

const (
	LangEN Language = "en"
	LangES Language = "es"
	LangFR Language = "fr"
)

// Messages contains all translatable strings for the interactive CLI.
type Messages struct {
	LanguageName       string
	BannerSubtitle     string
	SelectInterfaceLang string
	SelectSource       string
	SourceGemini       string
	SourceJSON         string
	EnterFolder        string
	FolderNotFound     string
	FoundImages        string
	SelectPreset       string
	PresetStock        string
	PresetPixabay      string
	SelectMetaLang     string
	LangEnglish        string
	LangSpanish        string
	LangFrench         string
	MetaLangEN         string
	MetaLangENES       string
	MetaLangENFR       string
	MetaLangAll        string
	DryRunPrompt       string
	DryRunYes          string
	DryRunNo           string
	APIKeyPrompt       string
	APIKeyFoundEnv     string
	APIKeyMissing      string
	ModelUnavailablePrompt string
	PreFlightTitle     string
	PreFlightFolder    string
	PreFlightImages    string
	PreFlightSource    string
	PreFlightPreset    string
	PreFlightMetaLang  string
	PreFlightMode      string
	ConfirmPrompt      string
	Cancelled          string
	StartingTagging    string
	ProgressTagging    string
	BackupSaved        string
	StartingExifTool   string
	SummaryTitle       string
	SummaryPreset      string
	SummaryMetaLang    string
	SummaryScanned     string
	SummaryMatches     string
	SummaryUpdated     string
	SummaryProjected   string
	SummaryErrors      string
	ExifToolNotFound   string
	ErrorWalkDir       string
	ErrorLoadJSON      string
	ErrorParseJSON     string
}

var translations = map[Language]Messages{
	LangEN: {
		LanguageName:        "English",
		BannerSubtitle:      "Stock Photography SEO Tagger (Pexels, Unsplash, Pixabay)",
		SelectInterfaceLang: "Select Interface Language / Seleccionar idioma / Choisir la langue:",
		SelectSource:        "Choose metadata source:",
		SourceGemini:        "1) Gemini Vision AI (Analyze images directly)",
		SourceJSON:          "2) Existing JSON file (Offline / pre-generated)",
		EnterFolder:         "Enter folder or drag & drop images/folders here:",
		FolderNotFound:      "No valid images or directory found. Please try again (drag & drop supported):",
		FoundImages:         "Found %d candidate images (JPG, JPEG, PNG, WEBP).",
		SelectPreset:        "Select stock platform preset:",
		PresetStock:         "1) Unsplash & Pexels (Max 30 tags, high search intent)",
		PresetPixabay:       "2) Pixabay (Max 25 tags)",
		SelectMetaLang:      "Select metadata language(s) (English is primary stock base):",
		LangEnglish:         "1) English",
		LangSpanish:         "2) Spanish (Español)",
		LangFrench:          "3) French (Français)",
		MetaLangEN:          "1) English only (Recommended: fastest & stock standard)",
		MetaLangENES:        "2) English + Spanish (Español as secondary language)",
		MetaLangENFR:        "3) English + French (Français as secondary language)",
		MetaLangAll:         "4) English + Spanish + French (Trilingual)",
		DryRunPrompt:        "Execution mode:",
		DryRunYes:           "1) Dry Run (Simulate only, no files modified)",
		DryRunNo:            "2) Live Update (Write EXIF/IPTC/XMP to images)",
		APIKeyPrompt:        "Enter Gemini API Key (input hidden): ",
		APIKeyFoundEnv:      "Using Gemini API key from GEMINI_API_KEY environment variable.",
		APIKeyMissing:       "Gemini API key is required to proceed.",
		ModelUnavailablePrompt: "\n[WARNING] Model '%s' is unavailable (%s).\nDo you want to switch to alternative model '%s'? [Y/n]: ",
		PreFlightTitle:      "=== PRE-FLIGHT VERIFICATION ===",
		PreFlightFolder:     "Target Folder  : %s",
		PreFlightImages:     "Images to tag  : %d",
		PreFlightSource:     "Source         : %s",
		PreFlightPreset:     "Platform Preset: %s",
		PreFlightMetaLang:   "Metadata Lang  : %s",
		PreFlightMode:       "Mode           : %s",
		ConfirmPrompt:       "Do you want to proceed? [Y/n]: ",
		Cancelled:           "Operation cancelled by user.",
		StartingTagging:     "--> Analyzing images with Gemini Vision AI...",
		ProgressTagging:     "[%d/%d] Analyzing %s...",
		BackupSaved:         "AI metadata backup safely saved to: %s",
		StartingExifTool:    "--> Writing metadata into images via ExifTool...",
		SummaryTitle:        "EXECUTION SUMMARY %s",
		SummaryPreset:       "Platform Preset : %s",
		SummaryMetaLang:     "Metadata Lang   : %s",
		SummaryScanned:      "Files Scanned   : %d",
		SummaryMatches:      "Matches Found   : %d",
		SummaryUpdated:      "Updated Files   : %d",
		SummaryProjected:    "Projected Files : %d",
		SummaryErrors:       "Errors          : %d",
		ExifToolNotFound:    "exiftool executable not found. Ensure exiftool is in PATH or placed in the application folder.",
		ErrorWalkDir:        "Error walking directory '%s': %v",
		ErrorLoadJSON:       "Failed to load JSON file '%s': %v",
		ErrorParseJSON:      "Failed to parse JSON file '%s': %v",
	},
	LangES: {
		LanguageName:        "Español",
		BannerSubtitle:      "Etiquetador SEO para Fotografía de Stock (Pexels, Unsplash, Pixabay)",
		SelectInterfaceLang: "Seleccione el idioma de la interfaz:",
		SelectSource:        "Elija el origen de los metadatos:",
		SourceGemini:        "1) Gemini Vision AI (Analizar imágenes directamente)",
		SourceJSON:          "2) Archivo JSON existente (Modo sin conexión / pregenerado)",
		EnterFolder:         "Introduzca carpeta o arrastre imágenes/carpetas aquí:",
		FolderNotFound:      "No se encontraron imágenes válidas ni el directorio. Inténtelo de nuevo (puede arrastrar archivos):",
		FoundImages:         "Se encontraron %d imágenes candidatas (JPG, JPEG, PNG, WEBP).",
		SelectPreset:        "Seleccione el perfil de plataforma de stock:",
		PresetStock:         "1) Unsplash y Pexels (Máx. 30 etiquetas, intención de búsqueda alta)",
		PresetPixabay:       "2) Pixabay (Máx. 25 etiquetas)",
		SelectMetaLang:      "Seleccione los idiomas de metadatos (el inglés se incluye como base para stock):",
		LangEnglish:         "1) Inglés",
		LangSpanish:         "2) Español",
		LangFrench:          "3) Francés",
		MetaLangEN:          "1) Solo Inglés (Recomendado: más rápido y estándar de stock)",
		MetaLangENES:        "2) Inglés + Español (Añadir Español como idioma secundario)",
		MetaLangENFR:        "3) Inglés + Francés (Añadir Francés como idioma secundario)",
		MetaLangAll:         "4) Inglés + Español + Francés (Trilingüe)",
		DryRunPrompt:        "Modo de ejecución:",
		DryRunYes:           "1) Simulación (Dry Run: no modifica ningún archivo)",
		DryRunNo:            "2) Actualización en vivo (Escribir EXIF/IPTC/XMP en archivos)",
		APIKeyPrompt:        "Introduzca la clave API de Gemini (entrada oculta): ",
		APIKeyFoundEnv:      "Usando la clave API de Gemini de la variable de entorno GEMINI_API_KEY.",
		APIKeyMissing:       "Se requiere una clave API de Gemini para continuar.",
		ModelUnavailablePrompt: "\n[AVISO] El modelo '%s' no está disponible (%s).\n¿Desea cambiar al modelo alternativo '%s'? [S/n]: ",
		PreFlightTitle:      "=== VERIFICACIÓN PREVIA ===",
		PreFlightFolder:     "Carpeta destino : %s",
		PreFlightImages:     "Fotos a procesar: %d",
		PreFlightSource:     "Origen          : %s",
		PreFlightPreset:     "Perfil de Stock : %s",
		PreFlightMetaLang:   "Idioma Metadatos: %s",
		PreFlightMode:       "Modo            : %s",
		ConfirmPrompt:       "¿Desea continuar? [S/n]: ",
		Cancelled:           "Operación cancelada por el usuario.",
		StartingTagging:     "--> Analizando imágenes con Gemini Vision AI...",
		ProgressTagging:     "[%d/%d] Analizando %s...",
		BackupSaved:         "Copia de seguridad de metadatos guardada en: %s",
		StartingExifTool:    "--> Escribiendo metadatos en las imágenes con ExifTool...",
		SummaryTitle:        "RESUMEN DE EJECUCIÓN %s",
		SummaryPreset:       "Perfil de Stock : %s",
		SummaryMetaLang:     "Idioma Metadatos: %s",
		SummaryScanned:      "Archivos analizados : %d",
		SummaryMatches:      "Coincidencias       : %d",
		SummaryUpdated:      "Archivos actualizados: %d",
		SummaryProjected:    "Archivos proyectados : %d",
		SummaryErrors:       "Errores              : %d",
		ExifToolNotFound:    "No se encontró el ejecutable exiftool. Asegúrese de que esté en el PATH o junto al programa.",
		ErrorWalkDir:        "Error al recorrer el directorio '%s': %v",
		ErrorLoadJSON:       "Error al cargar el archivo JSON '%s': %v",
		ErrorParseJSON:      "Error al interpretar el archivo JSON '%s': %v",
	},
	LangFR: {
		LanguageName:        "Français",
		BannerSubtitle:      "Baliseur SEO pour Photographie de Stock (Pexels, Unsplash, Pixabay)",
		SelectInterfaceLang: "Sélectionnez la langue de l'interface :",
		SelectSource:        "Choisissez la source des métadonnées :",
		SourceGemini:        "1) Gemini Vision AI (Analyser les images directement)",
		SourceJSON:          "2) Fichier JSON existant (Hors-ligne / pré-généré)",
		EnterFolder:         "Entrez un dossier ou glissez-déposez des images/dossiers ici :",
		FolderNotFound:      "Aucune image valide ou dossier trouvé. Veuillez réessayer (glisser-déposer supporté) :",
		FoundImages:         "Trouvé %d images compatibles (JPG, JPEG, PNG, WEBP).",
		SelectPreset:        "Sélectionnez le profil de plateforme stock :",
		PresetStock:         "1) Unsplash & Pexels (Max 30 mots-clés, forte intention de recherche)",
		PresetPixabay:       "2) Pixabay (Max 25 mots-clés)",
		SelectMetaLang:      "Sélectionnez les langues des métadonnées (l'anglais est inclus comme base stock) :",
		LangEnglish:         "1) Anglais",
		LangSpanish:         "2) Espagnol",
		LangFrench:          "3) Français",
		MetaLangEN:          "1) Anglais uniquement (Recommandé : rapide & standard de stock)",
		MetaLangENES:        "2) Anglais + Espagnol (Ajouter l'espagnol comme langue secondaire)",
		MetaLangENFR:        "3) Anglais + Français (Ajouter le français comme langue secondaire)",
		MetaLangAll:         "4) Anglais + Espagnol + Français (Trilingue)",
		DryRunPrompt:        "Mode d'exécution :",
		DryRunYes:           "1) Simulation (Dry Run : aucun fichier modifié)",
		DryRunNo:            "2) Mise à jour réelle (Écrire EXIF/IPTC/XMP sur les fichiers)",
		APIKeyPrompt:        "Entrez la clé API Gemini (saisie masquée) : ",
		APIKeyFoundEnv:      "Utilisation de la clé API Gemini depuis la variable GEMINI_API_KEY.",
		APIKeyMissing:       "La clé API Gemini est requise pour continuer.",
		ModelUnavailablePrompt: "\n[ATTENTION] Le modèle '%s' est indisponible (%s).\nSouhaitez-vous basculer vers le modèle alternatif '%s' ? [O/n] : ",
		PreFlightTitle:      "=== VÉRIFICATION PRÉALABLE ===",
		PreFlightFolder:     "Dossier cible   : %s",
		PreFlightImages:     "Images à traiter: %d",
		PreFlightSource:     "Source          : %s",
		PreFlightPreset:     "Profil Stock    : %s",
		PreFlightMetaLang:   "Langue Métadonnées: %s",
		PreFlightMode:       "Mode            : %s",
		ConfirmPrompt:       "Souhaitez-vous continuer ? [O/n] : ",
		Cancelled:           "Opération annulée par l'utilisateur.",
		StartingTagging:     "--> Analyse des images avec Gemini Vision AI...",
		ProgressTagging:     "[%d/%d] Analyse de %s...",
		BackupSaved:         "Sauvegarde des métadonnées IA enregistrée sous : %s",
		StartingExifTool:    "--> Écriture des métadonnées sur les images via ExifTool...",
		SummaryTitle:        "RÉSUMÉ D'EXÉCUTION %s",
		SummaryPreset:       "Profil Stock    : %s",
		SummaryMetaLang:     "Langue Métadonnées: %s",
		SummaryScanned:      "Fichiers scannés   : %d",
		SummaryMatches:      "Correspondances    : %d",
		SummaryUpdated:      "Fichiers mis à jour: %d",
		SummaryProjected:    "Fichiers projetés  : %d",
		SummaryErrors:       "Erreurs            : %d",
		ExifToolNotFound:    "Exécutable exiftool introuvable. Assurez-vous qu'il est dans le PATH ou dans le dossier de l'application.",
		ErrorWalkDir:        "Erreur lors du parcours du dossier '%s': %v",
		ErrorLoadJSON:       "Échec du chargement du fichier JSON '%s': %v",
		ErrorParseJSON:      "Échec de l'analyse du fichier JSON '%s': %v",
	},
}

// Get returns messages for the given language, falling back to English.
func Get(lang Language) Messages {
	if m, ok := translations[lang]; ok {
		return m
	}
	return translations[LangEN]
}

// DetectSystemLanguage inspects environment variables to guess user language.
func DetectSystemLanguage() Language {
	for _, env := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		val := strings.ToLower(os.Getenv(env))
		if strings.HasPrefix(val, "es") {
			return LangES
		}
		if strings.HasPrefix(val, "fr") {
			return LangFR
		}
		if strings.HasPrefix(val, "en") {
			return LangEN
		}
	}
	return LangEN
}
