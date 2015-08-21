// Copyright (c) 2014, Alexander Zaytsev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ytapigo implements console text translation
// method using Yandex web services.
//
package ytapigo

import (
    "encoding/json"
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
)

const (
    // ConfName is a name of configuration file
    ConfName string = ".ytapigo.json"
    traceMsg string = "%v [YtapiGo]: "
)

var (
    // ytJsonUrls is an array of API URLs:
    // 0-Spelling, 1-Translation, 2-Dictionary,
    // 3-Translation directions, 4-Dictionary directions
    ytJsonUrls = [5]string{
        "http://speller.yandex.net/services/spellservice.json/checkText?",
        "https://translate.yandex.net/api/v1.5/tr.json/translate?",
        "https://dictionary.yandex.net/api/v1/dicservice.json/lookup?",
        "https://translate.yandex.net/api/v1.5/tr.json/getLangs?",
        "https://dictionary.yandex.net/api/v1/dicservice.json/getLangs?",
    }
    // LdPattern is a regexp pattern to detect language direction.
    LdPattern = regexp.MustCompile(`^[a-z]{2}-[a-z]{2}$`)
)

// Translater is an interface to prepare JSON translation response.
type Translater interface {
    String() string
    Exists() bool
}

// LangChecker is an interface to check translation directions.
type LangChecker interface {
    String() string
    Contains(string) bool
    Description() string
}

// YtapiGo is a main structure
type YtapiGo struct {
    Cfg      *Config
    Debug    bool
    logError *log.Logger
    logDebug *log.Logger
}

// Config is current configuration info.
type Config struct {
    APItr   string              `json:"APItr"`
    APIdict string              `json:"APIdict"`
    Aliases map[string][]string `json:"Aliases"`
    Default string              `json:"Default"`
}

// LangsList is a  list of dictionary's languages (from JSON response).
// It is sorted in ascending order.
type LangsList []string

// LangsListTr is a list of translation's languages (from JSON response).
// "Dirs" field is an array that sorted in ascending order.
type LangsListTr struct {
    Dirs  []string          `json:"dirs"`
    Langs map[string]string `json:"langs"`
}

// JSONSpelResp is a type of a spell check (from JSON response).
// It supports "Translater" interface.
type JSONSpelResp struct {
    Word string   `json:"word"`
    S    []string `json:"s"`
    Code float64  `json:"code"`
    Pos  float64  `json:"pos"`
    Row  float64  `json:"row"`
    Col  float64  `json:"col"`
    Len  float64  `json:"len"`
}

// JSONSpelResps is an array of spelling results.
type JSONSpelResps []JSONSpelResp

// JSONTrResp is a type of a translation (from JSON response).
// It supports "Translater" interface.
type JSONTrResp struct {
    Code float64  `json:"code"`
    Lang string   `json:"lang"`
    Text []string `json:"text"`
}

// JSONTrDictExample is an internal type of JSONTrDict.
type JSONTrDictExample struct {
    Pos  string              `json:"pos"`
    Text string              `json:"text"`
    Tr   []map[string]string `json:"tr"`
}

// JSONTrDictItem is an internal type of JSONTrDict.
type JSONTrDictItem struct {
    Text string              `json:"text"`
    Pos  string              `json:"pos"`
    Syn  []map[string]string `json:"syn"`
    Mean []map[string]string `json:"mean"`
    Ex   []JSONTrDictExample `json:"ex"`
}

// JSONTrDictArticle is an internal type of JSONTrDict.
type JSONTrDictArticle struct {
    Pos  string           `json:"post"`
    Text string           `json:"text"`
    Ts   string           `json:"ts"`
    Gen  string           `json:"gen"`
    Tr   []JSONTrDictItem `json:"tr"`
}

// JSONTrDict is a type of a translation dictionary (from JSON response).
// It supports "Translater" interface.
type JSONTrDict struct {
    Head map[string]string   `json:"head"`
    Def  []JSONTrDictArticle `json:"def"`
}

