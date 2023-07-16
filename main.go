// Copyright (c) 2021, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package main is console text translation tool using Yandex web services.
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
	"time"

	"github.com/z0rr0/ytapigo/config"
	"github.com/z0rr0/ytapigo/handle"
)

// Name is a program name.
const Name = "YtAPI"

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
		debug      bool
		version    bool
		noCache    bool
		direction  string
		timeout    = 5 * time.Second
		configFile = filepath.Join(os.Getenv("HOME"), ".ytapigo3", "config.json")
		start      = time.Now()
	)

	defer func() {
		logger.Printf("duration=%v\n", time.Since(start).Truncate(time.Millisecond))

		if r := recover(); r != nil {
			if _, e := fmt.Fprintf(os.Stderr, "ERROR: %v\n", r); e != nil {
				panic(e)
			}
			os.Exit(1)
		}
	}()

	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.BoolVar(&version, "v", false, "print version")
	flag.StringVar(&configFile, "c", configFile, "configuration file")
	flag.BoolVar(&noCache, "r", false, "reset cache")
	flag.DurationVar(&timeout, "t", timeout, "timeout for request")
	flag.StringVar(
		&direction, "g", "",
		fmt.Sprintf("translation languages direction "+
			"(empty - auto en/ru, ru/en, %q - detected lang to ru)", handle.AutoLanguageDetect,
		),
	)

	flag.Parse()
	if version {
		fmt.Printf("%v: %v %v %v %v\n", Name, Version, Revision, GoVersion, BuildDate)
		flag.PrintDefaults()
		return
	}

	cfg, err := config.New(configFile, noCache, debug, logger)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	y := handle.New(cfg)
	if err = y.Run(ctx, direction, flag.Args()); err != nil {
		panic(err)
	}
}
