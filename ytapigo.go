// Copyright (c) 2014, Alexander Zaytsev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ytapigo implements console text translation
// method using Yandex web services.
//
package ytapigo

import (
    "fmt"
    "log"
    "github.com/spf13/viper"
    "net/http"
    "net/url"
    "io/ioutil"
    "encoding/json"
    "os"
    "strings"
    "time"
    "sync"
    "sort"
    "regexp"
)

// test method
func CheckYT() string {
    return TestMsg
}

const (
    TestMsg string = "YtapiGo"
    ConfDir string = "$HOME"
    ConfName string = ".ytapigo"
)
var (
    // 0-Spelling, 1-Translation, 2-Dictionary,
    // 3-Translation directions, 4-Dictionary directions
    YtJsonURLs = [5]string{
        "http://speller.yandex.net/services/spellservice.json/checkText?",
        "https://translate.yandex.net/api/v1.5/tr.json/translate?",
        "https://dictionary.yandex.net/api/v1/dicservice.json/lookup?",
        "https://translate.yandex.net/api/v1.5/tr.json/getLangs?",
        "https://dictionary.yandex.net/api/v1/dicservice.json/getLangs?",
    }
    // language direction regexp pattern
    LdPattern *regexp.Regexp = regexp.MustCompile(`^[a-z]{2}-[a-z]{2}$`)
    // should be initiated in LoggerInit before any use.
    LoggerError *log.Logger
    LoggerDebug *log.Logger
)

// It's an interface to prepare JSON translation response.
type Translater interface {
    String() string
    Exists() bool
}

// It's an interface to check translation directions
type LangChecker interface {
    String() string
    Contains(string) bool
    Description() string
}

// Configuration data.
type YtConfig struct {
    APItr string
    APIdict string
    Aliases map[string]map[string]string
    Default string
    Debug bool
}

// It shows info that was read from the configuration file.
func (yt YtConfig) String() string {
    return fmt.Sprintf("\nConfig:\n APItr=%v\n APIdict=%v\n Default=%v\n Aliases=%v\n Debug=%v", yt.APItr, yt.APIdict, yt.Default, yt.Aliases, yt.Debug)
}

// A list of dictionary's languages (from JSON response).
// It is sorted in ascending order.
type LangsList []string

// A list of translation's languages (from JSON response).
// "Dirs" field is an array that sorted in ascending order.
type LangsListTr struct {
    Dirs []string           `json:"dirs"`
    Langs map[string]string `json:"langs"`
}

// A type of a spell check (from JSON response).
// It support "Translater" interface.
type JsonSpelResp struct {
    Word string   `json:"word"`
    S []string    `json:"s"`
    Code float64  `json:"code"`
    Pos float64   `json:"pos"`
    Row float64   `json:"row"`
    Col float64   `json:"col"`
    Len float64   `json:"len"`
}

// A type of a translation (from JSON response).
// It support "Translater" interface.
type JsonTrResp struct {
    Code float64  `json:"code"`
    Lang string   `json:"lang"`
    Text []string `json:"text"`
}

// Internal type of JsonTrDict
type JsonTrDictExample struct {
    Pos string
    Text string
    Tr []map[string]string
}
// Internal type of JsonTrDict
type JsonTrDictItem struct {
    Text string
    Pos string
    Syn []map[string]string
    Mean []map[string]string
    Ex []JsonTrDictExample
}
// Internal type of JsonTrDict
type JsonTrDictArticle struct {
    Pos string
    Text string
    Ts string
    Gen string
    Tr []JsonTrDictItem
}

// A type of a translation dictionary (from JSON response).
// It support "Translater" interface.
type JsonTrDict struct {
    Head map[string]string   `json:"head"`
    Def []JsonTrDictArticle  `json:"def"`
}

// An array of spelling results.
type JsonSpelResps []JsonSpelResp

// Initiation of Logger handlers
func LoggerInit(ytconfig *YtConfig) {
    errorHandle, debugHandle := os.Stdout, ioutil.Discard
    if ytconfig.Debug {
        debugHandle = os.Stdout
    }

    LoggerDebug = log.New(debugHandle,
        "DEBUG: ",
        log.Ldate|log.Lmicroseconds|log.Lshortfile)

    LoggerError = log.New(errorHandle,
        "ERROR: ",
        log.Ldate|log.Ltime|log.Lshortfile)
}

