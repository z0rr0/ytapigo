// Package cloud contains WJT cloud methods.
// Based on https://cloud.yandex.ru/docs/iam/operations/iam-token/create-for-sa
package cloud

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// TTL is token live period.
	TTL = time.Hour

	// TokenURL is URL for authentication requests.
	TokenURL = "https://iam.api.cloud.yandex.net/iam/v1/tokens"
)

var ps256WithSaltLengthEqualsHash = &jwt.SigningMethodRSAPSS{
	SigningMethodRSA: jwt.SigningMethodPS256.SigningMethodRSA,
	Options:          &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash},
}

// Account is API cloud struct info.
type Account struct {
	FolderID         string `json:"folder_id"`
	KeyID            string `json:"key_id"`
	ServiceAccountID string `json:"service_account_id"`
	KeyFile          string `json:"key_file"`
	IAMToken         string
}

// Token is iam token struct.
type Token struct {
	IAMToken  string `json:"iamToken"`
	ExpiresAt string `json:"expiresAt"`
}

// loadPrivateKey reads and parses RSA key file.
func (a *Account) loadPrivateKey() (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(a.KeyFile)
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
	issuedAt := time.Now().UTC()
	clams := &jwt.RegisteredClaims{
		Issuer:    a.ServiceAccountID,
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(issuedAt.Add(TTL).UTC()),
		Audience:  jwt.ClaimStrings{TokenURL},
	}
	token := jwt.NewWithClaims(ps256WithSaltLengthEqualsHash, clams)

	token.Header["kid"] = a.KeyID

	privateKey, err := a.loadPrivateKey()
	if err != nil {
		return "", err
	}

	return token.SignedString(privateKey)
}

// SetIAMToken gets iam token and stores it to Account.
func (a *Account) SetIAMToken(ctx context.Context, client *http.Client, userAgent string, logger *log.Logger, url string) error {
	jot, err := a.signedToken()
	if err != nil {
		return fmt.Errorf("failed to get sigend token: %w", err)
	}

	if url == "" {
		url = TokenURL
	}

	data := strings.NewReader(fmt.Sprintf(`{"jwt":"%s"}`, jot))
	body, err := Request(ctx, client, data, url, "", userAgent, true, logger)

	if err != nil {
		return fmt.Errorf("failed to get iam token: %w", err)
	}

	token := &Token{}
	if err = json.Unmarshal(body, token); err != nil {
		return err
	}

	a.IAMToken = token.IAMToken
	return nil
}

func buildRequest(ctx context.Context, data io.Reader, uri, bearer, userAgent string, isJSON bool) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, data)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	if isJSON {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if bearer != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearer))
	}

	return req, nil
}

// Request does POST request.
func Request(ctx context.Context, client *http.Client, data io.Reader, uri, bearer, userAgent string, isJSON bool, logger *log.Logger) ([]byte, error) {
	req, err := buildRequest(ctx, data, uri, bearer, userAgent, isJSON)
	if err != nil {
		return nil, fmt.Errorf("can't create request: %w", err)
	}

	start := time.Now()
	defer func() {
		if logger != nil {
			logger.Printf("%s [%v] %s", req.Method, time.Since(start).Truncate(time.Millisecond), req.URL)
		}
	}()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't do request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("request status %v, can't read content: %v", resp.Status, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request status %s: %s", resp.Status, body)
	}

	return body, nil
}
