package i18n

import (
	"testing"
)

func TestI18nCompleteness(t *testing.T) {
	langs := []Language{LangEN, LangES, LangFR}

	for _, l := range langs {
		msg := Get(l)
		if msg.LanguageName == "" {
			t.Errorf("Missing LanguageName for %s", l)
		}
		if msg.SelectPreset == "" {
			t.Errorf("Missing SelectPreset for %s", l)
		}
		if msg.ConfirmPrompt == "" {
			t.Errorf("Missing ConfirmPrompt for %s", l)
		}
		if msg.SummaryTitle == "" {
			t.Errorf("Missing SummaryTitle for %s", l)
		}
	}
}
