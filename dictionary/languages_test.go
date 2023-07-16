package dictionary

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/z0rr0/ytapigo/config"
)

var logger = log.New(os.Stdout, "TEST ", log.Lmicroseconds|log.Lshortfile)

func TestLoadLanguages(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, e := fmt.Fprint(w, `["en-en", "en-ru"]`); e != nil {
			t.Error(e)
		}
	}))
	defer s.Close()

	cfg := &config.Config{Dictionary: "test", URL: map[string]string{LanguagesURL: s.URL}, Logger: logger}
	languages, err := LoadLanguages(context.Background(), s.Client(), cfg)

	if err != nil {
		t.Fatal(err)
	}

	if languages.Len() != 2 {
		t.Error("wrong languages count")
	}

	expected := "Dictionary languages:\n" + "en-en, en-ru"
	if tmp := languages.String(); tmp != expected {
		t.Errorf("wrong languages string: %s", tmp)
	}

	expected = "Length=2\n" + expected
	if tmp := languages.Description(); tmp != expected {
		t.Errorf("wrong languages description: %s", tmp)
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
		{languages: Languages{"en-en", "en-ru"}, fromLang: "en", toLang: "en", expected: true},
		{languages: Languages{"en-en", "en-ru"}, fromLang: "EN", toLang: "EN", expected: true},
		{languages: Languages{"en-en", "en-ru"}, fromLang: "en", toLang: "ru", expected: true},
		{languages: Languages{"en-en", "en-ru"}, fromLang: "en", toLang: "RU", expected: true},
		{languages: Languages{"en-en", "en-ru"}, fromLang: "eN", toLang: "En", expected: true},
		{languages: Languages{"en-en", "en-ru"}, fromLang: "de", toLang: "en"},
		{languages: Languages{"en-en", "en-ru"}, fromLang: "de"},
		{languages: Languages{"en-en", "en-ru"}, toLang: "en"},
	}

	for i, tc := range testCases {
		if tc.languages.Contains(tc.fromLang, tc.toLang) != tc.expected {
			t.Errorf("%d: wrong result", i)
		}
	}
}
