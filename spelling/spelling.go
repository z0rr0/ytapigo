// Package spelling implements text spell checks.
package spelling

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/z0rr0/ytapigo/cloud"
	"github.com/z0rr0/ytapigo/config"
)

// URL is API URL for spell check.
// Documentation https://yandex.ru/dev/speller/doc/ru/reference/checkText
const URL = "https://speller.yandex.net/services/spellservice.json/checkText"

// pre-defined languages to don't do extra HTTP requests
var availableLanguages = map[string]struct{}{"en": {}, "ru": {}, "uk": {}}

// Item is a type of spell check (from JSON API response).
type Item struct {
	Word string   `json:"word"`
	S    []string `json:"s"`
	Code float64  `json:"code"`
	Pos  float64  `json:"pos"`
	Row  float64  `json:"row"`
	Col  float64  `json:"col"`
	Len  float64  `json:"len"`
}

// Response is an array of spelling results.
type Response []Item

// Exists is an implementation of Exists() method for Response.
func (s *Response) Exists() bool {
	return s != nil && len(*s) > 0
}

// String is an implementation of String() method for Response.
func (s *Response) String() string {
	n := len(*s)
	if n == 0 {
		return ""
	}

	items := make([]string, 0, n)
	for _, v := range *s {
		if v.Exists() {
			items = append(items, v.String())
		}
	}
	return fmt.Sprintf("Spelling: \n\t%s", strings.Join(items, "\n\t"))
}

// Exists is an implementation of Exists() method for Item.
func (si *Item) Exists() bool {
	return (len(si.Word) > 0) || (len(si.S) > 0)
}

// String is an implementation of String() method for Item.
func (si *Item) String() string {
	return fmt.Sprintf("%v -> %v", si.Word, si.S)
}

func reader(lang, text string) io.Reader {
	params := url.Values{
		"lang":    {lang},
		"text":    {text},
		"format":  {"plain"},
		"options": {"518"},
	}
	return strings.NewReader(params.Encode())
}

// Request does a request to spelling check API.
func Request(ctx context.Context, client *http.Client, lang, text string, cfg *config.Config) (*Response, error) {
	if _, ok := availableLanguages[lang]; !ok {
		return nil, nil // skip, spelling check is not available for this language
	}

	body, err := cloud.Request(ctx, client, reader(lang, text), cfg.GetURL(URL), "", cfg.UserAgent, false, cfg.Logger)
	if err != nil {
		return nil, err
	}

	result := &Response{}
	if err = json.Unmarshal(body, result); err != nil {
		return nil, fmt.Errorf("failed to decode spelling response: %w", err)
	}

	return result, nil
}
