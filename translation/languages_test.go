package translation

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/z0rr0/ytapigo/cloud"
	"github.com/z0rr0/ytapigo/config"
)

func TestLoadLanguages(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, e := fmt.Fprint(w, `{"languages":[{"code":"af","name":"Afrikaans"},{"code":"en","name":"English"}]}`); e != nil {
			t.Error(e)
		}
	}))
	defer s.Close()

	cfg := &config.Config{
		Translation: cloud.Account{FolderID: "folder_id", IAMToken: "token"},
		URL:         map[string]string{LanguagesURL: s.URL},
		Logger:      logger,
	}
	languages, err := LoadLanguages(context.Background(), s.Client(), cfg)

	if err != nil {
		t.Fatal(err)
	}

	if languages.Len() != 2 {
		t.Error("wrong languages count")
	}

	if n := len(languages.Codes); n != 2 {
		t.Errorf("wrong languages codes count: %d", n)
	}

	expected := "af - Afrikaans            en - English             "
	if tmp := languages.Description(); tmp != expected {
		t.Errorf("wrong languages description: %q", tmp)
	}

	expected = "Translation languages:\naf, en\n" + expected
	if tmp := languages.String(); tmp != expected {
		t.Errorf("wrong languages string: %q", tmp)
	}
}

func TestLanguages_Contains(t *testing.T) {
	testCases := []struct {
		languages Languages // sorted languages slice
		fromLang  string
		toLang    string
		expected  bool
	}{
		{},
		{languages: Languages{}, fromLang: "en", toLang: "en"},
		{languages: Languages{}, fromLang: "en", toLang: "ru"},
		{fromLang: "en", toLang: "ru"},
		{fromLang: "en"},
		{toLang: "ru"},
		{languages: Languages{Codes: map[string]struct{}{"en": {}}}, fromLang: "en", toLang: "en", expected: true},
		{languages: Languages{Codes: map[string]struct{}{"en": {}}}, fromLang: "En", toLang: "eN", expected: true},
		{
			languages: Languages{Codes: map[string]struct{}{"en": {}, "ru": {}}},
			fromLang:  "en",
			toLang:    "en",
			expected:  true,
		},
		{
			languages: Languages{Codes: map[string]struct{}{"en": {}, "ru": {}}},
			fromLang:  "en",
			toLang:    "ru",
			expected:  true,
		},
		{
			languages: Languages{Codes: map[string]struct{}{"en": {}, "ru": {}}},
			fromLang:  "ru",
			toLang:    "EN",
			expected:  true,
		},
		{
			languages: Languages{Codes: map[string]struct{}{"en": {}, "ru": {}}},
			fromLang:  "de",
			toLang:    "en",
		},
		{
			languages: Languages{Codes: map[string]struct{}{"de": {}, "ru": {}}},
			fromLang:  "en",
			toLang:    "en",
		},
	}

	for i, tc := range testCases {
		if tc.languages.Contains(tc.fromLang, tc.toLang) != tc.expected {
			t.Errorf("%d: wrong result", i)
		}
	}
}
