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
)
func CheckYT() string {
    return TestMsg
}

const (
    TestMsg string = "YtapiGo"
    ConfDir string = "$HOME"
    ConfName string = ".ytapigo"
    YSpellJSON string = "http://speller.yandex.net/services/spellservice.json/checkText?"
)
type ShowLog interface {
    PrintLog(msg ...string)
}
type YtConfig struct {
    APItr string
    APIdict string
    Debug bool
}
func (yt *YtConfig) PrintLog(msg ...string) {
    if yt.Debug {
        log.Println(msg)
    }
}
func (yt YtConfig) String() string {
    return fmt.Sprintf("Config:\n APItr=%v\n APIdict=%v\n Debug=%v", yt.APItr, yt.APIdict, yt.Debug)
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
func (jspresp *JsonSpelResp) Exists() bool {
    if len(jspresp.s) > 0 {
        return true
    }
    return false
}
func (jspresp *JsonSpelResp) Spell() string {
    return fmt.Sprintf("Spelling: %v -> %v", jspresp.word, jspresp.s)
}
func GetJsonSpelResp(m map[string]interface{}) JsonSpelResp {
    s_array_i := m["s"].([]interface{})
    s_array := make([]string, len(s_array_i))
    for i, v := range s_array_i {
        s_array[i] = v.(string)
    }
    return JsonSpelResp{m["word"].(string), s_array, m["code"].(float64), m["pos"].(float64), m["row"].(float64), m["col"].(float64), m["len"].(float64)}
}

// =============================================
// read configuration file
func ReadConfig() (YtConfig, error) {
    var showl ShowLog
    viper.SetConfigName(ConfName)
    viper.SetConfigType("json")
    viper.AddConfigPath(ConfDir)
    viper.ReadInConfig()
    ytconfig := YtConfig{viper.GetString("APItr"), viper.GetString("APIdict"), viper.GetBool("Debug")}
    showl = &ytconfig
    if (ytconfig.APItr == "") || (ytconfig.APIdict == "") {
        showl.PrintLog(ytconfig.String())
        return ytconfig, fmt.Errorf("Can not read API keys value from config file: %v/%v", ConfDir, ConfName)
    }
    return ytconfig, nil
}

func GetTr() {
    if len(os.Args) < 2 {
        log.Fatalln("Too few parameters")
    }
    // fmt.Println(os.Args[5:])
    // strings.Join(os.Args, " ")

    cfg, err := ReadConfig()
    if err != nil {
        log.Fatalln(err)
    }
    fmt.Println(cfg)



    params := url.Values{"lang": {"ru"}, "text": {"малоко"}, "format": {"plain"}, "options": {"518"}}
    resp, err := http.Get(YSpellJSON + params.Encode())
    if err != nil {
        log.Fatalln(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        log.Fatalln("Wrong response code=%v", resp.StatusCode)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    var jsr []map[string]interface{}
    if err := json.Unmarshal(body, &jsr); err != nil {
        log.Fatal(err)
    }
    for _, val := range jsr {
        obj := GetJsonSpelResp(val)
        if obj.Exists() {
            fmt.Println(obj.Spell())
        }
    }
}