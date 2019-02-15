// Copyright (c) 2019, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package auth contains WJT auth methods.
// Based on https://cloud.yandex.ru/docs/iam/operations/iam-token/create-for-sa
package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	// TTL is token live period.
	TTL = time.Hour
	// URL is URL for authentication requests.
	URL = "https://iam.api.cloud.yandex.net/iam/v1/tokens"
)

var ps256WithSaltLengthEqualsHash = &jwt.SigningMethodRSAPSS{
	SigningMethodRSA: jwt.SigningMethodPS256.SigningMethodRSA,
	Options:          &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash},
}

// Account is API auth struct info.
type Account struct {
	FolderID         string `json:"folder_id"`
	KeyID            string `json:"key_id"`
	ServiceAccountID string `json:"service_account_id"`
	KeyFile          string `json:"key_file"`
	IAMToken         string
}

// Token is iam token struct.
type Token struct {
	IAMToken string `json:"iamToken"`
}

// loadPrivateKey reads and parses RSA key file.
func (a *Account) loadPrivateKey() (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(a.KeyFile)
	if err != nil {
		return nil, err
	}
	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		return nil, err
	}
	return rsaPrivateKey, nil
}

// signedToken prepares JWT signed token.
func (a *Account) signedToken() (string, error) {
	issuedAt := time.Now()
	token := jwt.NewWithClaims(ps256WithSaltLengthEqualsHash, jwt.StandardClaims{
		Issuer:    a.ServiceAccountID,
		IssuedAt:  issuedAt.Unix(),
		ExpiresAt: issuedAt.Add(TTL).Unix(),
		Audience:  URL,
	})
	token.Header["kid"] = a.KeyID
	privateKey, err := a.loadPrivateKey()
	if err != nil {
		return "", err
	}
	return token.SignedString(privateKey)
}

// getCachedToken tries to read cached token
func getCachedToken(path string) (string, error) {
	f, err := os.Stat(path)
	if err != nil {
		return "", nil
	}
	// is cache expired?
	if f.ModTime().Add(TTL).Before(time.Now()) {
		return "", nil
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// saveCacheToken tries to save cached toke to a file.
func saveCacheToken(path, token string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	_, err = f.WriteString(token)
	if err != nil {
		return err
	}
	return f.Close()
}

// SetIAMToken gets iam token and stores it to Account a.
func (a *Account) SetIAMToken(cacheFile string, client *http.Client, userAgent string, timeout time.Duration, li, le *log.Logger) error {
	if cacheFile != "" {
		value, err := getCachedToken(cacheFile)
		if err != nil {
			return err
		}
		if value != "" {
			a.IAMToken = value
			return nil
		}
		// no cache
	}

	jot, err := a.signedToken()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", URL, strings.NewReader(fmt.Sprintf(`{"jwt":"%s"}`, jot)))
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req = req.WithContext(ctx)

	ec := make(chan error)
	li.Printf("auth request: %v\n", URL)
	var resp *http.Response
	go func() {
		resp, err = client.Do(req)
		ec <- err
		close(ec)
	}()
	select {
	case <-ctx.Done():
		<-ec // wait error "context deadline exceeded"
		return fmt.Errorf("token get timed out (%v)", timeout)
	case err := <-ec:
		if err != nil {
			return err
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			le.Printf("failed body close: %v\n", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("get token status %v, can't read content: %v", resp.Status, err)
		}
		return fmt.Errorf("get token status %s: %s", resp.Status, body)
	}
	token := &Token{}
	err = json.NewDecoder(resp.Body).Decode(token)
	if err != nil {
		return err
	}
	a.IAMToken = token.IAMToken
	// save cache if it's needed
	if cacheFile != "" {
		go func() {
			if err := saveCacheToken(cacheFile, a.IAMToken); err != nil {
				le.Printf("can't save token cache [%v]: %v\n", cacheFile, err)
			}
		}()
	}
	return nil
}
