package translation

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
)

var logger = log.New(os.Stdout, "TEST ", log.Lmicroseconds|log.Lshortfile)

func TestTranslate(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, e := fmt.Fprint(w, `{"translations":[{"text": "пора начинать","detectedLanguageCode":"ru"}]}`); e != nil {
			t.Error(e)
		}
	}))
	defer s.Close()

	cfg := &config.Config{
		Translation: cloud.Account{FolderID: "folder_id", IAMToken: "token"},
		URL:         map[string]string{URL: s.URL},
		Logger:      logger,
	}
	req := &Request{
		FolderID:           "folder_id",
		Texts:              []string{"time to start"},
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

	expected := "пора начинать"
	if rs := resp.String(); rs != expected {
		t.Errorf("expected %q, got %q", expected, rs)
	}
}
