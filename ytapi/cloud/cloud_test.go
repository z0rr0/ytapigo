// Copyright (c) 2019, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cloud contains WJT cloud methods.
// Based on https://cloud.yandex.ru/docs/iam/operations/iam-token/create-for-sa
package cloud

import (
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
	"time"
)

const (
	ua = "test/1.0"
)

var (
	timeout = time.Second
	logger  = log.New(os.Stdout, "TEST", log.Ldate|log.Lmicroseconds|log.Lshortfile)
)

func getKey(name string, t *testing.T) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	f, err := os.Create(name)
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
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
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	client := s.Client()
	requestData := strings.NewReader(`{"jwt":"abc"}`)
	data, err := Request(client, requestData, s.URL, "", ua, timeout, logger, logger)
	if err != nil {
		t.Fatal(err)
	}
	token := &Token{}
	err = json.Unmarshal(data, token)
	if err != nil {
		t.Error(err)
	}
	if iamt := token.IAMToken; iamt != tokenValue {
		t.Errorf("failed token: %v != %v", iamt, tokenValue)
	}
}

func TestAccount_SetIAMToken(t *testing.T) {
	const tokenValue = "abc123"

	fileName := path.Join(os.TempDir(), "ytapigo_test.pem")
	err := getKey(fileName, t)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(fileName); err != nil {
			t.Error(err)
		}
	}()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{"iamToken":"%s","expiresAt":"2019-02-15T01:09:43.418711Z"}`, tokenValue)
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer s.Close()

	client := s.Client()
	account := &Account{
		FolderID:         "123",
		KeyID:            "456",
		ServiceAccountID: "789",
		KeyFile:          fileName,
	}
	URL = s.URL // overwrite default URL
	err = account.SetIAMToken("", client, ua, timeout, logger, logger)
	if err != nil {
		t.Error(err)
	}
	if account.IAMToken != tokenValue {
		t.Errorf("failed token value: %v", account.IAMToken)
	}
}
