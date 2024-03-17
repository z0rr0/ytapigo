package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Cache is a struct for cache file.
type Cache struct {
	Token   string `json:"token"`
	Expired string `json:"expired"`
}

// readCachedToken reads cached token from file.
func readCachedToken(fileName string) (string, error) {
	if fileName == "" {
		return "", nil // no file name, no cache
	}

	fileName = filepath.Clean(fileName)
	data, err := os.ReadFile(fileName)

	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // no cache, probably first run
		}

		return "", fmt.Errorf("read cache file: %w", err)
	}

	c := Cache{}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return "", fmt.Errorf("unmarshal cache file: %w", err)
	}

	expiresAt, err := time.Parse(time.DateTime, c.Expired)
	if err != nil {
		return "", fmt.Errorf("parse cache file expired time: %w", err)
	}

	if time.Now().UTC().After(expiresAt) {
		return "", nil
	}

	return c.Token, nil
}

// writeCachedToken writes token to a cache file.
func writeCachedToken(fileName, token string, expiresAt time.Time) error {
	if fileName == "" {
		return nil // no file name, no cache
	}

	cache := &Cache{Token: token, Expired: expiresAt.Format(time.DateTime)}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache file: %w", err)
	}

	return os.WriteFile(fileName, data, 0600)
}
