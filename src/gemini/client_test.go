package gemini

import (
	"strings"
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	// English prompt test
	promptEN := BuildSystemPrompt("en")
	if !strings.Contains(promptEN, "tags_en") || !strings.Contains(promptEN, "description_en") {
		t.Errorf("Expected English prompt to contain tags_en and description_en")
	}

	// Spanish prompt test
	promptES := BuildSystemPrompt("es")
	if !strings.Contains(promptES, "tags_es") || !strings.Contains(promptES, "description_es") {
		t.Errorf("Expected Spanish prompt to contain tags_es and description_es")
	}

	// French prompt test
	promptFR := BuildSystemPrompt("fr")
	if !strings.Contains(promptFR, "tags_fr") || !strings.Contains(promptFR, "description_fr") {
		t.Errorf("Expected French prompt to contain tags_fr and description_fr")
	}

	// Trilingual prompt test
	promptAll := BuildSystemPrompt("all")
	if !strings.Contains(promptAll, "tags_fr") || !strings.Contains(promptAll, "tags_es") || !strings.Contains(promptAll, "tags_en") {
		t.Errorf("Expected all-inclusive prompt to contain en, es, and fr tags")
	}
}

func TestCleanMarkdownFence(t *testing.T) {
	rawWithFence := "```json\n{\"filename\": \"test.jpg\"}\n```"
	expected := "{\"filename\": \"test.jpg\"}"
	cleaned := cleanMarkdownFence(rawWithFence)
	if cleaned != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cleaned)
	}

	rawWithoutFence := "{\"filename\": \"test.jpg\"}"
	cleanedNoFence := cleanMarkdownFence(rawWithoutFence)
	if cleanedNoFence != rawWithoutFence {
		t.Errorf("Expected '%s', got '%s'", rawWithoutFence, cleanedNoFence)
	}
}
