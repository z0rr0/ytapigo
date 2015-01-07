YtapiGo
=======

It is a program to translate and check spelling using the console, it based on [Yandex Translate API](http://api.yandex.ru/translate/). By default UTF-8 encoding are used.

It's a clone of the project [Ytapi](http://z0rr0.github.io/ytapi/) but on the [Go programming language](http://golang.org/). This is created as a package/library, but it can be used as a separate program (see main.go.example).

A spell check is supported only for English, Russian and Ukrainian languages.

### Usage

```go
package main

import (
    "github.com/z0rr0/ytapigo"
    "os"
)

func main() {
    var (
        spelling, translation string
        err error
    )

    // Example #1: en-ru direction (default)
    spelling, translation, err = ytapigo.GetTr("Hi All!")
    if err != nil {
        panic(err)
    }
    // Example #2: en-ru direction
    spelling, translation, err = ytapigo.GetTr("en-ru", "Hi All!")
    if err != nil {
        panic(err)
    }
    // Example #3: translation article for a word
    spelling, translation, err = ytapigo.GetTr("en-ru", "car")
    if err != nil {
        panic(err)
    }
    // Example #4: read command line parameters
    spelling, translation, err = ytapigo.GetTr(os.Args[1:])
    if err != nil {
        panic(err)
    }
    fmt.Println(spelling)
    fmt.Println(translation)
}
```

Download binary file:

* Linux - [amd64](https://yadi.sk/d/DkVXPeuIdpu8Z), [386](https://yadi.sk/d/VbPP1mgndpu7v), [ARM (RaspberryPI)](https://yadi.sk/d/raQBuVvmdpu9U)
* FreeBSD - [amd64](https://yadi.sk/d/1Rfh1rd5dpu5z), [386](https://yadi.sk/d/UbezQACmdpu4w), [ARM (RaspberryPI)](https://yadi.sk/d/3o-5wUVhdpu6q)
* Darwin (MacOS) - [amd64](https://yadi.sk/d/_dyoBofEdpu3x), [386](https://yadi.sk/d/5zNaMAwBdpu2R)
* Windows - [amd64](https://yadi.sk/d/lBBFPIBcdpuF8), [386](https://yadi.sk/d/HehMeTSvdpuCh)


### Dependencies

* [Go](http://golang.org/) is an open source programming language that makes it easy to build simple, reliable, and efficient software.
* [Viper](https://github.com/spf13/viper) is a complete configuration solution for go applications.

### API keys

You should get API KEYs before an using this program, them values have to be written to a file **$HOME/.ytapigo.json** (see the example `ytapigo_example.json`). **APIlangs** is a set of [available translate directions](https://tech.yandex.ru/translate/doc/dg/concepts/langs-docpage/), each one can have a list of possible user's aliases.

```javascript
{
  "APItr": "some key value",
  "APIdict": "some key value",
  "Aliases": {                      // User's languages aliases
    "en-ru": ["en", "англ"],
    "ru-en": ["ru", "ру"],
  },
  "Default": "en-ru"                // default translation direction
}
```

1. **APItr** - API KEY for [Yandex Translate](https://tech.yandex.ru/keys/get/?service=trnsl)
2. **APIdict** - API KEY for [Yandex Dictionary](https://tech.yandex.ru/keys/get/?service=dict)

It was implemented using the services:

* [Yandex Dictionary](http://api.yandex.com/dictionary/)
* [Yandex Translate](http://api.yandex.com/translate/)
* [Yandex Speller](http://api.yandex.ru/speller/)
