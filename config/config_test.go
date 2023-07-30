package config

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/z0rr0/ytapigo/cloud"
)

const userAgent = "test/1.0"

var (
	logger = log.New(os.Stdout, "TEST ", log.Lmicroseconds|log.Lshortfile)

	testConfig = `
{
  "user_agent": "ytapigo_test/1.0",
  "proxy_url": "http://user:password@127.0.0.1:54321",
  "dictionary": "dict_key",
  "auth_cache": "ytapigo_not_exists.json",
  "debug": true,
  "translation": {
    "folder_id": "translation_folder_id",
    "key_id": "translation_key_id",
    "service_account_id": "translation_service_account_id",
    "key_file": "rsa_key_file"
  }
}`
)

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

func TestNew(t *testing.T) {
	configFile := path.Join(os.TempDir(), "ytapigo_config_new.toml")

	err := os.WriteFile(configFile, []byte(testConfig), 0660)
	if err != nil {
		t.Fatal(err)
	}
	defer deleteFile(configFile, t)

	cfg, err := New(configFile, "", "", true, true, logger)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.UserAgent != "ytapigo_test/1.0" {
		t.Errorf("failed user agent, got %q", cfg.UserAgent)
	}

	if cfg.ProxyURL != "http://user:password@127.0.0.1:54321" {
		t.Errorf("failed proxy, got %q", cfg.ProxyURL)
	}

	if cfg.Dictionary != "dict_key" {
		t.Errorf("failed dictionary, got %q", cfg.Dictionary)
	}

	if cfg.AuthCache != "ytapigo_not_exists.json" {
		t.Errorf("failed auth cache, got %q", cfg.AuthCache)
	}

	if !cfg.Debug {
		t.Errorf("failed debug, got %v", cfg.Debug)
	}

	if cfg.Translation.FolderID != "translation_folder_id" {
		t.Errorf("failed translation folder, got %q", cfg.Translation.FolderID)
	}

	if cfg.Translation.KeyID != "translation_key_id" {
		t.Errorf("failed translation key id, got %q", cfg.Translation.KeyID)
	}

	if cfg.Translation.ServiceAccountID != "translation_service_account_id" {
		t.Errorf("failed translation service account id, got %q", cfg.Translation.ServiceAccountID)
	}

	if cfg.Translation.KeyFile != "rsa_key_file" {
		t.Errorf("failed translation key file, got %q", cfg.Translation.KeyFile)
	}
}

func TestNew_NoFile(t *testing.T) {
	configFile := path.Join(os.TempDir(), "ytapigo_config_not_exists.toml")

	_, err := New(configFile, "", "", true, true, logger)
	if err == nil {
		t.Fatal(err)
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("failed error, got %v", err)
	}
}

func TestConfig_GetURL(t *testing.T) {
	testCases := []struct {
		cfgURL    string
		urlString string
		expected  string
	}{
		{},
		{urlString: "https://github.com/z0rr0/ytapigo", expected: "https://github.com/z0rr0/ytapigo"},
		{urlString: "https://github.com/z0rr0/ytapigo", cfgURL: "https://github.com", expected: "https://github.com"},
	}

	for i, tc := range testCases {
		cfg := &Config{URL: map[string]string{tc.urlString: tc.cfgURL}}

		if result := cfg.GetURL(tc.urlString); result != tc.expected {
			t.Errorf("%d: expected %q, got %q", i, tc.expected, result)
		}
	}
}

func TestConfig_InitToken(t *testing.T) {
	keyFile := path.Join(os.TempDir(), "ytapigo_config_init_token.json")

	err := generateKey(keyFile, t)
	if err != nil {
		t.Fatal(err)
	}
	defer deleteFile(keyFile, t)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, e := fmt.Fprint(w, `{"iamToken":"abc123","expiresAt":"2019-02-15T01:09:43.418711Z"}`); e != nil {
			t.Error(e)
		}
	}))
	defer s.Close()

	account := cloud.Account{
		FolderID:         "123",
		KeyID:            "456",
		ServiceAccountID: "789",
		KeyFile:          keyFile,
	}
	cfg := &Config{
		Translation: account,
		Logger:      logger,
		UserAgent:   userAgent,
		URL:         map[string]string{cloud.TokenURL: s.URL},
	}

	if err = cfg.InitToken(context.Background(), s.Client()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cfg.Translation.IAMToken != "abc123" {
		t.Errorf("unexpected IAM token: %s", cfg.Translation.IAMToken)
	}
}

