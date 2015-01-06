package ytapigo

import (
    // "fmt"
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
