// Copyright (c) 2019, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ytapi implements console text translation
// method using Yandex web services.
package ytapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/z0rr0/ytapigo/ytapi/cloud"
)

const (
	// ConfName is a name of configuration file
	ConfName string = "cfg.json"
	traceMsg string = "%v [Ytapi]: "

	cacheTrLanguages        = "translate_languages.json"
	cacheDictLanguages      = "dictionary_languages.json"
	cacheAuth               = "cloud_auth.json"
	userAgent               = "Ytapi/2.0"
	defaultTimeout     uint = 10
)

var (
	// ServiceURLs contains map of used API URLs
	ServiceURLs = map[string]string{
		"spelling":         "http://speller.yandex.net/services/spellservice.json/checkText",
		"translate":        "https://translate.api.cloud.yandex.net/translate/v2/translate",
		"dictionary":       "https://dictionary.yandex.net/api/v1/dicservice.json/lookup",
		"translate_langs":  "https://translate.api.cloud.yandex.net/translate/v2/languages",
		"dictionary_langs": "https://dictionary.yandex.net/api/v1/dicservice.json/getLangs",
	}
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
}

// Ytapi is a main structure
type Ytapi struct {
	Cfg     *Config
	timeout time.Duration
	client  *http.Client
	caches  map[string]string
}

// Translation is common item for translation.
type Translation struct {
	Direction    string
	Text         string
	IsDictionary bool
	UseAlias     bool
}

func (t *Translation) String() string {
	return fmt.Sprintf("direction=%s, is_dictionary=%v, use_alias=%v", t.Direction, t.IsDictionary, t.UseAlias)
}

func (t *Translation) getAlias(lc LangChecker, value string, ytg *Ytapi) string {
	var direction string
	for alias, values := range ytg.Cfg.L.Aliases {
		if i := sort.SearchStrings(values, value); i < len(values) && values[i] == value {
			direction = alias
			break
		}
	}
	if direction == "" {
		return ""
	}
	if lc.Contains(direction) {
		return direction
	}
	return ""
}

func (t *Translation) getLanguages() (string, string, error) {
	languages := strings.SplitN(t.Direction, "-", 2)
	if len(languages) < 2 {
		return "", "", fmt.Errorf("cannot detect translation direction from: %v", t.Direction)
	}
	return languages[0], languages[1], nil
}

func (t *Translation) multiParse(ytg *Ytapi, params []string) error {
	// get some words, the 1st one can be a direction
	n := len(params)
	if n < 2 {
		return errors.New("failed usage multiParse method")
	}
	if n == 2 {
		// is it a dictionary request
		languages := &DictionaryLanguages{}
		err := ytg.getDictLanguageList(languages, ytg.caches["dictionary_langs"], ServiceURLs["dictionary_langs"])
		if err != nil {
			return err
		}
		direction := params[0]
		if languages.Contains(direction) {
			multiWords := strings.Split(params[1], " ")
			if len(multiWords) > 1 {
				t.Direction = direction
				t.Text = params[1]
				t.IsDictionary = false
				return nil
			}
			t.Direction = params[0]
			t.Text = params[1]
			t.IsDictionary = true
			return nil
		}
		direction = t.getAlias(languages, direction, ytg)
		if direction != "" {
			multiWords := strings.Split(params[1], " ")
			if len(multiWords) > 1 {
				t.Direction = direction
				t.Text = params[1]
				t.IsDictionary = false
				return nil
			}
			t.Direction = direction
			t.Text = params[1]
			t.IsDictionary = true
			t.UseAlias = true
			return nil
		}
	}
	languages := &TranslateLanguages{}
	err := ytg.getTrLanguageList(languages, ytg.caches["translate_langs"], ServiceURLs["dictionary_langs"])
	if err != nil {
		return err
	}
	direction := params[0]
	if languages.Contains(direction) {
		t.Direction = direction
		t.Text = strings.Join(params[1:], " ")
		t.IsDictionary = false
		return nil
	}
	direction = t.getAlias(languages, direction, ytg)
	if direction != "" {
		t.Direction = direction
		t.Text = strings.Join(params[1:], " ")
		t.IsDictionary = false
		t.UseAlias = true
		return nil
	}
	if languages.Contains(ytg.Cfg.L.Default) {
		t.Direction = ytg.Cfg.L.Default
		t.Text = strings.Join(params, " ")
		t.IsDictionary = false
		return nil
	}
	return fmt.Errorf("cannot verify 'Default' translation direction: %v", ytg.Cfg.L.Default)
}

