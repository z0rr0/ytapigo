package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/z0rr0/ytapigo/cloud"
	"github.com/z0rr0/ytapigo/config"
)

// LanguagesURL is a URL for API get languages request.
// Documentation https://cloud.yandex.com/en/docs/translate/api-ref/Translation/listLanguages
const LanguagesURL = "https://translate.api.cloud.yandex.net/translate/v2/languages"

// Language is language info item.
type Language struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Languages is a list of translation's languages (from JSON response).
// Example `{"languages":[{"code":"af","name":"Afrikaans"}]}`.
type Languages struct {
	Languages []Language `json:"languages"`
	Codes     map[string]struct{}
}

func (l *Language) String() string {
	return fmt.Sprintf("%v - %v", l.Code, l.Name)
}

// String is an implementation of String() method for Languages pointer.
func (languages *Languages) String() string {
	codes := make([]string, languages.Len())
	for i, v := range languages.Languages {
		codes[i] = v.Code
	}
	return fmt.Sprintf("Translation languages:\n%v\n%v", strings.Join(codes, ", "), languages.Description())
}

// Len return length of languages list.
func (languages *Languages) Len() int {
	return len(languages.Languages)
}

// Contains is an implementation of Contains() method for Languages
// pointer (LangChecker interface).
func (languages *Languages) Contains(fromLanguage, toLanguage string) bool {
	_, langFrom := languages.Codes[strings.ToLower(fromLanguage)]
	_, langTo := languages.Codes[strings.ToLower(toLanguage)]

	return langFrom && langTo
}

// Description is an implementation of Description() method.
func (languages *Languages) Description() string {
	const n int = 3
	var colLen int

	counter := languages.Len()
	if (counter % n) != 0 {
		colLen = counter/n + 1
	} else {
		colLen = counter / n
	}
	output := make([]string, colLen)
	for j := 0; j < colLen; j++ {
		switch {
		case j+2*colLen < counter:
			output[j] = fmt.Sprintf(
				"%-25v %-25v %-25v",
				languages.Languages[j].String(), languages.Languages[j+colLen].String(), languages.Languages[j+2*colLen].String(),
			)
		case j+colLen < counter:
			output[j] = fmt.Sprintf("%-25v %-25v", languages.Languages[j].String(), languages.Languages[j+colLen].String())
		default:
			output[j] = fmt.Sprintf("%-25v", languages.Languages[j].String())
		}
	}
	return strings.Join(output, "\n")
}

// LoadLanguages loads available dictionary languages.
func LoadLanguages(ctx context.Context, client *http.Client, cfg *config.Config) (*Languages, error) {
	err := cfg.InitToken(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to init IAM token: %w", err)
	}

	data := strings.NewReader(fmt.Sprintf(`{"folder_id":"%s"}`, cfg.Translation.FolderID))

	body, err := cloud.Request(ctx, client, data, cfg.GetURL(LanguagesURL), cfg.Translation.IAMToken, cfg.UserAgent, true, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get translation languages: %w", err)
	}

	languages := Languages{}
	if err = json.Unmarshal(body, &languages); err != nil {
		return nil, fmt.Errorf("failed to decode translation languages response: %w", err)
	}

	languages.Codes = make(map[string]struct{}, languages.Len())
	for _, v := range languages.Languages {
		languages.Codes[strings.ToLower(v.Code)] = struct{}{}
	}

	return &languages, nil
}
