package handle

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/z0rr0/ytapigo/cloud"
	"github.com/z0rr0/ytapigo/config"
	"github.com/z0rr0/ytapigo/dictionary"
	"github.com/z0rr0/ytapigo/spelling"
	"github.com/z0rr0/ytapigo/translation"
)

var logger = log.New(os.Stdout, "TEST ", log.Lmicroseconds|log.Lshortfile)

func testServer(t *testing.T) *httptest.Server {
	var (
		langResponse          = `{"languages":[{"code":"ru","name":"Russian"},{"code":"en","name":"English"}]}`
		translateResponse     = `{"translations":[{"text": "пора начинать","detectedLanguageCode":"ru"}]}`
		detectResponse        = `{"languageCode":"en"}`
		spellingResponse      = `[{"code": 1,"pos": 0,"row": 0,"col": 0,"len": 6,"word": "малоко","s": ["молоко","молока","малого"]}]`
		langDictResponse      = `["en-en", "en-ru"]`
		translateDictResponse = `
{ "head": {},
  "def": [
     { "text": "time", "pos": "noun", "ts": "taɪm",
       "tr": [
          { "text": "время", "pos": "noun", "gen": "ср",
            "syn": [
				{"text":"раз","pos":"noun","gen":"м","fr":10},
				{"text":"момент","pos":"noun","gen":"м","fr":5}
			],
            "mean": [{ "text": "timing" },{ "text": "fold" },{ "text": "half"}],
            "ex" : [
               { "text": "prehistoric time",
                 "tr": [{"text": "доисторическое время"}]
               },
               { "text": "hundredth time",
                 "tr": [{"text": "сотый раз"}]
               },
               { "text": "time-slot",
                 "tr": [{ "text": "тайм-слот" }]
               }
            ]
          }
       ]
    }
  ]
}`
	)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response string
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/dicservice.json/getLangs":
			response = langDictResponse
		case "/api/v1/dicservice.json/lookup":
			response = translateDictResponse
		case "/translate/v2/languages":
			response = langResponse
		case "/translate/v2/translate":
			response = translateResponse
		case "/translate/v2/detect":
			response = detectResponse
		case "/services/spellservice.json/checkText":
			response = spellingResponse
		}

		if _, e := fmt.Fprint(w, response); e != nil {
			t.Error(e)
		}
	}))
}

func TestHandler_Run(t *testing.T) {
	s := testServer(t)
	defer s.Close()

	cfg := &config.Config{
		Translation: cloud.Account{FolderID: "folder_id", IAMToken: "token"},
		Logger:      logger,
		URL: map[string]string{
			dictionary.LanguagesURL:       s.URL + "/api/v1/dicservice.json/getLangs",
			dictionary.TranslationURL:     s.URL + "/api/v1/dicservice.json/lookup",
			translation.LanguagesURL:      s.URL + "/translate/v2/languages",
			translation.URL:               s.URL + "/translate/v2/translate",
			translation.DetectLanguageURL: s.URL + "/translate/v2/detect",
			spelling.URL:                  s.URL + "/services/spellservice.json/checkText",
		},
	}

	h := New(cfg)
	h.client = s.Client()

	testCases := []struct {
		name      string
		direction string
		params    []string
		error     string
	}{
		{name: "empty", error: "empty text"},
		{name: "space_params", error: "empty text", params: []string{"  ", ""}},
		{name: "dictionary", params: []string{"time"}},
		{name: "dictionary_with_spaces", params: []string{"  ", "  time", ""}},
		{name: "dictionary_auto", direction: AutoLanguageDetect, params: []string{"time"}},
		{name: "translation", params: []string{"time to start"}},
		{name: "translation_separate", params: []string{"time", "to", "start"}},
		{name: "translation_auto", direction: AutoLanguageDetect, params: []string{"time to start"}},
		{
			name:      "dictionary_err",
			direction: "de-fr",
			params:    []string{"time"},
			error:     "unknown language direction: de -> fr",
		},
		{
			name:      "translation_err",
			direction: "de-fr",
			params:    []string{"time to start"},
			error:     "unknown language direction: de -> fr",
		},
	}

	for _, tc := range testCases {
		err := h.Run(context.Background(), tc.direction, tc.params)

		if err != nil {
			if e := err.Error(); e != tc.error {
				t.Errorf("expected error %q, got %q", tc.error, e)
			}
			continue
		}
	}
}

func TestBuildText(t *testing.T) {
	tests := []struct {
		params       []string
		expected     string
		isDictionary bool
		withError    bool
	}{
		{withError: true},
		{params: []string{}, withError: true},
		{
			params:   []string{"Hello", "world"},
			expected: "Hello world",
		},
		{
			params:   []string{"This", "is", "a", "test"},
			expected: "This is a test",
		},
		{
			params:   []string{" This", "  is\n", " \ta ", " test   "},
			expected: "This is a test",
		},
		{
			params:       []string{"  ", "test"},
			expected:     "test",
			isDictionary: true,
		},
		{
			params:       []string{"  ", "  book ", "   \n"},
			expected:     "book",
			isDictionary: true,
		},
	}

	for _, test := range tests {
		text, isDictionary, err := buildText(test.params)
		withoutError := !test.withError

		if err != nil && withoutError {
			t.Errorf("unexpected error: %v", err)
		}

		if withoutError && text != test.expected {
			t.Errorf("expected text %q, but got %q", test.expected, text)
		}

		if withoutError && isDictionary != test.isDictionary {
			t.Errorf("expected isDictionary %v, but got %v", test.isDictionary, isDictionary)
		}
	}
}
