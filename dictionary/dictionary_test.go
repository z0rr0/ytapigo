package dictionary

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/z0rr0/ytapigo/config"
)

var response = `
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

func TestTranslate(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, e := fmt.Fprint(w, response); e != nil {
			t.Error(e)
		}
	}))
	defer s.Close()

	cfg := &config.Config{Dictionary: "test", URL: map[string]string{TranslationURL: s.URL}, Logger: logger}
	req := &Request{
		Key:                "key",
		Text:               "time",
		SourceLanguageCode: "en",
		TargetLanguageCode: "ru",
	}

	resp, err := Translate(context.Background(), s.Client(), cfg, req)
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Exists() {
		t.Error("empty response")
	}

	expected := "time [taɪm] (noun)\n\tвремя (noun)\n\tsyn: раз (noun), момент (noun)\n\t" +
		"mean: timing, fold, half\n\texamples: \n\t\t" +
		"prehistoric time: доисторическое время\n\t\t" +
		"hundredth time: сотый раз\n\t\ttime-slot: тайм-слот"

	if rs := resp.String(); rs != expected {
		t.Errorf("expected %q, got %q", expected, rs)
	}
}