// The implementation of String() method for LangsList pointer (LangChecker interface).
func (lch *LangsList) String() string {
    return fmt.Sprintf("%v", strings.Join(*lch, ", "))
}
// The implementation of Contains() method for LangsList pointer (LangChecker interface).
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

// The implementation of Description() method for LangsList pointer (LangChecker interface).
func (lch *LangsList) Description() string {
    return fmt.Sprintf("Length=%v\n%v", len(*lch), lch.String())
}

// The implementation of String() method for LangsListTr pointer (LangChecker interface).
func (lch *LangsListTr) String() string {
    return fmt.Sprintf("%v", strings.Join(lch.Dirs, ", "))
}
// The implementation of Contains() method for LangsListTr pointer (LangChecker interface).
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

// The implementation of Description() method for LangsListTr pointer (LangChecker interface).
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

// The implementation of Exists() method for JsonSpelResp pointer
// (Translater interface).
func (jspell *JsonSpelResp) Exists() bool {
    if (len(jspell.Word) > 0) || (len(jspell.S) > 0) {
        return true
    }
    return false
}

// The implementation of String() method for JsonSpelResp pointer
// (Translater interface).
func (jspell *JsonSpelResp) String() string {
    return fmt.Sprintf("%v -> %v", jspell.Word, jspell.S)
}

// The implementation of Exists() method for JsonSpelResps pointer
// (Translater interface).
func (jspells *JsonSpelResps) Exists() bool {
    if len(*jspells) > 0 {
        return true
    }
    return false
}

// The implementation of String() method for JsonSpelResps pointer
// (Translater interface).
func (jspells *JsonSpelResps) String() string {
    spellstr := make([]string, len(*jspells))
    for i, v := range *jspells {
        if v.Exists() {
            spellstr[i] = v.String()
        }
    }
    return fmt.Sprintf("Spelling: \n\t%v", strings.Join(spellstr, "\n\t"))
}

// The implementation of String() method for JsonTrResp pointer
// (Translater interface).
func (jstr *JsonTrResp) String() string {
    if len(jstr.Text) == 0 {
        return ""
    }
    return jstr.Text[0]
}

// The implementation of Exists() method for JsonTrResp pointer
// (Translater interface).
func (jstr *JsonTrResp) Exists() bool {
    if jstr.String() != "" {
        return true
    }
    return false
}

// The implementation of Exists() method for JsonTrDict pointer
// (Translater interface).
func (jstr *JsonTrDict) Exists() bool {
    if jstr.String() != "" {
        return true
    }
    return false
}

// The implementation of String() method for JsonTrDict pointer
// (Translater interface). It returns a pretty formatted string.
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

// It reads the configuration file.
func ReadConfig() (YtConfig, error) {
    aliases := func(langs map[string]interface{}) map[string]map[string]string {
        var (
            it []interface{}
            s string
        )
        result := make(map[string]map[string]string, len(langs))
        for i, k := range langs {
            it = k.([]interface{})
            result[i] = make(map[string]string, len(it))
            for _, t := range it {
                s = t.(string)
                result[i][s] = s
            }
        }
        return result
    }
    viper.SetConfigName(ConfName)
    viper.SetConfigType("json")
    viper.AddConfigPath(ConfDir)
    viper.ReadInConfig()
    ytconfig := YtConfig{
        viper.GetString("APItr"),
        viper.GetString("APIdict"),
        aliases(viper.GetStringMap("Aliases")),
        viper.GetString("Default"),
        viper.GetBool("Debug")}
    if (ytconfig.APItr == "") || (ytconfig.APIdict == "") {
        return ytconfig, fmt.Errorf("Can not read API keys values from the config file: %v/%v.json", ConfDir, ConfName)
    }
    return ytconfig, nil
}

// Common method to send POST request and get []byte response.
func GetYtResponse(url string, params *url.Values) ([]byte, error) {
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
        return result, fmt.Errorf("ERROR: Network connection problems. Cannot send HTTP request.")
    }
    defer resp.Body.Close()
    LoggerDebug.Printf("%v: %v?%v\n", resp.Request.Method, resp.Request.URL, params.Encode())

    if resp.StatusCode != 200 {
        return result, fmt.Errorf("ERROR: Wrong response code=%v", resp.StatusCode)
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        LoggerError.Println(err)
        return result, err
    }
    // LoggerDebug.Println(string(body))
    return body, nil
}