// String is an implementation of String() method for JSONTrResp pointer.
func (jstr *JSONTrResp) String() string {
    if len(jstr.Text) == 0 {
        return ""
    }
    return jstr.Text[0]
}

// Exists is an implementation of Exists() method for JSONTrResp pointer.
func (jstr *JSONTrResp) Exists() bool {
    if jstr.String() != "" {
        return true
    }
    return false
}

// Exists is an implementation of Exists() method for JSONTrDict pointer.
func (jstr *JSONTrDict) Exists() bool {
    if jstr.String() != "" {
        return true
    }
    return false
}

// String is an implementation of String() method for JSONTrDict pointer.
// It returns a pretty formatted string.
func (jstr *JSONTrDict) String() string {
    var (
        result, arResult, syn, mean, ex, extr []string
        txtResult, txtSyn, txtMean, txtEx     string
    )
    result = make([]string, len(jstr.Def))
    for i, def := range jstr.Def {
        ts := ""
        if def.Ts != "" {
            ts = fmt.Sprintf(" [%v] ", def.Ts)
        }
        txtResult = fmt.Sprintf("%v%v(%v)", def.Text, ts, def.Pos)
        arResult = make([]string, len(def.Tr))
        for j, tr := range def.Tr {
            syn, mean, ex = make([]string, len(tr.Syn)), make([]string, len(tr.Mean)), make([]string, len(tr.Ex))
            txtSyn, txtMean, txtEx = "", "", ""
            for k, v := range tr.Syn {
                syn[k] = fmt.Sprintf("%v (%v)", v["text"], v["pos"])
            }
            for k, v := range tr.Mean {
                mean[k] = v["text"]
            }
            for k, v := range tr.Ex {
                extr = make([]string, len(v.Tr))
                for t, trv := range v.Tr {
                    extr[t] = trv["text"]
                }
                ex[k] = fmt.Sprintf("%v: %v", v.Text, strings.Join(extr, ", "))
            }
            if len(syn) > 0 {
                txtSyn = fmt.Sprintf("\n\tsyn: %v", strings.Join(syn, ", "))
            }
            if len(mean) > 0 {
                txtMean = fmt.Sprintf("\n\tmean: %v", strings.Join(mean, ", "))
            }
            if len(ex) > 0 {
                txtEx = fmt.Sprintf("\n\texamples: \n\t\t%v", strings.Join(ex, "\n\t\t"))
            }

            arResult[j] = fmt.Sprintf("\t%v (%v)%v%v%v", tr.Text, tr.Pos, txtSyn, txtMean, txtEx)
        }
        result[i] = fmt.Sprintf("%v\n%v", txtResult, strings.Join(arResult, "\n"))
    }
    return strings.Join(result, "\n")
}

// Exists is an implementation of Exists() method for JSONSpelResps pointer.
func (jspells *JSONSpelResps) Exists() bool {
    if len(*jspells) > 0 {
        return true
    }
    return false
}

// String is an implementation of String() method for JSONSpelResps pointer.
func (jspells *JSONSpelResps) String() string {
    spellstr := make([]string, len(*jspells))
    for i, v := range *jspells {
        if v.Exists() {
            spellstr[i] = v.String()
        }
    }
    return fmt.Sprintf("Spelling: \n\t%v", strings.Join(spellstr, "\n\t"))
}

// Exists is an implementation of Exists() method for JSONSpelResp pointer
// (Translater interface).
func (jspell *JSONSpelResp) Exists() bool {
    if (len(jspell.Word) > 0) || (len(jspell.S) > 0) {
        return true
    }
    return false
}

// String is an implementation of String() method for JSONSpelResp pointer.
func (jspell *JSONSpelResp) String() string {
    return fmt.Sprintf("%v -> %v", jspell.Word, jspell.S)
}

