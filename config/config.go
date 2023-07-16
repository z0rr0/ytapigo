package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/z0rr0/ytapigo/cloud"
)

// Config is a struct of used services.
type Config struct {
	sync.Mutex
	Translation cloud.Account `json:"translation"`
	UserAgent   string        `json:"user_agent"`
	ProxyURL    string        `json:"proxy_url"`
	Dictionary  string        `json:"dictionary"`
	AuthCache   string        `json:"auth_cache"`
	Debug       bool          `json:"debug"`
	Proxy       func(*http.Request) (*url.URL, error)
	Logger      *log.Logger
	URL         map[string]string // override URLs map for testing only
}

// New reads configuration file.
func New(fileName string, noCache, debug bool, logger *log.Logger) (*Config, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.setLogger(logger, debug)
	if err = cfg.setProxy(); err != nil {
		return nil, err
	}

	if noCache {
		return cfg, nil // don't read cache, but write after data load
	}

	token, err := readCachedToken(cfg.AuthCache)
	if err != nil {
		return nil, fmt.Errorf("read cached token: %w", err)
	}

	cfg.Translation.IAMToken = token
	return cfg, nil
}

func (c *Config) setLogger(logger *log.Logger, debug bool) {
	if debug || c.Debug {
		logger.SetOutput(os.Stdout)
	}
	c.Logger = logger
}

// setProxy sets HTTP proxy or uses environment variables.
func (c *Config) setProxy() error {
	if c.ProxyURL != "" {
		u, err := url.Parse(c.ProxyURL)

		if err != nil {
			return fmt.Errorf("parse proxy url: %w", err)
		}

		// logging proxy without credentials
		c.Logger.Printf("proxy URL: %s://%s/%s", u.Scheme, u.Host, u.Path)

		c.Proxy = http.ProxyURL(u)
		return nil
	}

	c.Proxy = http.ProxyFromEnvironment
	return nil
}

// InitToken sets IAM token if it's empty.
func (c *Config) InitToken(ctx context.Context, client *http.Client) error {
	c.Lock()
	defer c.Unlock()

	if c.Translation.IAMToken != "" {
		return nil
	}

	err := c.Translation.SetIAMToken(ctx, client, c.UserAgent, c.Logger, c.GetURL(cloud.TokenURL))
	if err != nil {
		return fmt.Errorf("set iam token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(cloud.TTL)
	return writeCachedToken(c.AuthCache, c.Translation.IAMToken, expiresAt)
}

// GetURL returns URL from config or default value.
func (c *Config) GetURL(urlString string) string {
	if newURL := c.URL[urlString]; newURL != "" {
		return newURL
	}

	return urlString
}
