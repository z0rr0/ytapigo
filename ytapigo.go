// Copyright (c) 2019, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ytapigo implements console text translation
// method using Yandex web services.
//
package ytapigo

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/z0rr0/ytapigo/cloud"
)

const (
	// ConfName is a name of configuration file
	ConfName string = ".ytapigo.json"
	traceMsg string = "%v [Ytapi]: "

	cacheTrLanguages        = "ytapigo_langs.json"
	cacheDictLanguages      = "ytapigo_dict_langs.json"
	cacheAuth               = "ytapigo.cloud"
	userAgent               = "Ytapi/2.0"
	defaultTimeout     uint = 10

	// expirationAuth is cloud "iamToken" expiration period.
	expirationAuth = time.Duration(12 * time.Hour)
)

var (
	// ServiceURLs contains map of used API URLs
	ServiceURLs = map[string]string{
		"spelling":         "http://speller.yandex.net/services/spellservice.json/checkText",
		"translate":        "https://translate.api.cloud.yandex.net/translate/v2/translate",
		"dictionary":       "https://dictionary.yandex.net/api/v1/dicservice.json/lookup",
		"translate_langs":  "https://translate.api.cloud.yandex.net/translate/v2/languages",
		"dictionary_langs": "https://dictionary.yandex.net/api/v1/dicservice.json/getLangs",
		//"translate_token":  "https://iam.api.cloud.yandex.net/iam/v1/tokens",
	}
	// LangDirection is a regexp pattern to detect language direction.
	LangDirection = regexp.MustCompile(`^[a-z]{2}-[a-z]{2}$`)

	loggerError = log.New(os.Stderr, fmt.Sprintf(traceMsg, "ERROR"), log.Ldate|log.Ltime|log.Lshortfile)
	loggerDebug = log.New(ioutil.Discard, fmt.Sprintf(traceMsg, "DEBUG"), log.Ldate|log.Lmicroseconds|log.Lshortfile)
)

// Translator is an interface to prepare JSON translation response.
type Translator interface {
	String() string
	Exists() bool
}

// LangChecker is an interface to check translation directions.
type LangChecker interface {
	String() string
	Contains(string) bool
	Description() string
	Sort()
}

// Services is a struct of used services.
type Services struct {
	Translation cloud.Account `json:"translation"`
	Dictionary  string        `json:"dictionary"`
}

// Languages is languages configuration.
type Languages struct {
	Default string              `json:"default"`
	Aliases map[string][]string `json:"aliases"`
}

// Config is current configuration info.
type Config struct {
	S       Services  `json:"services"`
	L       Languages `json:"languages"`
	Timeout uint      `json:"timeout"`
	key     map[string]*rsa.PrivateKey
}

// Ytapi is a main structure
type Ytapi struct {
	Cfg     *Config
	timeout time.Duration
	client  *http.Client
	caches  map[string]string
}

