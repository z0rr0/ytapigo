package spelling

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

func TestRequest(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `[{"code": 1,"pos": 0,"row": 0,"col": 0,"len": 6,"word": "малоко","s": ["молоко","молока","малого"]}]`

		if _, err := fmt.Fprint(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	cfg := &config.Config{
		URL:    map[string]string{URL: s.URL},
		Logger: log.New(os.Stdout, "TEST ", log.Lmicroseconds|log.Lshortfile),
	}

	resp, err := Request(context.Background(), s.Client(), "ru", "малоко", cfg)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	if resp == nil {
		t.Fatal("response is nil")
	}

	expected := "Spelling:\n\tмалоко -> [молоко молока малого]"

	if respString := resp.String(); respString != expected {
		t.Errorf("expected: %q , got: %q", expected, respString)
	}
}
