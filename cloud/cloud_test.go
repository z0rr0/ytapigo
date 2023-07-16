package cloud

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

const userAgent = "test/1.0"

var logger = log.New(os.Stdout, "TEST ", log.Lmicroseconds|log.Lshortfile)

func generateKey(name string, t *testing.T) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	f, err := os.Create(name)
	if err != nil {
		return err
	}

	defer func() {
		if e := f.Close(); e != nil {
			t.Error(e)
		}
	}()

	privateKey := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.Encode(f, privateKey)
}

func TestRequest(t *testing.T) {
	const tokenValue = "abc123"
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{"iamToken":"%s","expiresAt":"2019-02-15T01:09:43.418711Z"}`, tokenValue)

		if _, err := fmt.Fprint(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	client := s.Client()
	requestData := strings.NewReader(`{"jwt":"abc"}`)
	ctx := context.Background()

	data, err := Request(ctx, client, requestData, s.URL, "", userAgent, true, logger)
	if err != nil {
		t.Fatal(err)
	}

	token := &Token{}
	if err = json.Unmarshal(data, token); err != nil {
		t.Fatal(err)
	}

	if iamt := token.IAMToken; iamt != tokenValue {
		t.Errorf("failed token: %v != %v", iamt, tokenValue)
	}
}

func TestAccount_SetIAMToken(t *testing.T) {
	const tokenValue = "abc123"

	fileName := path.Join(os.TempDir(), "ytapigo_test.pem")
	err := generateKey(fileName, t)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if e := os.Remove(fileName); e != nil {
			t.Error(e)
		}
	}()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{"iamToken":"%s","expiresAt":"2019-02-15T01:09:43.418711Z"}`, tokenValue)

		if _, e := fmt.Fprint(w, response); e != nil {
			t.Error(e)
		}
	}))
	defer s.Close()

	client := s.Client()
	ctx := context.Background()
	account := &Account{
		FolderID:         "123",
		KeyID:            "456",
		ServiceAccountID: "789",
		KeyFile:          fileName,
	}

	err = account.SetIAMToken(ctx, client, userAgent, logger, s.URL)
	if err != nil {
		t.Fatal(err)
	}

	if account.IAMToken != tokenValue {
		t.Errorf("failed token value: %v", account.IAMToken)
	}
}
