package handle

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
	var err error

	y.text, y.isDictionary, err = buildText(params)
	if err != nil {
		return err
	}

	err = y.setLanguages(ctx, direction)
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

	var languages result.Languages

	if y.isDictionary {
		languages, err = dictionary.LoadLanguages(ctx, y.client, y.config)
	} else {
		languages, err = translation.LoadLanguages(ctx, y.client, y.config)
	}

	if err != nil {
		return fmt.Errorf("can not set languages: %w", err)
	}

	if !languages.Contains(fromLanguage, toLanguage) {
		return fmt.Errorf("unknown language direction: %s -> %s", fromLanguage, toLanguage)
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

// buildText parses and builds text from parameters.
// It returns result text, true if it is a dictionary request and error.
func buildText(params []string) (string, bool, error) {
	var (
		builder strings.Builder
		count   uint
	)

	for _, p := range params {
		for _, word := range strings.Split(p, " ") {
			if w := strings.Trim(word, " \t\n\r"); len(w) > 0 {
				builder.WriteString(w)
				builder.WriteString(" ")
				count++
			}
		}
	}

	if count == 0 {
		// not found any words for translation
		return "", false, fmt.Errorf("empty text")
	}

	return strings.TrimSuffix(builder.String(), " "), count == 1, nil
}
