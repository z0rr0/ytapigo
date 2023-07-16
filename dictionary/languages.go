package dictionary

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/z0rr0/ytapigo/cloud"
	"github.com/z0rr0/ytapigo/config"
)

// LanguagesURL is a URL to load dictionary languages.
// Documentation https://yandex.com/dev/dictionary/doc/dg/reference/getLangs.html
const LanguagesURL = "https://dictionary.yandex.net/api/v1/dicservice.json/getLangs"

// Languages is a  list of dictionary's languages.
// It is sorted in ascending order, example: `["en-en", "en-ru"]`.
type Languages []string

// LoadLanguages loads available dictionary languages.
func LoadLanguages(ctx context.Context, client *http.Client, cfg *config.Config) (*Languages, error) {
	params := &url.Values{"key": {cfg.Dictionary}}
	data := strings.NewReader(params.Encode())

	body, err := cloud.Request(ctx, client, data, cfg.GetURL(LanguagesURL), "", cfg.UserAgent, false, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get dictionary languages: %w", err)
	}

	languages := &Languages{}
	if err = json.Unmarshal(body, languages); err != nil {
		return nil, fmt.Errorf("failed to decode dictionary languages response: %w", err)
	}

	languages.Sort()
	return languages, nil
}

// String is an implementation of String() method for Languages pointer (LangChecker interface).
func (languages *Languages) String() string {
	return fmt.Sprintf("Dictionary languages:\n%v", strings.Join(*languages, ", "))
}

// Len returns a list of dictionary languages.
func (languages *Languages) Len() int {
	if languages == nil {
		return 0
	}
	return len(*languages)
}

// Sort sorts dictionary languages.
func (languages *Languages) Sort() {
	sort.Slice(*languages, func(i, j int) bool {
		return (*languages)[i] < (*languages)[j]
	})
}

// Contains is an implementation of Contains() method for Languages pointer (LangChecker interface).
func (languages *Languages) Contains(fromLanguage, toLanguage string) bool {
	length := languages.Len()
	if length == 0 {
		return false
	}

	search := strings.ToLower(fmt.Sprintf("%s-%s", fromLanguage, toLanguage))
	data := *languages

	i := sort.Search(length, func(i int) bool { return data[i] >= search })
	return i < length && data[i] == search
}

// Description is an implementation of Description() method for
// Languages pointer (LangChecker interface).
func (languages *Languages) Description() string {
	return fmt.Sprintf("Length=%v\n%v", len(*languages), languages.String())
}
