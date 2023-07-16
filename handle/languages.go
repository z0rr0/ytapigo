package handle

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/z0rr0/ytapigo/config"
	"github.com/z0rr0/ytapigo/translation"
)

// Prefetched known languages.
const (
	// En is a constant for English language.
	En = "en"
	// Ru is a constant for Russian language.
	Ru = "ru"
	// Uk is a constant for Ukrainian language.
	Uk = "uk"

	// AutoLanguageDetect is a constant for auto-detect language using API request.
	AutoLanguageDetect = "auto"

	// in common case, the max length of direction is 7 symbols
	// but most used cases have format with 5 ones: "en-ru" and "ru-en"
	minDirectionLength = 5
	maxDirectionLength = 7
)

var autoAllowedLanguages = map[string]struct{}{En: {}, Ru: {}, Uk: {}}

// knowLanguages returns true if all languages are known and additional check requests are not required.
func knowLanguages(lang ...string) bool {
	if len(lang) == 0 {
		return false
	}

	for _, l := range lang {
		if _, ok := autoAllowedLanguages[l]; !ok {
			return false
		}
	}

	return true
}

// asciiDetection detects direction of translation for en/ru languages by a larger number of character matches.
func asciiDetection(text string) (string, string) {
	asciiCount, noASCIICount := 0, 0

	for _, runeValue := range text {
		asciiUpper := runeValue >= 65 && runeValue <= 90  // A-Z
		asciiLower := runeValue >= 97 && runeValue <= 122 // a-z

		if asciiUpper || asciiLower {
			asciiCount++
		} else {
			if runeValue > 127 {
				// exclude special ascii characters like space, comma, etc.
				noASCIICount++
			}
		}
	}

	if asciiCount > noASCIICount {
		return En, Ru
	}

	return Ru, En
}

// autoAPIDetection does API detection request.
// If successful, returns detected language and Russian as a target language.
func autoAPIDetection(ctx context.Context, client *http.Client, cfg *config.Config, text string) (string, string, error) {
	fromLanguage, err := translation.DetectLanguage(ctx, client, cfg, text)
	if err != nil {
		return "", "", fmt.Errorf("auto detect language error: %w", err)
	}

	return fromLanguage, Ru, nil
}

// splitDetection splits direction string to two languages: source and target.
func splitDetection(direction string) (string, string, error) {
	var n = len(direction)

	if n > maxDirectionLength {
		return "", "", fmt.Errorf("too long direction format='%s'", direction)
	}

	if n < minDirectionLength {
		return "", "", fmt.Errorf("too short direction format='%s'", direction)
	}

	languages := strings.SplitN(direction, "-", 3)

	if len(languages) != 2 {
		return "", "", fmt.Errorf("invalid direction format: %s", direction)
	}

	return languages[0], languages[1], nil
}

// detectLanguages tries to detect languages for translation and spelling check.
func (y *Handler) detectLanguages(ctx context.Context, direction, text string) (string, string, error) {
	if direction == AutoLanguageDetect {
		return autoAPIDetection(ctx, y.client, y.config, text)
	}

	if direction == "" {
		fromLanguage, toLanguage := asciiDetection(text)
		return fromLanguage, toLanguage, nil
	}

	return splitDetection(direction)
}