// This method checks a spelling of income text message.
// It returns JsonSpelResps as Translater interface.
func GetSpelling(lang string, txt string) (Translater, error) {
    LoggerDebug.Println("Call GetSpelling")
    var result Translater
    jsr := JsonSpelResps{}
    result = &jsr
    params := url.Values{
        "lang": {lang},
        "text": {txt},
        "format": {"plain"},
        "options": {"518"}}
    body, err := GetYtResponse(YtJsonURLs[0], &params)
    if err != nil {
        return result, err
    }
    if err := json.Unmarshal(body, result); err != nil {
        LoggerError.Println(err)
        return result, err
    }
    return result, nil
}
// This method translates the text message using full machine translation
// or a search of a dictionary article by a word.
// It returns JsonTrResp or JsonTrDict as Translater interface.
func GetTranslation(lang string, txt string, ytconf *YtConfig, tr bool) (Translater, error) {
    LoggerDebug.Println("Call GetTranslation")
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

    body, err := GetYtResponse(trurl, &params)
    if err != nil {
        return result, err
    }
    if err := json.Unmarshal(body, result); err != nil {
        LoggerError.Println(err)
        return result, err
    }
    return result, nil
}

// It gets LangChecker interface as a result
// (using API-dictionary if dict=true, or API-translate)
// and writes it to a channel.
func GetLangsList(ytconf *YtConfig, dict bool, c chan LangChecker) {
    LoggerDebug.Println("Call GetLangsList", dict)
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
    body, err := GetYtResponse(urlstr, &params)
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

// It verifies translation direction, checks its support by
// dictionary and translate API.
func CheckDirection(cfg *YtConfig, direction string) (bool, bool) {
    if len(direction) == 0 {
        return false, false
    }
    langs_dic, langs_tr := make(chan LangChecker), make(chan LangChecker)
    go GetLangsList(cfg, true, langs_dic)
    go GetLangsList(cfg, false, langs_tr)
    lch_dic, lch_tr := <-langs_dic, <-langs_tr
    return lch_dic.Contains(direction), lch_tr.Contains(direction)
}

// It verifies translation direction, checks its support by
// dictionary and translate API, but additionally considers users' aliases.
func CheckAliasDirection(cfg *YtConfig, direction string, langs *string, isalias *bool) (bool, bool) {
    *langs, *isalias = cfg.Default, false
    if len(direction) == 0 {
        return false, false
    }
    alias := direction
    for k, v := range cfg.Aliases {
        if k == direction {
            alias = k
            break
        }
        _, ok := v[direction]
        if ok {
            alias = k
            break
        }
    }

    langs_dic, langs_tr := make(chan LangChecker), make(chan LangChecker)
    go GetLangsList(cfg, true, langs_dic)
    go GetLangsList(cfg, false, langs_tr)
    lch_dic, lch_tr := <-langs_dic, <-langs_tr

    if LdPattern.MatchString(alias) {
        LoggerDebug.Printf("Maybe it is a direction \"%v\"", alias)
        lchd_ok, lchtr_ok := lch_dic.Contains(alias), lch_tr.Contains(alias)
        if lchd_ok || lchtr_ok {
            *langs, *isalias = alias, true
            return lchd_ok, lchtr_ok
        }
    }
    LoggerDebug.Printf("Not found lang for alias \"%v\", default direction \"%v\" will be used.", alias, cfg.Default)
    return lch_dic.Contains(cfg.Default), lch_tr.Contains(cfg.Default)
}

// It returns source language from a string of translation direction.
func GetSourceLang(cfg *YtConfig, direction string) (string, error) {
    langs := strings.SplitN(direction, "-", 2)
    if (len(langs) > 0) && (len(langs[0]) > 0) {
        return langs[0], nil
    }
    return "", fmt.Errorf("Cannot detect translation direction. Please check the config file: %v/%v.json", ConfDir, ConfName)
}

// The main YtapiGo function to get translation results.
func GetTr(params []string) (string, string, error) {
    var (
        wg sync.WaitGroup
        spell_result, tr_result Translater
        spell_err, tr_err error
        langs, txt, source string
        lenparams int
        alias, ddir_ok, tdir_ok bool
    )
    cfg, err := ReadConfig()
    if err != nil {
        return "", "", err
    }
    LoggerInit(&cfg)

    lenparams = len(params)
    if lenparams < 1 {
        return "", "", fmt.Errorf("Too few parameters.")
    } else if lenparams == 1 {
        ddir_ok, tdir_ok = CheckDirection(&cfg, cfg.Default)
        if !ddir_ok {
            LoggerDebug.Println("Unknown translation direction")
            return "", "", fmt.Errorf("Cannot verify 'Default' translation direction. Please check the config file: %v/%v.json", ConfDir, ConfName)
        }
        langs, alias = cfg.Default, false
        txt = params[0]
    } else {
        ddir_ok, tdir_ok = CheckAliasDirection(&cfg, params[0], &langs, &alias)
        if (!ddir_ok) && (!tdir_ok) {
            LoggerDebug.Println("Unknown translation direction")
            return "", "", fmt.Errorf("Cannot verify translation direction. Please check the config file: %v/%v.json", ConfDir, ConfName)
        }
        if alias {
            if (lenparams == 2) && (!ddir_ok) {
                return "", "", fmt.Errorf("Cannot verify dictionary direction. Please check the config file: %v/%v.json", ConfDir, ConfName)
            }
            txt = strings.Join(params[1:], " ")
        } else {
            txt = strings.Join(params, " ")
        }
    }
    LoggerDebug.Printf("direction=%v, alias=%v (%v, %v)", langs, alias, ddir_ok, tdir_ok)
    source, spell_err = GetSourceLang(&cfg, langs)
    if source, spell_err = GetSourceLang(&cfg, langs); spell_err == nil {
        switch source {
            case "ru", "en", "uk":
                wg.Add(1)
                go func(i *Translater, e *error, l string, t string) {
                    defer wg.Done()
                    *i, *e = GetSpelling(l, t)
                }(&spell_result, &spell_err, source, txt)
            default:
                spell_result = &JsonSpelResps{}
                LoggerDebug.Printf("Spelling is skipped. Incorrect language - %v\n", source)

        }
    }
    wg.Add(1)
    go func(i *Translater, e *error, l string, t string, c *YtConfig, tr bool) {
        defer wg.Done()
        *i, *e = GetTranslation(l, t, c, tr)
    }(&tr_result, &tr_err, langs, txt, &cfg, false)
    wg.Wait()

    if spell_err != nil {
        LoggerDebug.Println("Spelling error.")
        return "", "", spell_err
    }
    if tr_err != nil {
        LoggerDebug.Println("Translation error.")
        return "", "", tr_err
    }
    if spell_result.Exists() {
        return spell_result.String(), tr_result.String(), nil
    }
    return "", tr_result.String(), nil
}

// It returns a help string.
func ShowHelp(name string) string {
    var program string
    if name == "" {
        program = "main"
    } else {
        program = name
    }
    return fmt.Sprintf("YtapiGo is a program to translate and check spelling using the console, it based on Yandex Translate API.\nParameters:\n\t-help\t- help message\n\t-langs\t- show available translation directions and languages\nExamples:\n\t# input explicit translation direction and text\n\t%v [translation direction] text\n\t# input only text message, a default translation direction from the cofig file will be used\n\t%v text", program, program)
}

// It returns a list of available languages
func GetLangs() (string, error) {
    cfg, err := ReadConfig()
    if err != nil {
        return "", err
    }
    LoggerInit(&cfg)

    langs_dic, langs_tr := make(chan LangChecker), make(chan LangChecker)
    go GetLangsList(&cfg, true, langs_dic)
    go GetLangsList(&cfg, false, langs_tr)
    lgch, lgct := <-langs_dic, <-langs_tr

    if (lgch.String() == "") && (lgct.String() == "") {
        return "", fmt.Errorf("Cannot read languages descriptions.")
    }
    return fmt.Sprintf("Dictionary languages:\n%v\nTranslation languages:\n%v\n%v", lgch.String(), lgct.String(), lgct.Description()), nil
}

// It searches a string in a string array that is sorted in ascending order.
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