// StringBinarySearch searches a string in a string array that is sorted in ascending order.
// This function uses binary search method and returns an index of
// the found element or -1.
func StringBinarySearch(strs []string, s string, a int, b int) int {
    l := b - a
    if l < 1 {
        if strs[a] == s {
            return a
        }
        return -1
    }
    med := (b + a) / 2
    if s > strs[med] {
        a = med + 1
    } else {
        b = med
    }
    return StringBinarySearch(strs, s, a, b)
}

// readConfig reads YtapiGo configuration.
func readConfig(file string) (*Config, error) {
    if file == "" {
        file = filepath.Join(os.Getenv("HOME"), ConfName)
    }
    _, err := os.Stat(file)
    if err != nil {
        return nil, err
    }
    jsondata, err := ioutil.ReadFile(file)
    if err != nil {
        return nil, err
    }
    cfg := &Config{}
    err = json.Unmarshal(jsondata, &cfg)
    if err != nil {
        return nil, err
    }
    for key := range cfg.Aliases {
        sort.Strings(cfg.Aliases[key])
    }
    return cfg, nil
}

// New creates new YtapiGo structure
func New(filename string, debug bool) (*YtapiGo, error) {
    cfg, err := readConfig(filename)
    if err != nil {
        return nil, err
    }
    logErr := log.New(os.Stderr, fmt.Sprintf(traceMsg, "ERROR"), log.Ldate|log.Ltime|log.Lshortfile)
    logDebug := log.New(ioutil.Discard, fmt.Sprintf(traceMsg, "DEBUG"), log.Ldate|log.Lmicroseconds|log.Lshortfile)
    if debug {
        logDebug = log.New(os.Stdout, fmt.Sprintf(traceMsg, "DEBUG"), log.Ldate|log.Lmicroseconds|log.Lshortfile)
    }
    ytg := &YtapiGo{cfg, debug, logErr, logDebug}
    return ytg, nil
}

// Request is a common method to send POST request and get []byte response.
func (ytg *YtapiGo) request(url string, params *url.Values) ([]byte, error) {
    result := []byte("")
    tr := &http.Transport{
        Proxy:               http.ProxyFromEnvironment,
        TLSHandshakeTimeout: 10 * time.Second,
    }
    client := &http.Client{Transport: tr}
    resp, err := client.PostForm(url, *params)
    if err != nil {
        return result, fmt.Errorf("network connection problems")
    }
    defer resp.Body.Close()
    ytg.logDebug.Printf("%v: %v\n", resp.Request.Method, resp.Request.URL)
    if resp.StatusCode != 200 {
        return result, fmt.Errorf("wrong response code=%v", resp.StatusCode)
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return result, err
    }
    return body, nil
}

// getLangsList gets LangChecker interface as a result to check available languages.
func (ytg *YtapiGo) getLangsList(dict bool, c chan LangChecker) {
    var (
        result LangChecker
        params url.Values
        urlstr string
    )
    if dict {
        result = &LangsList{}
        urlstr, params = ytJsonUrls[4], url.Values{"key": {ytg.Cfg.APIdict}}
    } else {
        result = &LangsListTr{}
        urlstr, params = ytJsonUrls[3], url.Values{"key": {ytg.Cfg.APItr}, "ui": {"en"}}
    }
    body, err := ytg.request(urlstr, &params)
    if err != nil {
        ytg.logError.Println(err)
        c <- result
        return
    }
    if err := json.Unmarshal(body, result); err != nil {
        ytg.logError.Println(err)
        c <- result
        return
    }
    c <- result
}

// String is an implementation of String() method for LangsList pointer (LangChecker interface).
func (lch *LangsList) String() string {
    return fmt.Sprintf("%v", strings.Join(*lch, ", "))
}

// Contains is an implementation of Contains() method for LangsList pointer (LangChecker interface).
func (lch *LangsList) Contains(s string) bool {
    result := false
    if !sort.StringsAreSorted(*lch) {
        sort.Strings(*lch)
    }
    if i := StringBinarySearch(*lch, s, 0, len(*lch)-1); i >= 0 {
        result = true
    }
    return result
}

