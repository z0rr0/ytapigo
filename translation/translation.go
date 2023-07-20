// Package translation implements console text multi-word translation.
package translation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/z0rr0/ytapigo/cloud"
	"github.com/z0rr0/ytapigo/config"
)

// URL is a URL for translation request.
// Documentation https://cloud.yandex.com/en/docs/translate/api-ref/Translation/translate
const URL = "https://translate.api.cloud.yandex.net/translate/v2/translate"

// ResponseItem is an item for translation request.
type ResponseItem struct {
	Text                 string `json:"text"`
	DetectedLanguageCode string `json:"detectedLanguageCode"`
}

// Response is a type of translation response.
type Response struct {
	Translations []ResponseItem `json:"translations"`
}

// String is an implementation of String() method for Response pointer.
func (t *Response) String() string {
	if len(t.Translations) == 0 {
		return ""
	}
	texts := make([]string, len(t.Translations))
	for i := range t.Translations {
		texts[i] = t.Translations[i].Text
	}
	return strings.Join(texts, "\n")
}

// Exists is an implementation of Exists() method for Response pointer.
func (t *Response) Exists() bool {
	return t.String() != ""
}

// Request is a type of translation request.
type Request struct {
	FolderID           string   `json:"folder_id"`
	Texts              []string `json:"texts"`
	TargetLanguageCode string   `json:"targetLanguageCode"`
	SourceLanguageCode string   `json:"sourceLanguageCode"`
}

// Translate returns translated text.
func Translate(ctx context.Context, client *http.Client, cfg *config.Config, r *Request) (*Response, error) {
	err := cfg.InitToken(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to init IAM token to translation: %w", err)
	}

	data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal translation request: %w", err)
	}

	reader := bytes.NewReader(data)
	url := cfg.GetURL(URL)

	body, err := cloud.Request(ctx, client, reader, url, cfg.Translation.IAMToken, cfg.UserAgent, true, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to translate: %w", err)
	}

	response := &Response{}
	if err = json.Unmarshal(body, response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal translation response: %w", err)
	}

	return response, nil
}