func (t *Translation) parse(ytg *Ytapi, params []string) error {
	switch n := len(params); {
	case n < 1:
		return errors.New("too few parameters")
	case (n == 1) && (len(strings.SplitN(params[0], " ", 2)) == 1):
		languages := &DictionaryLanguages{}
		err := ytg.getDictLanguageList(languages, ytg.caches["dictionary_langs"], ServiceURLs["dictionary_langs"])
		if err != nil {
			return err
		}
		if !languages.Contains(ytg.Cfg.L.Default) {
			return fmt.Errorf("cannot verify 'Default' dictionary direction: %v", ytg.Cfg.L.Default)
		}
		t.Direction = ytg.Cfg.L.Default
		t.Text = params[0]
		t.IsDictionary = true
	case n == 1:
		// multi-word params[0]
		err := t.multiParse(ytg, strings.Split(params[0], " "))
		if err != nil {
			return err
		}
	default:
		err := t.multiParse(ytg, params)
		if err != nil {
			return err
		}
	}
	return nil
}

// readConfig reads Ytapi configuration.
func readConfig(file string) (*Config, error) {
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
func New(cfgDir string, nocache, debug bool) (*Ytapi, error) {
	fileName := filepath.Join(cfgDir, ConfName)
	cfg, err := readConfig(fileName)
	if err != nil {
		return nil, err
	}
	if debug {
		loggerDebug.SetOutput(os.Stdout)
	}
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	// alias: file path
	caches := make(map[string]string, 3)
	if nocache {
		cleanCache(caches)
	} else {
		caches["translate_langs"] = filepath.Join(cfgDir, cacheTrLanguages)
		caches["dictionary_langs"] = filepath.Join(cfgDir, cacheDictLanguages)
		caches["cloud"] = filepath.Join(cfgDir, cacheAuth)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrong response code=%v: %v", resp.StatusCode, string(body))
	}
	return body, nil
}

// getLanguageList requests dictionary language list.
func (ytg *Ytapi) getDictLanguageList(lc LangChecker, cache, uri string) error {
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
	params := &url.Values{"key": {ytg.Cfg.S.Dictionary}}
	body, err = ytg.Request(uri, params)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, lc)
	if err != nil {
		return err
	}
	lc.Sort()
	if cache != "" {
		// cache sorted struct
		body, err := json.Marshal(lc)
		if err != nil {
			loggerError.Printf("prepare cache: %v", err)
		}
		if err := ioutil.WriteFile(cache, body, 0600); err != nil {
			loggerError.Printf("save cache: %v", err)
		}
	}
	return nil
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
	lc.Sort()
	if cache != "" {
		// cache sorted struct
		body, err := json.Marshal(lc)
		if err != nil {
			loggerError.Printf("prepare cache: %v", err)
		}
		if err := ioutil.WriteFile(cache, body, 0600); err != nil {
			loggerError.Printf("save cache: %v", err)
		}
	}
	return nil
}

// dictionaryLanguageList requests dictionary languages and sends it to channel c.
func (ytg *Ytapi) dictionaryLanguageList(c chan LangChecker) {
	lc := &DictionaryLanguages{}
	err := ytg.getDictLanguageList(lc, ytg.caches["dictionary_langs"], ServiceURLs["dictionary_langs"])
	if err != nil {
		loggerError.Println(err)
	}
	c <- lc
}

