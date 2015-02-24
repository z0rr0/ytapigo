package ytapigo

import (
    "fmt"
    "strings"
    "testing"
)

func ExampleStringBinarySearch() {
    var strs = []string{"aaa", "aab", "aac", "aad"}
    if i := StringBinarySearch(strs, "aac", 0, len(strs)-1); i < 0 {
        fmt.Println("Not found")
    } else {
        fmt.Printf("Found: strs[%v]=%v \n", i, strs[i])
    }
    // Output:
    // Found: strs[2]=aac
}

func TestDebugMode(t *testing.T) {
    if (LoggerError == nil) || (LoggerDebug == nil) {
        t.Errorf("incorrect references")
    }
    DebugMode(false)
    if (LoggerError.Prefix() != "YtapiGo ERROR: ") || (LoggerDebug.Prefix() != "YtapiGo DEBUG: ") {
        t.Errorf("incorrect loggers settings")
    }
    DebugMode(true)
    if (LoggerError.Flags() != 19) || (LoggerDebug.Flags() != 21) {
        t.Errorf("incorrect loggers settings")
    }
}

func TestStringBinarySearch(t *testing.T) {
    var strs = []string{"aaa", "aab", "aac", "aad"}
    if i := StringBinarySearch(strs, "aac", 0, len(strs)-1); i != 2 {
        t.Errorf("Incorrect BinarySearch result")
    }
    if i := StringBinarySearch(strs, "aae", 0, len(strs)-1); i != -1 {
        t.Errorf("Incorrect BinarySearch result")
    }
}

func TestGetLangs(t *testing.T) {
    DebugMode(true)
    // LangsList
    ll := &LangsList{"en-ru", "ru-en", "fr-en"}
    if ll.String() != "en-ru, ru-en, fr-en" {
        t.Errorf("incorrect result")
    }
    if len(ll.Description()) == 0 {
        t.Errorf("incorrect result")
    }
    if ll.Contains("ru-en") == false {
        t.Errorf("incorrect result")
    }
    if ll.Contains("jp-en") == true {
        t.Errorf("incorrect result")
    }
    // LangsListTr
    lltr := &LangsListTr{
        []string{"en-ru", "ru-en", "fr-en", "de-en"},
        map[string]string{"en": "English", "ru": "Russina", "fr": "French", "de": "German"},
    }
    if lltr.String() != "en-ru, ru-en, fr-en, de-en" {
        t.Errorf("incorrect result")
    }
    if len(lltr.Description()) == 0 {
        t.Errorf("incorrect result")
    }
    lltr = &LangsListTr{
        []string{"en-ru", "ru-en", "fr-en", "de-en"},
        map[string]string{"en": "English", "ru": "Russina", "fr": "French"},
    }
    if len(lltr.Description()) == 0 {
        t.Errorf("incorrect result")
    }
    if lltr.Contains("ru-en") == false {
        t.Errorf("incorrect result")
    }
    if lltr.Contains("jp-en") == true {
        t.Errorf("incorrect result")
    }
    ytg := New()
    if len(ytg.String()) == 0 {
        t.Errorf("failed string YtapiGo")
    }
    if err := ytg.Read(); err != nil {
        t.Errorf("can't read config file")
    }

    if langs, err := GetLangs(); err != nil {
        t.Errorf("GetLangs error")
    } else {
        if len(langs) == 0 {
            t.Errorf("empty langs")
        }
    }
    jsr1 := &JsonSpelResp{}
    if jsr1.Exists() == true {
        t.Errorf("incorrect result")
    }
    jsr2 := &JsonTrResp{}
    if jsr2.Exists() == true {
        t.Errorf("incorrect result")
    }
    jsr3 := &JsonTrDict{}
    if jsr3.Exists() == true {
        t.Errorf("incorrect result")
    }
}

