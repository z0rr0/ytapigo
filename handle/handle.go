package handle

import (
	"context"
	"fmt"
	"net/http"

	"github.com/z0rr0/ytapigo/arguments"
	"github.com/z0rr0/ytapigo/config"
	"github.com/z0rr0/ytapigo/dictionary"
	"github.com/z0rr0/ytapigo/result"
	"github.com/z0rr0/ytapigo/spelling"
	"github.com/z0rr0/ytapigo/translation"
)

// Handler is a common meta-data storage for translation and spelling check requests.
type Handler struct {
	config       *config.Config
	client       *http.Client
	isDictionary bool
	text         string
	fromLanguage string
	toLanguage   string
}

// New creates a new handler.
func New(cfg *config.Config) *Handler {
	return &Handler{
		config: cfg,
		client: &http.Client{Transport: &http.Transport{Proxy: cfg.Proxy}},
	}
}

// Run runs translation, spelling check and prints their results.
func (y *Handler) Run(ctx context.Context, direction string, params []string) error {
	y.text, y.isDictionary = arguments.TextWithDictionary(params)

	err := y.setLanguages(ctx, direction)
	if err != nil {
		return err
	}

	// concurrent translation and spelling check requests
	results, err := y.translationAndSpelling(ctx)
	if err != nil {
		return err
	}

	result.Show(results)
	return nil
}

// loadLanguages loads languages defined by dictionary or translation API will be used.
func (y *Handler) loadLanguages(ctx context.Context) (result.Languages, error) {
	if y.isDictionary {
		return dictionary.LoadLanguages(ctx, y.client, y.config)
	}
	return translation.LoadLanguages(ctx, y.client, y.config)
}

// setLanguages detects language direction.
func (y *Handler) setLanguages(ctx context.Context, direction string) error {
	fromLanguage, toLanguage, err := y.detectLanguages(ctx, direction, y.text)
	if err != nil {
		return err
	}

	if knowLanguages(fromLanguage, toLanguage) {
		y.fromLanguage, y.toLanguage = fromLanguage, toLanguage
		return nil
	}

	languages, err := y.loadLanguages(ctx)
	if err != nil {
		return fmt.Errorf("can not set languages: %w", err)
	}

	if !languages.Contains(fromLanguage, toLanguage) {
		return fmt.Errorf("unknown language direction: %s -> %s\n%v", fromLanguage, toLanguage, languages.String())
	}

	y.fromLanguage, y.toLanguage = fromLanguage, toLanguage
	return nil
}

// translationAndSpelling runs translation and spelling check requests concurrently.
func (y *Handler) translationAndSpelling(ctx context.Context) ([]result.Translation, error) {
	const handlersCount = 2

	ch := make(chan result.Item, 1)
	defer close(ch)

	go func() {
		t, e := spelling.Request(ctx, y.client, y.fromLanguage, y.text, y.config)
		ch <- result.Item{Translation: t, Priority: 1, Err: e}
	}()

	go func() {
		t, e := y.translation(ctx)
		ch <- result.Item{Translation: t, Priority: 2, Err: e}
	}()

	return result.Build(ch, handlersCount)
}

// translation does translation API request.
func (y *Handler) translation(ctx context.Context) (result.Translation, error) {
	if y.isDictionary {
		request := &dictionary.Request{
			Key:                y.config.Dictionary,
			Text:               y.text,
			TargetLanguageCode: y.toLanguage,
			SourceLanguageCode: y.fromLanguage,
		}
		return dictionary.Translate(ctx, y.client, y.config, request)
	}

	request := &translation.Request{
		FolderID:           y.config.Translation.FolderID,
		Texts:              []string{y.text},
		SourceLanguageCode: y.fromLanguage,
		TargetLanguageCode: y.toLanguage,
	}
	return translation.Translate(ctx, y.client, y.config, request)
}