// translationLanguageList requests translation languages and sends it to channel c.
func (ytg *Ytapi) translationLanguageList(c chan LangChecker) {
	lc := &TranslateLanguages{}
	err := ytg.getTrLanguageList(lc, ytg.caches["translate_langs"], ServiceURLs["translate_langs"])
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

// Spelling checks a spelling of income text message.
// It returns SpellerResponse as Translator interface.
func (ytg *Ytapi) Spelling(lang, txt string) (Translator, error) {
	result := &SpellerResponse{}
	params := url.Values{
		"lang":    {lang},
		"text":    {txt},
		"format":  {"plain"},
		"options": {"518"},
	}
	loggerDebug.Printf("spelling: %v", params.Encode())
	body, err := ytg.Request(ServiceURLs["spelling"], &params)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ytg *Ytapi) translationProcess(t *Translation, source, target string) (Translator, error) {
	const format = `{"folder_id":"%s","texts":["%s"],"targetLanguageCode":"%s","sourceLanguageCode":"%s"}`
	var result Translator

	data := fmt.Sprintf(format, ytg.Cfg.S.Translation.FolderID, t.Text, target, source)
	loggerDebug.Printf("translation process: %v", data)

	requestData := strings.NewReader(data)
	body, err := cloud.Request(ytg.client, requestData, ServiceURLs["translate"],
		ytg.Cfg.S.Translation.IAMToken, userAgent, ytg.timeout, loggerDebug, loggerError)
	if err != nil {
		return nil, err
	}
	result = &TranslateResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ytg *Ytapi) dictionaryProcess(t *Translation) (Translator, error) {
	var result Translator
	result = &DictionaryResponse{}
	params := url.Values{
		"lang": {t.Direction},
		"text": {t.Text},
		"key":  {ytg.Cfg.S.Dictionary},
	}
	loggerDebug.Printf("dictionary process: %v", params.Encode())
	body, err := ytg.Request(ServiceURLs["dictionary"], &params)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Translation translates the text message using full machine translation
// or a search of a dictionary article by a word.
func (ytg *Ytapi) Translation(t *Translation, source, target string) (Translator, error) {
	var (
		result Translator
		err    error
	)
	if t.IsDictionary {
		result, err = ytg.dictionaryProcess(t)
	} else {
		result, err = ytg.translationProcess(t, source, target)
	}
	return result, err
}

// GetTranslations returns spelling and translation results.
func (ytg *Ytapi) GetTranslations(params []string) (string, string, error) {
	var (
		wg              sync.WaitGroup
		result          Translator
		spellingResult  string
		translateResult string
	)
	t := &Translation{}
	err := t.parse(ytg, params)
	if err != nil {
		return "", "", err
	}
	loggerDebug.Println(t)
	source, target, err := t.getLanguages()
	if err != nil {
		return "", "", err
	}
	wg.Add(2) // spelling + translation
	go func() {
		switch source {
		// only 3 languages are supported for spelling
		case "ru", "en", "uk":
			s, err := ytg.Spelling(source, t.Text)
			if err != nil {
				loggerError.Printf("spelling error: %v", err)
			} else {
				spellingResult = s.String()
			}
		default:
			loggerDebug.Printf("spelling is skipped [%v]", source)
		}
		wg.Done()
	}()
	go func() {
		result, err = ytg.Translation(t, source, target)
		if err != nil {
			loggerError.Printf("translation error: %v", err)
		} else {
			translateResult = result.String()
		}
		wg.Done()
	}()
	wg.Wait()
	return spellingResult, translateResult, nil
}

// Duration prints a time duration by debug logger.
func (ytg *Ytapi) Duration(t time.Time) {
	diff := time.Now().Sub(t)
	loggerDebug.Printf("duration=%s\n", diff)
}
