// Copyright (c) 2014, Alexander Zaytsev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ytapigo implements console text translation
// method using Yandex web services.
//
package ytapigo

import (
    "os"
    "fmt"
    "log"
    "time"
    "sync"
    "sort"
    "regexp"
    "strings"
    "net/url"
    "net/http"
    "io/ioutil"
    "path/filepath"
    "encoding/json"
)

const (
    ConfName string = ".ytapigo.json"
)
var (
    // YtJsonURLs is an array of API URLs:
    // 0-Spelling, 1-Translation, 2-Dictionary,
    // 3-Translation directions, 4-Dictionary directions
    YtJsonURLs = [5]string{
        "http://speller.yandex.net/services/spellservice.json/checkText?",
        "https://translate.yandex.net/api/v1.5/tr.json/translate?",
        "https://dictionary.yandex.net/api/v1/dicservice.json/lookup?",
        "https://translate.yandex.net/api/v1.5/tr.json/getLangs?",
        "https://dictionary.yandex.net/api/v1/dicservice.json/getLangs?",
    }
    // LdPattern is a regexp pattern to detect language direction.
    LdPattern *regexp.Regexp = regexp.MustCompile(`^[a-z]{2}-[a-z]{2}$`)
    // LoggerError implements error logger.
    LoggerError = log.New(os.Stderr, "YtapiGo ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
    // LoggerDebug implements debug logger, it's disabled by default.
    LoggerDebug = log.New(ioutil.Discard, "YtapiGo DEBUG: ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
)

type YtapiGo struct {
    Cfg Config
    Debug bool
}

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

// Config is current configuration info.
type Config struct {
    APItr string                `json:"APItr"`
    APIdict string              `json:"APIdict"`
    Aliases map[string][]string `json:"Aliases"`
    Default string              `json:"Default"`
}

// String shows info that was read from the configuration file.
func (yt *YtapiGo) String() string {
    return fmt.Sprintf("\nConfig:\n APItr=%v\n APIdict=%v\n Default=%v\n Aliases=%v\n Debug=%v", yt.Cfg.APItr, yt.Cfg.APIdict, yt.Cfg.Default, yt.Cfg.Aliases, yt.Debug)
}

// LangsList is a  list of dictionary's languages (from JSON response).
// It is sorted in ascending order.
type LangsList []string

// LangsListTr is a list of translation's languages (from JSON response).
// "Dirs" field is an array that sorted in ascending order.
type LangsListTr struct {
    Dirs []string           `json:"dirs"`
    Langs map[string]string `json:"langs"`
}

// JsonSpelResp is a type of a spell check (from JSON response).
// It supports "Translater" interface.
type JsonSpelResp struct {
    Word string   `json:"word"`
    S []string    `json:"s"`
    Code float64  `json:"code"`
    Pos float64   `json:"pos"`
    Row float64   `json:"row"`
    Col float64   `json:"col"`
    Len float64   `json:"len"`
}

// JsonTrResp is a type of a translation (from JSON response).
// It supports "Translater" interface.
type JsonTrResp struct {
    Code float64  `json:"code"`
    Lang string   `json:"lang"`
    Text []string `json:"text"`
}

// JsonTrDictExample is an internal type of JsonTrDict.
type JsonTrDictExample struct {
    Pos string
    Text string
    Tr []map[string]string
}

// JsonTrDictItem is an internal type of JsonTrDict.
type JsonTrDictItem struct {
    Text string
    Pos string
    Syn []map[string]string
    Mean []map[string]string
    Ex []JsonTrDictExample
}
// JsonTrDictArticle is an internal type of JsonTrDict.
type JsonTrDictArticle struct {
    Pos string
    Text string
    Ts string
    Gen string
    Tr []JsonTrDictItem
}

// JsonTrDict is a type of a translation dictionary (from JSON response).
// It supports "Translater" interface.
type JsonTrDict struct {
    Head map[string]string   `json:"head"`
    Def []JsonTrDictArticle  `json:"def"`
}

// JsonSpelResps is an array of spelling results.
type JsonSpelResps []JsonSpelResp

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
func (lch *LangsListTr) String() string {
    return fmt.Sprintf("%v", strings.Join(lch.Dirs, ", "))
}

// The implementation of Contains() method for LangsListTr
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
    i, desc_str := 0, make([]string, counter)
    for k, v := range ltr.Langs {
        if len(v) > 0 {
            desc_str[i] = fmt.Sprintf("%v - %v", k, v)
            i++
        }
    }
    sort.Strings(desc_str)

    if (counter % n) != 0 {
        collen = counter / n + 1
    } else {
        collen = counter / n
    }
    output := make([]string, collen)
    for j := 0; j < collen; j++ {
        switch {
            case j+2*collen < counter:
                output[j] = fmt.Sprintf("%-25v %-25v %-25v", desc_str[j], desc_str[j+collen], desc_str[j+2*collen])
            case j+collen < counter:
                output[j] = fmt.Sprintf("%-25v %-25v", desc_str[j], desc_str[j+collen])
            default:
                output[j] = fmt.Sprintf("%-25v", desc_str[j])
        }
    }
    return strings.Join(output, "\n")
}

// Exists is an implementation of Exists() method for JsonSpelResp pointer
// (Translater interface).
func (jspell *JsonSpelResp) Exists() bool {
    if (len(jspell.Word) > 0) || (len(jspell.S) > 0) {
        return true
    }
    return false
}

// String is an implementation of String() method for JsonSpelResp pointer.
func (jspell *JsonSpelResp) String() string {
    return fmt.Sprintf("%v -> %v", jspell.Word, jspell.S)
}

// Exists is an implementation of Exists() method for JsonSpelResps pointer.
func (jspells *JsonSpelResps) Exists() bool {
    if len(*jspells) > 0 {
        return true
    }
    return false
}

// String is an implementation of String() method for JsonSpelResps pointer.
func (jspells *JsonSpelResps) String() string {
    spellstr := make([]string, len(*jspells))
    for i, v := range *jspells {
        if v.Exists() {
            spellstr[i] = v.String()
        }
    }
    return fmt.Sprintf("Spelling: \n\t%v", strings.Join(spellstr, "\n\t"))
}

// String is an implementation of String() method for JsonTrResp pointer.
func (jstr *JsonTrResp) String() string {
    if len(jstr.Text) == 0 {
        return ""
    }
    return jstr.Text[0]
}

// Exists is an implementation of Exists() method for JsonTrResp pointer.
func (jstr *JsonTrResp) Exists() bool {
    if jstr.String() != "" {
        return true
    }
    return false
}

// Exists is an implementation of Exists() method for JsonTrDict pointer.
func (jstr *JsonTrDict) Exists() bool {
    if jstr.String() != "" {
        return true
    }
    return false
}

// String is an implementation of String() method for JsonTrDict pointer.
// It returns a pretty formatted string.
func (jstr *JsonTrDict) String() string {
    var (
        result, ar_result, syn, mean, ex, extr []string
        txt_result, txt_syn, txt_mean, txt_ex string
    )
    result = make([]string, len(jstr.Def))
    for i, def := range jstr.Def {
        ts := ""
        if def.Ts != "" {
            ts = fmt.Sprintf(" [%v] ", def.Ts)
        }
        txt_result = fmt.Sprintf("%v%v(%v)", def.Text, ts, def.Pos)
        ar_result = make([]string, len(def.Tr))
        for j, tr := range def.Tr {
            syn, mean, ex = make([]string, len(tr.Syn)), make([]string, len(tr.Mean)), make([]string, len(tr.Ex))
            txt_syn, txt_mean, txt_ex = "", "", ""
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
                txt_syn = fmt.Sprintf("\n\tsyn: %v", strings.Join(syn, ", "))
            }
            if len(mean) > 0 {
                txt_mean = fmt.Sprintf("\n\tmean: %v", strings.Join(mean, ", "))
            }
            if len(ex) > 0 {
                txt_ex = fmt.Sprintf("\n\texamples: \n\t\t%v", strings.Join(ex, "\n\t\t"))
            }

            ar_result[j] = fmt.Sprintf("\t%v (%v)%v%v%v", tr.Text, tr.Pos, txt_syn, txt_mean, txt_ex)
        }
        result[i] = fmt.Sprintf("%v\n%v", txt_result, strings.Join(ar_result, "\n"))
    }
    return strings.Join(result, "\n")
}

