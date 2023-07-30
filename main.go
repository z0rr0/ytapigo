package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/z0rr0/ytapigo/config"
	"github.com/z0rr0/ytapigo/handle"
)

// Name is a program name.
const Name = "YtAPIGo"

var (
	// Version is a version from GIT tags
	Version = "0.0.0"
	// Revision is GIT revision number
	Revision = "git:000000"
	// BuildDate is build date
	BuildDate = "2016-01-01_01:01:01UTC"
	// GoVersion is runtime Go language version
	GoVersion = runtime.Version()

	logger = log.New(io.Discard, "DEBUG: ", log.Lmicroseconds|log.Lshortfile)
)

func main() {
	var (
		debug     bool
		version   bool
		noCache   bool
		direction string
		timeout   = 5 * time.Second
		start     = time.Now()
	)

	defer func() {
		logger.Printf("total duration %v\n", time.Since(start).Truncate(time.Millisecond))

		if r := recover(); r != nil {
			if _, e := fmt.Fprintf(os.Stderr, "ERROR: %v\n", r); e != nil {
				panic(e)
			}
			os.Exit(1)
		}
	}()

	configDir, cacheDir, err := defaultDirectories()
	if err != nil {
		panic(err)
	}
	configFile := filepath.Join(configDir, "config.json")

	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.BoolVar(&version, "v", false, "print version")
	flag.StringVar(&configFile, "c", configFile, "configuration file")
	flag.BoolVar(&noCache, "r", false, "reset cache")
	flag.DurationVar(&timeout, "t", timeout, "timeout for requests")
	flag.StringVar(
		&direction, "g", "",
		fmt.Sprintf("translation direction "+
			"(empty - auto en/ru, ru/en, %q - auto-detected language to ru)", handle.AutoLanguageDetect,
		),
	)

	flag.Parse()
	if version {
		fmt.Printf("%v: %v %v %v %v\n", Name, Version, Revision, GoVersion, BuildDate)
		flag.PrintDefaults()
		return
	}

	cfg, err := config.New(configFile, configDir, cacheDir, noCache, debug, logger)
	if err != nil {
		panic(err)
	}

	cfg.Logger.Printf("configuration"+
		"\n\tCONFIG:\t%v\n\tKEY:\t%v\n\tCACHE:\t%v", configFile, cfg.Translation.KeyFile, cfg.AuthCache,
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	y := handle.New(cfg)
	if err = y.Run(ctx, direction, flag.Args()); err != nil {
		panic(err)
	}
}

// defaultDirectories returns default configuration and cache directories.
func defaultDirectories() (string, string, error) {
	var (
		configDir string
		cacheDir  string
		err       error
		appFolder = strings.ToLower(Name)
	)

	if configDir, err = os.UserConfigDir(); err != nil {
		return "", "", fmt.Errorf("user config dir: %v", err)
	}

	if cacheDir, err = os.UserCacheDir(); err != nil {
		return "", "", fmt.Errorf("user cache dir: %v", err)
	}

	return filepath.Join(configDir, appFolder), filepath.Join(cacheDir, appFolder), nil
}
