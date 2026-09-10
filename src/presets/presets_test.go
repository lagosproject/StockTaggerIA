package presets

import (
	"testing"
)

func TestResolvePreset(t *testing.T) {
	// Test aliases
	p1 := ResolvePreset("unsplash")
	if p1.ID != "stock" || p1.MaxTags != 30 {
		t.Errorf("Expected stock preset with 30 max tags for 'unsplash', got %v", p1)
	}

	p2 := ResolvePreset("pexels")
	if p2.ID != "stock" || p2.MaxTags != 30 {
		t.Errorf("Expected stock preset with 30 max tags for 'pexels', got %v", p2)
	}

	p3 := ResolvePreset("pixabay")
	if p3.ID != "pixabay" || p3.MaxTags != 25 {
		t.Errorf("Expected pixabay preset with 25 max tags, got %v", p3)
	}

	// Unknown preset fallback
	pDefault := ResolvePreset("unknown_platform")
	if pDefault.ID != "stock" {
		t.Errorf("Expected fallback to stock preset, got %v", pDefault)
	}
}

func TestResolveLanguage(t *testing.T) {
	stock := AvailablePresets["stock"]
	pixabay := AvailablePresets["pixabay"]

	// Default fallback
	if l := ResolveLanguage(stock, ""); l != "en" {
		t.Errorf("Expected default 'en' for stock, got %s", l)
	}
	if l := ResolveLanguage(pixabay, ""); l != "es" {
		t.Errorf("Expected default 'es' for pixabay, got %s", l)
	}

	// Overrides
	if l := ResolveLanguage(stock, "es"); l != "es" {
		t.Errorf("Expected override to 'es', got %s", l)
	}
	if l := ResolveLanguage(pixabay, "fr"); l != "fr" {
		t.Errorf("Expected override to 'fr', got %s", l)
	}
}