// New create new YtapiGo object.
func New() *YtapiGo {
    return &YtapiGo{}
}

// Read initializes YtapiGo configuration.
func (ytg *YtapiGo) Read() error {
    file := filepath.Join(os.Getenv("HOME"), ConfName)
    _, err := os.Stat(file);
    if err != nil {
        LoggerDebug.Println("can't find config file")
        return err
    }
    jsondata, err := ioutil.ReadFile(file)
    if err != nil {
        LoggerDebug.Println("can't read config file")
        return err
    }
    err = json.Unmarshal(jsondata, &ytg.Cfg)
    if err != nil {
        LoggerDebug.Println("json error")
        return err
    }
    for key := range ytg.Cfg.Aliases {
        sort.Strings(ytg.Cfg.Aliases[key])
    }
    return nil
}

// Spelling checks a spelling of income text message.
// It returns JsonSpelResps as Translater interface.
func (ytg *YtapiGo) Spelling(lang, txt string) (Translater, error) {
    result := &JsonSpelResps{}
    params := url.Values{
        "lang": {lang},
        "text": {txt},
        "format": {"plain"},
        "options": {"518"}}
    body, err := Request(YtJsonURLs[0], &params)
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
    ytconf := ytg.Cfg
    var (
        result Translater
        trurl string
        params url.Values
    )
    if word_counter := len(strings.Split(txt, " ")); tr || (word_counter > 1) {
        result = &JsonTrResp{}
        trurl, params = YtJsonURLs[1], url.Values{
            "lang": {lang},
            "text": {txt},
            "key": {ytconf.APItr},
            "format": {"plain"}}
    } else {
        result = &JsonTrDict{}
        trurl, params = YtJsonURLs[2], url.Values{
            "lang": {lang},
            "text": {txt},
            "key": {ytconf.APIdict}}
    }
    body, err := Request(trurl, &params)
    if err != nil {
        return result, err
    }
    err = json.Unmarshal(body, result)
    if err != nil {
        return result, err
    }
    return result, nil
}

