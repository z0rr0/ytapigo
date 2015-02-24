YtapiGo
=======

It is a program to translate and check spelling using the console, it based on [Yandex Translate API](http://api.yandex.com/translate/). By default UTF-8 encoding is used.

It's a clone of the project [Ytapi](http://z0rr0.github.io/ytapi/) but on the [Go programming language](http://golang.org/). This is created as a package/library, but it can be used as a separate program (see main.go.example), the [documentation](http://godoc.org/github.com/z0rr0/ytapigo) contains details about all methods and variables.

A spell check is supported only for English, Russian and Ukrainian languages.

### Usage

[![GoDoc](https://godoc.org/github.com/z0rr0/ytapigo?status.svg)](https://godoc.org/github.com/z0rr0/ytapigo) [![Build Status](https://travis-ci.org/z0rr0/ytapigo.svg?branch=master)](https://travis-ci.org/z0rr0/ytapigo)

See example in [main.go.example](https://github.com/z0rr0/ytapigo/main.go.example).


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
    spelling, translation, err = ytapigo.GetTranslations("Hi All!")
    if err != nil {
        panic(err)
    }
    // Example #2: en-ru direction
    spelling, translation, err = ytapigo.GetTranslations("en-ru", "Hi All!")
    if err != nil {
        panic(err)
    }
    // Example #3: translation article for a word
    spelling, translation, err = ytapigo.GetTranslations("en-ru", "car")
    if err != nil {
        panic(err)
    }
    // Example #4: read command line parameters
    spelling, translation, err = ytapigo.GetTranslations(os.Args[1:])
    if err != nil {
        panic(err)
    }
    fmt.Println(spelling)
    fmt.Println(translation)
}
```

Download binary file:

* Linux - [amd64](https://e.t34.me/open/YtapiGo/linux_amd64/main), [386](https://e.t34.me/open/YtapiGo/linux_386/main), [ARM (RaspberryPI)](https://e.t34.me/open/YtapiGo/linux_arm/main)
* FreeBSD - [amd64](https://e.t34.me/open/YtapiGo/freebsd_amd64/main), [386](https://e.t34.me/open/YtapiGo/freebsd_386/main), [ARM (RaspberryPI)](https://e.t34.me/open/YtapiGo/freebsd_arm/main)
* Darwin (MacOS) - [amd64](https://e.t34.me/open/YtapiGo/darwin_amd64/main), [386](https://e.t34.me/open/YtapiGo/darwin_386/main)
* Windows - [amd64](https://e.t34.me/open/YtapiGo/windows_amd64/main.exe), [386](https://e.t34.me/open/YtapiGo/windows_386/main.exe)

Usage:

```shell
chmod u+x main
./main en-fr Hello dear fried!
# output: Bonjour chers frit!
```

### Dependencies

Standard [Go library](http://golang.org/pkg/).

### API keys

Users should get API KEYs before an using this program, these values have to be written to a file **$HOME/.ytapigo.json** (see the example `ytapigo_example.json`). **APIlangs** is a set of [available translate directions](https://tech.yandex.ru/translate/doc/dg/concepts/langs-docpage/), each one can have a list of possible user's aliases.

```javascript
{
  "APItr": "some key value",
  "APIdict": "some key value",
  "Aliases": {                      // User's languages aliases
    "en-ru": ["en", "англ"],
    "ru-en": ["ru", "ру"]
  },
  "Default": "en-ru"                // default translation direction
}
```

1. **APItr** - API KEY for [Yandex Translate](http://api.yandex.com/key/form.xml?service=trnsl)
2. **APIdict** - API KEY for [Yandex Dictionary](http://api.yandex.com/key/form.xml?service=dict)

It was implemented using the services:

* [Yandex Dictionary](http://api.yandex.com/dictionary/)
* [Yandex Translate](http://api.yandex.com/translate/)
* [Yandex Speller](http://api.yandex.ru/speller/)
