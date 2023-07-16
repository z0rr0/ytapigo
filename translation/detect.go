package translation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/z0rr0/ytapigo/cloud"
	"github.com/z0rr0/ytapigo/config"
)

// DetectLanguageURL is a URL for API detect language request.
// Documentation https://cloud.yandex.com/en/docs/translate/api-ref/Translation/detectLanguage
const DetectLanguageURL = "https://translate.api.cloud.yandex.net/translate/v2/detect"

// Detect is a type of language detection request.
type Detect struct {
	LanguageCode string `json:"languageCode"`
}

// DetectRequest is a type of detect language request.
type DetectRequest struct {
	FolderID          string   `json:"folder_id"`
	Text              string   `json:"text"`
	LanguageCodeHints []string `json:"languageCodeHints"`
}

// detectRequestData prepares detect language request data.
func detectRequestData(cfg *config.Config, text string) (io.Reader, error) {
	r := &DetectRequest{
		FolderID: cfg.Translation.FolderID,
		Text:     text,
	}

	data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal detect request: %w", err)
	}

	return bytes.NewReader(data), nil
}

// DetectLanguage returns automatically detected language.
func DetectLanguage(ctx context.Context, client *http.Client, cfg *config.Config, text string) (string, error) {
	err := cfg.InitToken(ctx, client)
	if err != nil {
		return "", fmt.Errorf("failed to init IAM token: %w", err)
	}

	data, err := detectRequestData(cfg, text)
	if err != nil {
		return "", fmt.Errorf("failed to get detect request data: %w", err)
	}

	body, err := cloud.Request(ctx, client, data, cfg.GetURL(DetectLanguageURL), cfg.Translation.IAMToken, cfg.UserAgent, true, cfg.Logger)
	if err != nil {
		return "", fmt.Errorf("failed to get detected language: %w", err)
	}

	detect := Detect{}
	if err = json.Unmarshal(body, &detect); err != nil {
		return "", fmt.Errorf("failed to decode detect language response: %w", err)
	}

	cfg.Logger.Printf("detected language: %s", detect.LanguageCode)
	return detect.LanguageCode, nil
}