// GetLangsList gets LangChecker interface as a result to check available languages.
func (ytg *YtapiGo) GetLangsList(dict bool, c chan LangChecker) {
    ytconf := ytg.Cfg
    var (
        result LangChecker
        params url.Values
        urlstr string
    )
    if dict {
        result = &LangsList{}
        urlstr = YtJsonURLs[4]
        urlstr, params = YtJsonURLs[4], url.Values{"key": {ytconf.APIdict}}
    } else {
        result = &LangsListTr{}
        urlstr, params = YtJsonURLs[3], url.Values{"key": {ytconf.APItr}, "ui": {"en"}}
    }
    body, err := Request(urlstr, &params)
    if err != nil {
        LoggerError.Println(err)
        c <- result
        return
    }
    if err := json.Unmarshal(body, result); err != nil {
        LoggerError.Println(err)
        c <- result
        return
    }
    c <- result
}

// Direction erifies translation direction,
// checks its support by dictionary and translate API.
func (ytg *YtapiGo) Direction(direction string) (bool, bool) {
    if len(direction) == 0 {
        return false, false
    }
    langs_dic, langs_tr := make(chan LangChecker), make(chan LangChecker)
    go ytg.GetLangsList(true, langs_dic)
    go ytg.GetLangsList(false, langs_tr)
    lch_dic, lch_tr := <-langs_dic, <-langs_tr
    return lch_dic.Contains(direction), lch_tr.Contains(direction)
}

// AliasDirection verifies translation direction,
// checks its support by dictionary and translate API, but additionally considers users' aliases.
func (ytg *YtapiGo) AliasDirection(direction string, langs *string, isalias *bool) (bool, bool) {
    cfg := ytg.Cfg
    *langs, *isalias = cfg.Default, false
    if len(direction) == 0 {
        return false, false
    }
    alias := direction
    for k, v := range cfg.Aliases {
        if i := StringBinarySearch(v, alias, 0, len(v)-1); i >= 0 {
            alias = k
            break
        }
    }
    langs_dic, langs_tr := make(chan LangChecker), make(chan LangChecker)
    go ytg.GetLangsList(true, langs_dic)
    go ytg.GetLangsList(false, langs_tr)
    lch_dic, lch_tr := <-langs_dic, <-langs_tr

    if LdPattern.MatchString(alias) {
        LoggerDebug.Printf("maybe it is a direction \"%v\"", alias)
        lchd_ok, lchtr_ok := lch_dic.Contains(alias), lch_tr.Contains(alias)
        if lchd_ok || lchtr_ok {
            *langs, *isalias = alias, true
            return lchd_ok, lchtr_ok
        }
    }
    LoggerDebug.Printf("not found lang for alias \"%v\", default direction \"%v\" will be used.", alias, cfg.Default)
    return lch_dic.Contains(cfg.Default), lch_tr.Contains(cfg.Default)
}

// GetSourceLang returns source language from a string of translation direction.
func (ytg *YtapiGo) GetSourceLang(direction string) (string, error) {
    langs := strings.SplitN(direction, "-", 2)
    if (len(langs) > 0) && (len(langs[0]) > 0) {
        return langs[0], nil
    }
    return "", fmt.Errorf("cannot detect translation direction")
}

