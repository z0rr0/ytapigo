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

func TestDetectLanguage(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, e := fmt.Fprint(w, `{"languageCode":"en"}`); e != nil {
			t.Error(e)
		}
	}))
	defer s.Close()

	cfg := &config.Config{
		Translation: cloud.Account{FolderID: "folder_id", IAMToken: "token"},
		URL:         map[string]string{DetectLanguageURL: s.URL},
		Logger:      logger,
	}

	resp, err := DetectLanguage(context.Background(), s.Client(), cfg, "test")
	if err != nil {
		t.Fatal(err)
	}

	if resp != "en" {
		t.Errorf("unexpected result: %q", resp)
	}
}