func TestGetTranslations(t *testing.T) {
    DebugMode(true)
    var (
        examples_en = map[string][]string{"the lion": {"", "Лев"}, "the car": {"", "автомобиль"}}
        examples_ru = map[string][]string{"красная машина": {"", "red car"}, "большой дом": {"", "big house"}}
        example_dict = map[string]string{"car": "автомобиль", "house": "дом", "lion": "лев"}
        example_spell = map[string]string{"carr": "[car]", "housee": "[house]", "lionn": "[lion]"}
        example_aliases = map[string]bool{"enru": true, "er": false}
        // example_wrong = map[string]bool{"": false}
        params []string
    )

    params = make([]string, 1)
    for key, val := range examples_en {
        params[0] = key
        if spelling, translation, err := GetTranslations(params); err != nil {
            t.Errorf("failed GetTranslations test: %v", err)
        } else {
            if (val[0] != spelling) || (val[1] != translation) {
                t.Errorf("failed GetTranslations test result")
            }
        }
    }
    params = make([]string, 2)
    params[0] = "ru-en"
    for key, val := range examples_ru {
        params[1] = key
        if spelling, translation, err := GetTranslations(params); err != nil {
            t.Errorf("failed GetTranslations (ru) test: %v", err)
        } else {
            if (val[0] != spelling) || (val[1] != translation) {
                t.Errorf("failed GetTranslations (ru) test result")
            }
        }
    }
    params[0] = "en-ru"
    for key, val := range example_dict {
        params[1] = key
        if _, translation, err := GetTranslations(params); err != nil {
            t.Errorf("failed GetTranslations (dict) test: %v", err)
        } else {
            if !strings.Contains(translation, val) {
                t.Errorf("failed GetTranslations (dict) test result")
            }
        }
    }
    for key, val := range example_spell {
        params[1] = key
        if spelling, translation, err := GetTranslations(params); err != nil {
            t.Errorf("failed GetTranslations (dict) test: %v", err)
        } else {
            if !strings.Contains(spelling, val) || (len(translation) != 0) {
                t.Errorf("failed GetTranslations (dict) test result")
            }
        }
    }
    for key, val := range example_aliases {
        params[0], params[1] = key, "hi"
        if spelling, _, err := GetTranslations(params); err != nil {
            t.Errorf("failed GetTranslations (dict) test: %v", err)
        } else {
            if val && (len(spelling) != 0 ) {
                t.Errorf("incorrect alias result")
            }
        }
    }

}

// test request
// http://speller.yandex.net/services/spellservice.json/checkText?format=plain&lang=ru&options=518&text=приветт

// func TestGetSourceLang(t *testing.T) {
//     cfg, err := ReadConfig()
//     if err != nil {
//         t.Errorf("Config file error")
//     }
//     LoggerInit(&cfg)
//     sources1 := map[string]string{"en-ru": "en", "ru-hu": "ru", "hu-zh": "hu"}
//     for k, v := range sources1 {
//         source, err := GetSourceLang(&cfg, k)
//         if (err != nil) || (source != v) {
//             t.Errorf("Wrong GetSourceLang")
//         }
//     }
//     sources2 := [2]string{"", "-hu"}
//     for _, v := range sources2 {
//         _, err := GetSourceLang(&cfg, v)
//         if err == nil {
//             t.Errorf("Wrong GetSourceLang (bad attempts)")
//         }
//     }
// }

// func TestGetTr(t *testing.T) {
//     var (
//         examples_en = map[string][]string{"the lion": {"", "Лев"}, "the car": {"", "автомобиль"}}
//         examples_ru = map[string][]string{"красная машина": {"", "red car"}, "большой дом": {"", "big house"}}
//         params []string
//     )

//     params = make([]string, 1)
//     for key, val := range examples_en {
//         params[0] = key
//         spelling, translation, err := GetTr(params)
//         if err != nil {
//             t.Errorf("Failed GetTr test:%v", err)
//         }
//         if (val[0] != spelling) || (val[1] != translation) {
//             t.Errorf("Failed GetTr test")
//         }
//     }

//     params = make([]string, 2)
//     params[0] = "ru"
//     for key, val := range examples_ru {
//         params[1] = key
//         spelling, translation, err := GetTr(params)
//         if err != nil {
//             t.Errorf("Failed GetTr test:%v", err)
//         }
//         if (val[0] != spelling) || (val[1] != translation) {
//             t.Errorf("Failed GetTr test")
//         }
//     }
// }
