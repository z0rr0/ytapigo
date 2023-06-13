// Copyright (c) 2021, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package main is console text translation tool using Yandex web services.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/z0rr0/ytapigo/ytapi"
)

// Name is a program name.
const Name = "Ytapi"

var (
	// Version is a version from GIT tags
	Version = "0.0.0"
	// Revision is GIT revision number
	Revision = "git:000000"
	// BuildDate is build date
	BuildDate = "2016-01-01_01:01:01UTC"
	// GoVersion is runtime Go language version
	GoVersion = runtime.Version()
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ERROR: %v\n", r)
			os.Exit(1)
		}
	}()
	languages := flag.Bool("languages", false, "show available languages")
	debug := flag.Bool("debug", false, "debug mode")
	version := flag.Bool("version", false, "print version")
	nocache := flag.Bool("nocache", false, "reset cache")
	config := flag.String("config", "", "configuration directory, default $HOME/.ytapigo")
	flag.Parse()
	if *version {
		fmt.Printf("%v: %v %v %v %v\n", Name, Version, Revision, GoVersion, BuildDate)
		flag.PrintDefaults()
		return
	}
	configDir := *config
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".ytapigo")
	}
	ytg, err := ytapi.New(configDir, *nocache, *debug)
	if err != nil {
		panic(err)
	}
	t := time.Now()
	defer func() {
		ytg.Duration(t)
	}()
	if *languages {
		if langs, err := ytg.GetLanguages(); err != nil {
			panic(err)
		} else {
			fmt.Println(langs)
		}
	} else {
		if s, t, err := ytg.GetTranslations(flag.Args()); err != nil {
			panic(err)
		} else {
			fmt.Printf("%v\n%v\n", s, t)
		}
	}
}
