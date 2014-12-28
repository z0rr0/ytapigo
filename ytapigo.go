package ytapigo

import (
    "fmt"
    "log"
    "github.com/spf13/viper"
)
func CheckYT() string {
    return TestMsg
}

const (
    TestMsg string = "YtapiGo"
    ConfDir string = "$HOME"
    ConfName string = ".ytapigo"
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
    cfg, err := ReadConfig()
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(cfg)
}