package ytapigo

import (
    "fmt"
    "testing"
)

func TestCheckYT(t *testing.T) {
    if TestMsg != CheckYT() {
        t.Errorf("Failed simple test")
    }
}

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

func TestReadConfig(t *testing.T) {
    _, err := ReadConfig()
    if err != nil {
        t.Errorf("Config file error")
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

func TestGetSourceLang(t *testing.T) {
    cfg, err := ReadConfig()
    if err != nil {
        t.Errorf("Config file error")
    }
    LoggerInit(&cfg)
    sources1 := map[string]string{"en-ru": "en", "ru-hu": "ru", "hu-zh": "hu"}
    for k, v := range sources1 {
        source, err := GetSourceLang(&cfg, k)
        if (err != nil) || (source != v) {
            t.Errorf("Wrong GetSourceLang")
        }
    }
    sources2 := [2]string{"", "-hu"}
    for _, v := range sources2 {
        _, err := GetSourceLang(&cfg, v)
        if err == nil {
            t.Errorf("Wrong GetSourceLang (bad attempts)")
        }
    }
}

func TestGetTr(t *testing.T) {
    var (
        examples_en = map[string][]string{"the lion": {"", "Лев"}, "the car": {"", "автомобиль"}}
        examples_ru = map[string][]string{"красная машина": {"", "red car"}, "большой дом": {"", "big house"}}
        params []string
    )

    params = make([]string, 1)
    for key, val := range examples_en {
        params[0] = key
        spelling, translation, err := GetTr(params)
        if err != nil {
            t.Errorf("Failed GetTr test:%v", err)
        }
        if (val[0] != spelling) || (val[1] != translation) {
            t.Errorf("Failed GetTr test")
        }
    }

    params = make([]string, 2)
    params[0] = "ru"
    for key, val := range examples_ru {
        params[1] = key
        spelling, translation, err := GetTr(params)
        if err != nil {
            t.Errorf("Failed GetTr test:%v", err)
        }
        if (val[0] != spelling) || (val[1] != translation) {
            t.Errorf("Failed GetTr test")
        }
    }
}