// Description is an implementation of Description() method for
// LangsList pointer (LangChecker interface).
func (lch *LangsList) Description() string {
    return fmt.Sprintf("Length=%v\n%v", len(*lch), lch.String())
}

// String is an implementation of String() method for LangsListTr
// pointer (LangChecker interface).
func (ltr *LangsListTr) String() string {
    return fmt.Sprintf("%v", strings.Join(ltr.Dirs, ", "))
}

// Contains is an implementation of Contains() method for LangsListTr
// pointer (LangChecker interface).
func (ltr *LangsListTr) Contains(s string) bool {
    result := false
    if !sort.StringsAreSorted(ltr.Dirs) {
        sort.Strings(ltr.Dirs)
    }
    if i := StringBinarySearch(ltr.Dirs, s, 0, len(ltr.Dirs)-1); i >= 0 {
        result = true
    }
    return result
}

// Description is an implementation of Description() method
// for LangsListTr pointer (LangChecker interface).
func (ltr *LangsListTr) Description() string {
    const n int = 3
    var (
        collen, counter int
    )
    counter = len(ltr.Langs)
    i, descstr := 0, make([]string, counter)
    for k, v := range ltr.Langs {
        if len(v) > 0 {
            descstr[i] = fmt.Sprintf("%v - %v", k, v)
            i++
        }
    }
    sort.Strings(descstr)

    if (counter % n) != 0 {
        collen = counter/n + 1
    } else {
        collen = counter / n
    }
    output := make([]string, collen)
    for j := 0; j < collen; j++ {
        switch {
        case j+2*collen < counter:
            output[j] = fmt.Sprintf("%-25v %-25v %-25v", descstr[j], descstr[j+collen], descstr[j+2*collen])
        case j+collen < counter:
            output[j] = fmt.Sprintf("%-25v %-25v", descstr[j], descstr[j+collen])
        default:
            output[j] = fmt.Sprintf("%-25v", descstr[j])
        }
    }
    return strings.Join(output, "\n")
}

// GetLangs returns a list of available languages for current configuration.
func (ytg *YtapiGo) GetLangs() (string, error) {
    langsDic, langsTr := make(chan LangChecker), make(chan LangChecker)
    go ytg.getLangsList(true, langsDic)
    go ytg.getLangsList(false, langsTr)
    lgch, lgct := <-langsDic, <-langsTr

    if (lgch.String() == "") && (lgct.String() == "") {
        return "", fmt.Errorf("cannot read languages descriptions")
    }
    return fmt.Sprintf("Dictionary languages:\n%v\nTranslation languages:\n%v\n%v", lgch.String(), lgct.String(), lgct.Description()), nil
}