func TestConfig_InitCachedToken(t *testing.T) {
	account := cloud.Account{
		FolderID:         "123",
		KeyID:            "456",
		ServiceAccountID: "789",
		KeyFile:          "not used",
		IAMToken:         "abc123",
	}
	cfg := &Config{Translation: account, Logger: logger, UserAgent: userAgent}

	if err := cfg.InitToken(context.Background(), nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cfg.Translation.IAMToken != "abc123" {
		t.Errorf("unexpected IAM token: %s", cfg.Translation.IAMToken)
	}
}

func TestConfig_setFiles(t *testing.T) {
	const testDir = "ytapigo_test_set_files"
	var (
		tmpDir        = os.TempDir()
		testConfigDir = path.Join(tmpDir, testDir, "config")
		testCacheDir  = path.Join(tmpDir, testDir, "cache")
	)
	tmpFile, err := os.CreateTemp(testCacheDir, "ytapigo_caceh_*.json")
	if err != nil {
		t.Fatal(err)
	}

	cleanFunc := func() {
		if e := os.RemoveAll(testDir); e != nil {
			t.Fatal(e)
		}
	}

	if err = os.MkdirAll(testConfigDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(testCacheDir, 0750); err != nil {
		t.Fatal(err)
	}

	// clean before and after tests
	cleanFunc()
	defer cleanFunc()

	testCases := []struct {
		name      string
		configDir string
		cacheDir  string
		keyFile   string
		cacheFile string
		expected  [2]string // config and cache files paths
		err       string
	}{
		{
			name:      "empty",
			keyFile:   "key.json",
			cacheFile: "cache.json",
			expected:  [2]string{"key.json", "cache.json"},
		},
		{
			name:      "abs_paths",
			configDir: testConfigDir,
			cacheDir:  testCacheDir,
			keyFile:   path.Join(tmpDir, "abs_path_ytapigo", "key.json"),
			cacheFile: path.Join(tmpDir, "abs_path_ytapigo", "cache.json"),
			expected: [2]string{
				path.Join(tmpDir, "abs_path_ytapigo", "key.json"),
				path.Join(tmpDir, "abs_path_ytapigo", "cache.json"),
			},
		},
		{
			name:      "empty_cache",
			configDir: testConfigDir,
			cacheDir:  testCacheDir,
			keyFile:   path.Join(tmpDir, "abs_path_ytapigo", "key.json"),
			cacheFile: "",
			expected:  [2]string{path.Join(tmpDir, "abs_path_ytapigo", "key.json"), ""},
		},
		{
			name:      "set_all",
			configDir: testConfigDir,
			cacheDir:  testCacheDir,
			keyFile:   "key.json",
			cacheFile: "cache.json",
			expected: [2]string{
				path.Join(testConfigDir, "key.json"),
				path.Join(testCacheDir, "cache.json"),
			},
		},
		{
			name:      "cache_is_not_dir",
			configDir: testConfigDir,
			cacheDir:  tmpFile.Name(),
			keyFile:   "key.json",
			cacheFile: "cache.json",
			err:       "cache is not a directory: ",
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				Translation: cloud.Account{KeyFile: tc.keyFile},
				AuthCache:   tc.cacheFile,
			}

			tcErr := cfg.setFiles(tc.configDir, tc.cacheDir)
			if tcErr != nil {
				if tc.err == "" {
					t.Errorf("unexpected error: %v", tcErr)
				} else if e := tcErr.Error(); !strings.HasPrefix(e, tc.err) {
					t.Errorf("not match error: %q, but expected %q", e, tc.err)
				}
				return
			}

			// no error from tested function
			if tc.err != "" {
				t.Errorf("expected error: %q", tc.err)
				return
			}

			// no error and no expected error
			if tc.expected[0] != cfg.Translation.KeyFile {
				t.Errorf("unexpected key file: %q", cfg.Translation.KeyFile)
			}

			if tc.expected[1] != cfg.AuthCache {
				t.Errorf("unexpected cache file: %q", cfg.AuthCache)
			}
		})
	}
}
