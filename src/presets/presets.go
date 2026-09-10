package presets

import (
	"strings"
)

// Preset represents a stock photography platform metadata profile.
type Preset struct {
	ID          string
	Name        string
	DefaultLang string
	MaxTags     int // 0 means unlimited
	Summary     string
}

// AvailablePresets contains all registered platform presets.
// Unsplash and Pexels are merged under "stock" (max 30 tags).
var AvailablePresets = map[string]Preset{
	"stock": {
		ID:          "stock",
		Name:        "Unsplash & Pexels",
		DefaultLang: "en",
		MaxTags:     30,
		Summary:     "Optimized for Unsplash and Pexels (Up to 30 high-relevance tags)",
	},
	"pixabay": {
		ID:          "pixabay",
		Name:        "Pixabay",
		DefaultLang: "es",
		MaxTags:     25,
		Summary:     "Optimized for Pixabay (Up to 25 tags)",
	},
}

// ResolvePreset finds a preset by name or alias, falling back to "stock".
func ResolvePreset(name string) Preset {
	norm := strings.ToLower(strings.TrimSpace(name))
	switch norm {
	case "stock", "unsplash", "pexels", "unsplash-pexels", "default":
		return AvailablePresets["stock"]
	case "pixabay":
		return AvailablePresets["pixabay"]
	default:
		if p, ok := AvailablePresets[norm]; ok {
			return p
		}
		return AvailablePresets["stock"]
	}
}

// ResolveLanguage returns the target language, respecting explicit overrides.
func ResolveLanguage(preset Preset, userLang string) string {
	l := strings.ToLower(strings.TrimSpace(userLang))
	if l == "en" || l == "es" || l == "fr" {
		return l
	}
	if preset.DefaultLang != "" {
		return preset.DefaultLang
	}
	return "en"
}