// Langs returns a list of available languages for current configuration.
func (ytg *YtapiGo) Langs() (string, error) {
    langs_dic, langs_tr := make(chan LangChecker), make(chan LangChecker)
    go ytg.GetLangsList(true, langs_dic)
    go ytg.GetLangsList(false, langs_tr)
    lgch, lgct := <-langs_dic, <-langs_tr

    if (lgch.String() == "") && (lgct.String() == "") {
        return "", fmt.Errorf("cannot read languages descriptions")
    }
    return fmt.Sprintf("Dictionary languages:\n%v\nTranslation languages:\n%v\n%v", lgch.String(), lgct.String(), lgct.Description()), nil
}

// Translations is a main YtapiGo method to get translation and spelling results.
func (ytg *YtapiGo) Translations(params []string) (string, string, error) {
    var (
        wg sync.WaitGroup
        alias, ddir_ok, tdir_ok bool
        langs, txt, source string
        spell_result, tr_result Translater
        spell_err, tr_err error
    )
    switch l := len(params); {
        case l < 1:
            return "", "", fmt.Errorf("too few parameters")
        case l == 1:
            langs = ytg.Cfg.Default
            ddir_ok, tdir_ok = ytg.Direction(langs)
            if !ddir_ok {
                LoggerDebug.Println("unknown translation direction")
                return "", "", fmt.Errorf("cannot verify 'Default' translation direction")
            }
            alias, txt = false, params[0]
        default:
            ddir_ok, tdir_ok = ytg.AliasDirection(params[0], &langs, &alias)
            if (!ddir_ok) && (!tdir_ok) {
                LoggerDebug.Println("unknown translation direction")
                return "", "", fmt.Errorf("cannot verify translation direction")
            }
            if alias {
                txt = strings.Join(params[1:], " ")
                if (len(strings.SplitN(txt, " ", 2)) == 1) && (!ddir_ok) {
                    return "", "", fmt.Errorf("cannot verify dictionary direction")
                }
            } else {
                txt = strings.Join(params, " ")
            }
    }
    LoggerDebug.Printf("direction=%v, alias=%v (%v, %v)", langs, alias, ddir_ok, tdir_ok)
    if source, spell_err = ytg.GetSourceLang(langs); spell_err == nil {
        switch source {
            case "ru", "en", "uk":
                wg.Add(1)
                go func(i *Translater, e *error, l string, t string) {
                    defer wg.Done()
                    *i, *e = ytg.Spelling(l, t)
                }(&spell_result, &spell_err, source, txt)
            default:
                spell_result = &JsonSpelResps{}
                LoggerDebug.Printf("spelling is skipped [%v]\n", source)
        }
    }
    wg.Add(1)
    go func(i *Translater, e *error, l string, t string, tr bool) {
        defer wg.Done()
        *i, *e = ytg.Translation(l, t, tr)
    }(&tr_result, &tr_err, langs, txt, false)
    wg.Wait()
    if spell_err != nil {
        return "", "", spell_err
    }
    if tr_err != nil {
        return "", "", tr_err
    }
    if spell_result.Exists() {
        return spell_result.String(), tr_result.String(), nil
    }
    return "", tr_result.String(), nil
}

// GetTranslations is a main YtapiGo function to get translation results.
func GetTranslations(params []string) (string, string, error) {
    ytg := New()
    err := ytg.Read()
    if err != nil {
        return "", "", err
    }
    return ytg.Translations(params)
}

// GetLangs returns a list of available languages.
func GetLangs() (string, error) {
    var result string
    ytg := New()
    err := ytg.Read()
    if err != nil {
        return result, err
    }
    return ytg.Langs()
}

// DebugMode is a initialization of Logger handlers.
func DebugMode(debugmode bool) {
    debugHandle := ioutil.Discard
    if debugmode {
        debugHandle = os.Stdout
    }
    LoggerDebug = log.New(debugHandle, "YtapiGo DEBUG: ",
        log.Ldate|log.Lmicroseconds|log.Lshortfile)
}

// Request is a common method to send POST request and get []byte response.
func Request(url string, params *url.Values) ([]byte, error) {
    result := []byte("")
    tr := &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        TLSHandshakeTimeout: 10 * time.Second,
    }
    client := &http.Client{Transport: tr}
    resp, err := client.PostForm(url, *params)
    // resp, err := client.Get(url + params.Encode())
    if err != nil {
        LoggerDebug.Println(err)
        return result, fmt.Errorf("network connection problems")
    }
    defer resp.Body.Close()
    // LoggerDebug.Printf("%v: %v?%v\n", resp.Request.Method, resp.Request.URL, params.Encode())
    LoggerDebug.Printf("%v: %v\n", resp.Request.Method, resp.Request.URL)

    if resp.StatusCode != 200 {
        return result, fmt.Errorf("wrong response code=%v", resp.StatusCode)
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        LoggerError.Println(err)
        return result, err
    }
    return body, nil
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
