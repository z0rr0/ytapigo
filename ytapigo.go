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
        "https://dictionary.yandex.net/api/v1/dicservice.json/lookup?"
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
// Interface to prepare JSON response
type Builder interface {
    Build(m map[string]interface{})
}
// configuration data
type YtConfig struct {
    APItr string
    APIdict string
    Debug bool
}
func (yt YtConfig) String() string {
    return fmt.Sprintf("\nConfig:\n APItr=%v\n APIdict=%v\n Debug=%v", yt.APItr, yt.APIdict, yt.Debug)
}
type JsonSpelResp struct {
    word string
    s []string
    code float64
    pos float64
    row float64
    col float64
    len float64
}
type JsonTrResp struct {
    code float64
    lang string
    text []string
}
type JsonSpelResps []JsonSpelResp

func (jspell *JsonSpelResp) Exists() bool {
    if len(jspell.s) > 0 {
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
    return fmt.Sprintf("%v -> %v", jspell.word, jspell.s)
}
func (jstr *JsonTrResp) String() string {
    if len(jstr.text) == 0 {
        return ""
    }
    return jstr.text[0]
}
func (jspells JsonSpelResps) String() string {
    spellstr := make([]string, len(jspells))
    for i, v := range jspells {
        if v.Exists() {
            spellstr[i] = v.String()
        }
    }
    return fmt.Sprintf("Spelling: \n\t%v\n", strings.Join(spellstr, "\n\t"))
}
func (jspell *JsonSpelResp) Build(m map[string]interface{}) {
    s_array := m["s"].([]interface{})
    jspell.s = make([]string, len(s_array))
    for i, v := range s_array {
        jspell.s[i] = v.(string)
    }
    jspell.word = m["word"].(string)
    jspell.code = m["code"].(float64)
    jspell.pos = m["pos"].(float64)
    jspell.row = m["row"].(float64)
    jspell.col = m["col"].(float64)
    jspell.len = m["len"].(float64)
}
func (jstr *JsonTrResp) Build(m map[string]interface{}) {
    s_array := m["text"].([]interface{})
    jstr.text = make([]string, len(s_array))
    for i, v := range s_array {
        jstr.text[i] = v.(string)
    }
    jstr.code = m["code"].(float64)
    jstr.lang = m["lang"].(string)
}

// read configuration file
func ReadConfig() (YtConfig, error) {
    viper.SetConfigName(ConfName)
    viper.SetConfigType("json")
    viper.AddConfigPath(ConfDir)
    viper.ReadInConfig()
    ytconfig := YtConfig{viper.GetString("APItr"), viper.GetString("APIdict"), viper.GetBool("Debug")}
    if (ytconfig.APItr == "") || (ytconfig.APIdict == "") {
        return ytconfig, fmt.Errorf("Can not read API keys value from config file: %v/%v", ConfDir, ConfName)
    }
    return ytconfig, nil
}

func GetYtResponse(url string, params *url.Values) ([]byte, error) {
    // TODO: http://golang.org/pkg/net/http/#Transport
    result := []byte("")
    tr := &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        TLSHandshakeTimeout: 10 * time.Second,
    }
    client := &http.Client{Transport: tr}
    resp, err := client.Get(url + params.Encode())
    if err != nil {
        LoggerError.Println(err)
        return result, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return result, fmt.Errorf("Wrong response code=%v", resp.StatusCode)
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
func GetSpelling(lang string, txt string) (JsonSpelResps, error) {
    LoggerDebug.Println("Call GetSpelling")
    result := make(JsonSpelResps, 0)
    params := url.Values{
        "lang": {lang},
        "text": {txt},
        "format": {"plain"},
        "options": {"518"}}
    body, err := GetYtResponse(YtJsonURLs[0], &params)
    if err != nil {
        return result, err
    }
    var jsr []map[string]interface{}
    if err := json.Unmarshal(body, &jsr); err != nil {
        LoggerError.Println(err)
        return result, err
    }
    result = make(JsonSpelResps, len(jsr))
    for i, val := range jsr {
        result[i].Build(val)
    }
    return result, nil
}
// get Translation
func GetTranslation(lang string, txt string, key string) (JsonTrResp, error) {
    LoggerDebug.Println("Call GetTranslation")
    result := JsonTrResp{}
    params := url.Values{
        "lang": {lang},
        "text": {txt},
        "key": {key},
        "format": {"plain"}}
    body, err := GetYtResponse(YtJsonURLs[1], &params)
    if err != nil {
        return result, err
    }
    var jsr map[string]interface{}
    if err := json.Unmarshal(body, &jsr); err != nil {
        LoggerError.Println(err)
        return result, err
    }
    result.Build(jsr)
    return result, nil
}

// main YtapiGo function
func GetTr() {
    var lang, langs, txt string
    lenparams := len(os.Args)
    if lenparams < 2 {
        log.Fatalln("Too few parameters")
    } else if lenparams == 2 {
        lang, langs, txt = "en", "en-ru", os.Args[1]
    } else {
        switch os.Args[1] {
            case "ru", "ру":
                lang, langs = "ru", "ru-en"
                txt = strings.Join(os.Args[2:], " ")
            case "en", "анг":
                lang, langs = "en", "en-ru"
                txt = strings.Join(os.Args[2:], " ")
            default:
                lang, langs = "en", "en-ru"
                txt = strings.Join(os.Args[1:], " ")
        }
    }
    cfg, err := ReadConfig()
    if err != nil {
        log.Fatalln(err)
    }
    LoggerInit(&cfg)
    LoggerDebug.Printf(cfg.String())
    // spellig
    spells, err := GetSpelling(lang, txt)
    if err != nil {
        LoggerError.Fatalln(err)
    }
    if spells.Exists() {
        fmt.Println(spells)
    }
    // translation
    trs, err := GetTranslation(langs, txt, cfg.APItr)
    if err != nil {
        LoggerError.Fatalln(err)
    }
    fmt.Println(trs.String())
}