// readConfig reads Ytapi configuration.
func readConfig(file string) (*Config, error) {
	if file == "" {
		file = filepath.Join(os.Getenv("HOME"), ConfName)
	}
	_, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	for key := range cfg.L.Aliases {
		sort.Strings(cfg.L.Aliases[key])
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	return cfg, nil
}

// cleanCache removes cache files.
func cleanCache(caches map[string]string) {
	for _, f := range caches {
		// ignore errors without debug
		if err := os.Remove(f); err != nil {
			loggerDebug.Println(err)
		}
	}
}

// New creates new Ytapi structure
func New(filename string, nocache, debug bool) (*Ytapi, error) {
	cfg, err := readConfig(filename)
	if err != nil {
		return nil, err
	}
	if debug {
		loggerDebug.SetOutput(os.Stdout)
	}
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	tmpDir := os.TempDir()
	// alias: file path
	caches := make(map[string]string, 3)
	if nocache {
		cleanCache(caches)
	} else {
		caches["translate_langs"] = filepath.Join(tmpDir, cacheTrLanguages)
		caches["dictionary_langs"] = filepath.Join(tmpDir, cacheDictLanguages)
		caches["cloud"] = filepath.Join(tmpDir, cacheAuth)
	}
	client := &http.Client{Transport: tr}
	timeout := time.Duration(cfg.Timeout) * time.Second
	err = cfg.S.Translation.SetIAMToken(caches["cloud"], client, userAgent, timeout, loggerDebug, loggerError)
	if err != nil {
		return nil, err
	}
	ytg := &Ytapi{
		Cfg:     cfg,
		timeout: timeout,
		client:  client,
		caches:  caches,
	}
	return ytg, nil
}

// Request is a common method to send POST Request and get []byte response.
func (ytg *Ytapi) Request(url string, params *url.Values) ([]byte, error) {
	var resp *http.Response
	start := time.Now()
	req, err := http.NewRequest("POST", url, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ctx, cancel := context.WithTimeout(context.Background(), ytg.timeout)
	defer cancel()
	req = req.WithContext(ctx)

	ec := make(chan error)
	go func() {
		resp, err = ytg.client.Do(req)
		ec <- err
		close(ec)
	}()
	select {
	case <-ctx.Done():
		<-ec // wait error "context deadline exceeded"
		return nil, fmt.Errorf("timed out (%v)", ytg.timeout)
	case err := <-ec:
		if err != nil {
			return nil, err
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			loggerError.Printf("close body: %v\n", err)
		}
	}()
	loggerDebug.Printf(
		"done %v-%v [%v]: %v\n",
		resp.Request.Method, resp.StatusCode, time.Now().Sub(start), resp.Request.URL,
	)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrong response code=%v", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// getLanguageList requests dictionary language list.
func (ytg *Ytapi) getDictLanguageList(cache, uri string, params *url.Values) (LangChecker, error) {
	var (
		body []byte
		err  error
	)
	lc := &DictionaryLanguages{}
	if cache != "" {
		// try read cache file
		body, err = ioutil.ReadFile(cache)
		if err == nil {
			err = json.Unmarshal(body, lc)
			if err == nil {
				return lc, nil
			}
			loggerError.Printf("failed json unmarshal [%v]: %v", cache, err)
			// cache error, do Request
		}
	}
	body, err = ytg.Request(uri, params)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, lc)
	if err != nil {
		return nil, err
	}
	lc.Sort()
	if cache != "" {
		go func() {
			// cache sorted struct
			body, err := json.Marshal(lc)
			if err !=  nil {
				loggerError.Printf("prepare cache: %v", err)
			}
			if err := ioutil.WriteFile(cache, body, 0600); err != nil {
				loggerError.Printf("save cache: %v", err)
			}
		}()
	}
	return lc, nil
}

// getLanguageList requests translation language list.
func (ytg *Ytapi) getTrLanguageList(lc LangChecker, cache, uri string) error {
	var (
		body []byte
		err  error
	)
	if cache != "" {
		// try read cache file
		body, err = ioutil.ReadFile(cache)
		if err == nil {
			err = json.Unmarshal(body, lc)
			if err == nil {
				return nil
			}
			loggerError.Printf("failed json unmarshal [%v]: %v", cache, err)
			// cache error, do Request
		}
	}
	requestData := strings.NewReader(fmt.Sprintf(`{"folder_id":"%s"}`, ytg.Cfg.S.Translation.FolderID))
	body, err = cloud.Request(ytg.client, requestData, ServiceURLs["translate_langs"],
		ytg.Cfg.S.Translation.IAMToken, userAgent, ytg.timeout, loggerDebug, loggerError)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, lc)
	if err != nil {
		return err
	}
	if cache != "" {
		go func() {
			if err := ioutil.WriteFile(cache, body, 0600); err != nil {
				loggerError.Printf("save cache: %v", err)
			}
		}()
	}
	return nil
}

// dictionaryLanguageList requests dictionary languages and sends it to channel c.
func (ytg *Ytapi) dictionaryLanguageList(c chan LangChecker) {
	lc := &DictionaryLanguages{}
	params := &url.Values{"key": {ytg.Cfg.S.Dictionary}}
	err := ytg.getDictLanguageList(lc, ytg.caches["dict"], ServiceURLs["dictionary_langs"], params)
	if err != nil {
		loggerError.Println(err)
	}
	c <- lc
}

// translationLanguageList requests translation languages and sends it to channel c.
func (ytg *Ytapi) translationLanguageList(c chan LangChecker) {
	lc := &TranslateLanguages{}
	err := ytg.getTrLanguageList(lc, ytg.caches["dict"], ServiceURLs["dictionary_langs"])
	if err != nil {
		loggerError.Println(err)
	}
	c <- lc
}

// GetLanguages returns a list of available languages for current configuration.
func (ytg *Ytapi) GetLanguages() (string, error) {
	c := make(chan LangChecker)
	go ytg.translationLanguageList(c)
	go ytg.dictionaryLanguageList(c)
	result := ""
	for i := 0; i < 2; i++ {
		v := <-c
		if v != nil {
			result += v.String() + "\n"
		}
	}
	close(c)
	if result == "" {
		return "", errors.New("cannot read languages descriptions")
	}
	return result, nil
}

// direction verifies translation direction,
// checks its support by dictionary and translate API.
func (ytg *Ytapi) direction(direction string) (bool, bool) {
	if direction == "" {
		return false, false
	}
	langsDic, langsTr := make(chan LangChecker), make(chan LangChecker)
	go ytg.getLangsList(true, langsDic)
	go ytg.getLangsList(false, langsTr)
	lchDic, lchTr := <-langsDic, <-langsTr
	return lchDic.Contains(direction), lchTr.Contains(direction)
}

// aliasDirection verifies translation direction,
// checks its support by dictionary and translate API, but additionally considers users' aliases.
func (ytg *Ytapi) aliasDirection(direction string, langs *string, isAlias *bool) (bool, bool) {
	*langs, *isAlias = ytg.Cfg.Default, false
	if direction == "" {
		return false, false
	}
	alias := direction
	for k, v := range ytg.Cfg.Aliases {
		if i := sort.SearchStrings(v, alias); i < len(v) && v[i] == alias {
			alias = k
			break
		}
	}
	langsDic, langsTr := make(chan LangChecker), make(chan LangChecker)
	go ytg.getLangsList(true, langsDic)
	go ytg.getLangsList(false, langsTr)
	lchDic, lchTr := <-langsDic, <-langsTr

	if LangDirection.MatchString(alias) {
		loggerDebug.Printf("maybe it is a direction \"%v\"", alias)
		lchdOk, lchtrOk := lchDic.Contains(alias), lchTr.Contains(alias)
		if lchdOk || lchtrOk {
			*langs, *isAlias = alias, true
			return lchdOk, lchtrOk
		}
	}
	loggerDebug.Printf("not found lang for alias \"%v\", default direction \"%v\" will be used.",
		alias, ytg.Cfg.Default)
	return lchDic.Contains(ytg.Cfg.Default), lchTr.Contains(ytg.Cfg.Default)
}

// getSourceLang returns source language from a string of translation direction.
func (ytg *Ytapi) getSourceLang(direction string) (string, error) {
	langs := strings.SplitN(direction, "-", 2)
	if (len(langs) > 0) && (len(langs[0]) > 0) {
		return langs[0], nil
	}
	return "", errors.New("cannot detect translation direction")
}

// Spelling checks a spelling of income text message.
// It returns SpellerResponse as Translator interface.
func (ytg *Ytapi) Spelling(lang, txt string) (Translator, error) {
	result := &SpellerResponse{}
	params := url.Values{
		"lang":    {lang},
		"text":    {txt},
		"format":  {"plain"},
		"options": {"518"}}
	body, err := ytg.Request(ServiceURLs["spelling"], &params)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(body, result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Translation translates the text message using full machine translation
// or a search of a dictionary article by a word.
func (ytg *Ytapi) Translation(lang, txt string, tr bool) (Translator, error) {
	var (
		result Translator
		trurl  string
		params url.Values
	)
	if wordCounter := len(strings.Split(txt, " ")); tr || (wordCounter > 1) {
		result = &TranslateResponse{}
		trurl, params = ytJSONUrls[1], url.Values{
			"lang":   {lang},
			"text":   {txt},
			"key":    {ytg.Cfg.APItr},
			"format": {"plain"}}
	} else {
		result = &DictionaryResponse{}
		trurl, params = ytJSONUrls[2], url.Values{
			"lang": {lang},
			"text": {txt},
			"key":  {ytg.Cfg.APIdict}}
	}
	body, err := ytg.Request(trurl, &params)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(body, result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetTranslations is a main Ytapi method to get spelling and translation results.
func (ytg *Ytapi) GetTranslations(params []string) (string, string, error) {
	var (
		wg                     sync.WaitGroup
		alias, ddirOk, tdirOk  bool
		languages, txt, source string
		spellResult, trResult  Translator
		spellErr, trErr        error
	)
	switch l := len(params); {
	case l < 1:
		return "", "", errors.New("too few parameters")
	case l == 1:
		languages = ytg.Cfg.Default
		ddirOk, tdirOk = ytg.direction(languages)
		if !ddirOk {
			return "", "", errors.New("cannot verify 'Default' translation direction")
		}
		alias, txt = false, params[0]
	default:
		ddirOk, tdirOk = ytg.aliasDirection(params[0], &languages, &alias)
		if (!ddirOk) && (!tdirOk) {
			return "", "", errors.New("cannot verify translation direction")
		}
		if alias {
			txt = strings.Join(params[1:], " ")
			if (len(strings.SplitN(txt, " ", 2)) == 1) && (!ddirOk) {
				return "", "", errors.New("cannot verify dictionary direction")
			}
		} else {
			txt = strings.Join(params, " ")
		}
	}
	loggerDebug.Printf("direction=%v, alias=%v (%v, %v)", languages, alias, ddirOk, tdirOk)
	if source, spellErr = ytg.getSourceLang(languages); spellErr == nil {
		switch source {
		// only 3 languages are supported for spelling
		case "ru", "en", "uk":
			wg.Add(1)
			go func(i *Translator, e *error, l string, t string) {
				defer wg.Done()
				*i, *e = ytg.Spelling(l, t)
			}(&spellResult, &spellErr, source, txt)
		default:
			spellResult = &SpellerResponse{}
			loggerDebug.Printf("spelling is skipped [%v]\n", source)
		}
	}
	wg.Add(1)
	go func(i *Translator, e *error, l string, t string, tr bool) {
		defer wg.Done()
		*i, *e = ytg.Translation(l, t, tr)
	}(&trResult, &trErr, languages, txt, false)
	wg.Wait()
	if spellErr != nil {
		return "", "", spellErr
	}
	if trErr != nil {
		return "", "", trErr
	}
	if spellResult.Exists() {
		return spellResult.String(), trResult.String(), nil
	}
	return "", trResult.String(), nil
}

// Duration prints a time duration by debug logger.
func (ytg *Ytapi) Duration(t time.Time) {
	diff := time.Now().Sub(t)
	loggerDebug.Printf("duration=%s\n", diff)
}