// direction erifies translation direction,
// checks its support by dictionary and translate API.
func (ytg *YtapiGo) direction(direction string) (bool, bool) {
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
func (ytg *YtapiGo) aliasDirection(direction string, langs *string, isalias *bool) (bool, bool) {
    *langs, *isalias = ytg.Cfg.Default, false
    if direction == "" {
        return false, false
    }
    alias := direction
    for k, v := range ytg.Cfg.Aliases {
        if i := StringBinarySearch(v, alias, 0, len(v)-1); i >= 0 {
            alias = k
            break
        }
    }
    langsDic, langsTr := make(chan LangChecker), make(chan LangChecker)
    go ytg.getLangsList(true, langsDic)
    go ytg.getLangsList(false, langsTr)
    lchDic, lchTr := <-langsDic, <-langsTr

    if LdPattern.MatchString(alias) {
        ytg.logDebug.Printf("maybe it is a direction \"%v\"", alias)
        lchdOk, lchtrOk := lchDic.Contains(alias), lchTr.Contains(alias)
        if lchdOk || lchtrOk {
            *langs, *isalias = alias, true
            return lchdOk, lchtrOk
        }
    }
    ytg.logDebug.Printf("not found lang for alias \"%v\", default direction \"%v\" will be used.", alias, ytg.Cfg.Default)
    return lchDic.Contains(ytg.Cfg.Default), lchTr.Contains(ytg.Cfg.Default)
}

// getSourceLang returns source language from a string of translation direction.
func (ytg *YtapiGo) getSourceLang(direction string) (string, error) {
    langs := strings.SplitN(direction, "-", 2)
    if (len(langs) > 0) && (len(langs[0]) > 0) {
        return langs[0], nil
    }
    return "", fmt.Errorf("cannot detect translation direction")
}

// Spelling checks a spelling of income text message.
// It returns JSONSpelResps as Translater interface.
func (ytg *YtapiGo) Spelling(lang, txt string) (Translater, error) {
    result := &JSONSpelResps{}
    params := url.Values{
        "lang":    {lang},
        "text":    {txt},
        "format":  {"plain"},
        "options": {"518"}}
    body, err := ytg.request(ytJsonUrls[0], &params)
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
func (ytg *YtapiGo) Translation(lang, txt string, tr bool) (Translater, error) {
    var (
        result Translater
        trurl  string
        params url.Values
    )
    if wordCounter := len(strings.Split(txt, " ")); tr || (wordCounter > 1) {
        result = &JSONTrResp{}
        trurl, params = ytJsonUrls[1], url.Values{
            "lang":   {lang},
            "text":   {txt},
            "key":    {ytg.Cfg.APItr},
            "format": {"plain"}}
    } else {
        result = &JSONTrDict{}
        trurl, params = ytJsonUrls[2], url.Values{
            "lang": {lang},
            "text": {txt},
            "key":  {ytg.Cfg.APIdict}}
    }
    body, err := ytg.request(trurl, &params)
    if err != nil {
        return result, err
    }
    err = json.Unmarshal(body, result)
    if err != nil {
        return result, err
    }
    return result, nil
}

// GetTranslations is a main YtapiGo method to get translation and spelling results.
func (ytg *YtapiGo) GetTranslations(params []string) (string, string, error) {
    var (
        wg                    sync.WaitGroup
        alias, ddirOk, tdirOk bool
        langs, txt, source    string
        spellResult, trResult Translater
        spellErr, trErr       error
    )
    switch l := len(params); {
    case l < 1:
        return "", "", fmt.Errorf("too few parameters")
    case l == 1:
        langs = ytg.Cfg.Default
        ddirOk, tdirOk = ytg.direction(langs)
        if !ddirOk {
            return "", "", fmt.Errorf("cannot verify 'Default' translation direction")
        }
        alias, txt = false, params[0]
    default:
        ddirOk, tdirOk = ytg.aliasDirection(params[0], &langs, &alias)
        if (!ddirOk) && (!tdirOk) {
            return "", "", fmt.Errorf("cannot verify translation direction")
        }
        if alias {
            txt = strings.Join(params[1:], " ")
            if (len(strings.SplitN(txt, " ", 2)) == 1) && (!ddirOk) {
                return "", "", fmt.Errorf("cannot verify dictionary direction")
            }
        } else {
            txt = strings.Join(params, " ")
        }
    }
    ytg.logDebug.Printf("direction=%v, alias=%v (%v, %v)", langs, alias, ddirOk, tdirOk)
    if source, spellErr = ytg.getSourceLang(langs); spellErr == nil {
        switch source {
        case "ru", "en", "uk":
            wg.Add(1)
            go func(i *Translater, e *error, l string, t string) {
                defer wg.Done()
                *i, *e = ytg.Spelling(l, t)
            }(&spellResult, &spellErr, source, txt)
        default:
            spellResult = &JSONSpelResps{}
            ytg.logDebug.Printf("spelling is skipped [%v]\n", source)
        }
    }
    wg.Add(1)
    go func(i *Translater, e *error, l string, t string, tr bool) {
        defer wg.Done()
        *i, *e = ytg.Translation(l, t, tr)
    }(&trResult, &trErr, langs, txt, false)
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
