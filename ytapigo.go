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
    // "sort"
)
func CheckYT() string {
    return TestMsg
}

const (
    TestMsg string = "YtapiGo"
    ConfDir string = "$HOME"
    ConfName string = ".ytapigo"
)
var (
    // 0-Spelling, 1-Translation, 2-Dict
    YtJsonURLs = [3]string{
        "http://speller.yandex.net/services/spellservice.json/checkText?",
        "https://translate.yandex.net/api/v1.5/tr.json/translate?",
        "https://dictionary.yandex.net/api/v1/dicservice.json/lookup?",
    }
    LoggerError *log.Logger
    LoggerDebug *log.Logger
)

// Initiate Logger handlers
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
// Interface to prepare JSON translation response
type Translater interface {
    String() string
    Exists() bool
}
// configuration data
type YtConfig struct {
    APItr string
    APIdict string
    APIlangs map[string][]string
    Default string
    Debug bool
}
func (yt YtConfig) String() string {
    return fmt.Sprintf("\nConfig:\n APItr=%v\n APIdict=%v\n Default=%v\n APIlangs=%v\n Debug=%v", yt.APItr, yt.APIdict, yt.Default, yt.APIlangs, yt.Debug)
}
type JsonSpelResp struct {
    Word string   `json:"word"`
    S []string    `json:"s"`
    Code float64  `json:"code"`
    Pos float64   `json:"pos"`
    Row float64   `json:"row"`
    Col float64   `json:"col"`
    Len float64   `json:"len"`
}
type JsonTrResp struct {
    Code float64  `json:"code"`
    Lang string   `json:"lang"`
    Text []string `json:"text"`
}

type JsonTrDictExample struct {
    Pos string
    Text string
    Tr []map[string]string
}
type JsonTrDictItem struct {
    Text string
    Pos string
    Syn []map[string]string
    Mean []map[string]string
    Ex []JsonTrDictExample
}
type JsonTrDictArticle struct {
    Pos string
    Text string
    Ts string
    Gen string
    Tr []JsonTrDictItem
}
type JsonTrDict struct {
    Head map[string]string   `json:"head"`
    Def []JsonTrDictArticle  `json:"def"`
}

type JsonSpelResps []JsonSpelResp

func (jspell *JsonSpelResp) Exists() bool {
    if (len(jspell.Word) > 0) || (len(jspell.S) > 0) {
        return true
    }
    return false
}
func (jspells *JsonSpelResps) Exists() bool {
    if len(*jspells) > 0 {
        return true
    }
    return false
}
func (jspell *JsonSpelResp) String() string {
    return fmt.Sprintf("%v -> %v", jspell.Word, jspell.S)
}
func (jstr *JsonTrResp) String() string {
    if len(jstr.Text) == 0 {
        return ""
    }
    return jstr.Text[0]
}
func (jstr *JsonTrResp) Exists() bool {
    if jstr.String() != "" {
        return true
    }
    return false
}
func (jspells *JsonSpelResps) String() string {
    spellstr := make([]string, len(*jspells))
    for i, v := range *jspells {
        if v.Exists() {
            spellstr[i] = v.String()
        }
    }
    return fmt.Sprintf("Spelling: \n\t%v", strings.Join(spellstr, "\n\t"))
}
func (jstr *JsonTrDict) Exists() bool {
    if jstr.String() != "" {
        return true
    }
    return false
}
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
// read configuration file
func ReadConfig() (YtConfig, error) {
    get_langs := func(langs map[string]interface{}) map[string][]string {
        var it []interface{}
        result := make(map[string][]string, len(langs))
        for i, k := range langs {
            it = k.([]interface{})
            result[i] = make([]string, len(it), len(it))
            for j, t := range it {
                result[i][j] = t.(string)
                // sorting
                // for p := j-1; (p >= 0) && (result[i][p] > result[i][p+1]); p-- {
                //     result[i][p], result[i][p+1] = result[i][p+1], result[i][p]
                // }
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
        get_langs(viper.GetStringMap("APIlangs")),
        viper.GetString("Default"),
        viper.GetBool("Debug")}
    if (ytconfig.APItr == "") || (ytconfig.APIdict == "") {
        return ytconfig, fmt.Errorf("Can not read API keys values from config file: %v/%v.json", ConfDir, ConfName)
    }
    // fmt.Println(ytconfig)
    return ytconfig, nil
}

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
    LoggerDebug.Printf("%v: %v\n", resp.Request.Method, resp.Request.URL)

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

// get spelling
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
// get Translation or a dictionary article
func GetTranslation(lang string, txt string, ytconf *YtConfig) (Translater, error) {
    LoggerDebug.Println("Call GetTranslation")
    var (
        result Translater
        trurl string
        params url.Values
    )
    if word_counter := len(strings.Split(txt, " ")); word_counter > 1 {
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

func GetSourceLang(cfg *YtConfig, direction string) (string, error) {
    langs := strings.SplitN(direction, "-", 2)
    if (len(langs) > 0) && (len(langs[0]) > 0) {
        return langs[0], nil
    }
    return "", fmt.Errorf("Cannot detect translation direction. Please check config file: %v/%v.json", ConfDir, ConfName)
}

// return translation direction
func CheckLang(cfg *YtConfig, lang string) (string, string, bool, error) {
    direction, found := "", false
    for i, lv := range cfg.APIlangs {
        if lang == i {
            direction, found = i, true
            break
        }
        for _, v := range lv {
            if lang == v {
                direction, found = i, true
                break
            }
        }
        if found {
            break
        }
    }
    if direction == "" {
        direction = cfg.Default
    }
    source, err := GetSourceLang(cfg, direction)
    return source, direction, found, err
}

// main YtapiGo function
func GetTr(params []string) (string, string, error) {
    var (
        lang, langs, txt string
        wg sync.WaitGroup
        spell_result, tr_result Translater
        spell_err, tr_err error
        lenparams int
        tr_found bool
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
        langs, tr_found = cfg.Default, true
        lang, tr_err = GetSourceLang(&cfg, langs)
        txt = params[0]
    } else {
        lang, langs, tr_found, tr_err = CheckLang(&cfg, params[0])
        if tr_found {
            txt = strings.Join(params[1:], " ")
        } else {
            txt = strings.Join(params, " ")
        }
    }
    if tr_err != nil {
        return "", "", tr_err
    }
    LoggerDebug.Printf("source=%v, direction=%v, found=%v", lang, langs, tr_found)

    switch lang {
        case "en", "ru", "uk":
            wg.Add(1)
            go func(i *Translater, e *error, l string, t string) {
                defer wg.Done()
                *i, *e = GetSpelling(l, t)
            }(&spell_result, &spell_err, lang, txt)
        default:
            spell_result = &JsonSpelResps{}
            LoggerDebug.Println("Spelling is skipped")
    }

    wg.Add(1)
    go func(i *Translater, e *error, l string, t string, c *YtConfig) {
        defer wg.Done()
        *i, *e = GetTranslation(l, t, c)
    }(&tr_result, &tr_err, langs, txt, &cfg)

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
