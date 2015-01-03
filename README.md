YtapiGo
=======

It is a program to translate and check spelling using the console, it based on [Yandex Translate API](http://api.yandex.ru/translate/). By default **en-ru** direction and UTF-8 encoding are used.

It's a clone of the project [Ytapi](http://z0rr0.github.io/ytapi/) but on the [Go programming language](http://golang.org/). This is created as a package/library, but it can be used as a separate program (see the example #5 below).

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
    spelling, translation, err = ytapigo.GetTr("en", "Hi All!")
    if err != nil {
        panic(err)
    }
    // Example #3: ru-en direction
    spelling, translation, err = ytapigo.GetTr("en", "Привет Всем!")
    if err != nil {
        panic(err)
    }
    // Example #4: translation article for a word
    spelling, translation, err = ytapigo.GetTr("car")
    if err != nil {
        panic(err)
    }
    // Example #5: read command line parameters
    spelling, translation, err = ytapigo.GetTr(os.Args[1:])
    if err != nil {
        panic(err)
    }
    fmt.Println(spelling)
    fmt.Println(translation)
}
```

Download binary file:

* Linux - [amd64](https://yadi.sk/d/mANlwqJDdmGDL), [386](https://yadi.sk/d/2Q_OsAtJdmFzL), [ARM (RaspberryPI)](https://yadi.sk/d/uIAc_mH0dmG2s)
* FreeBSD - [amd64](https://yadi.sk/d/GqdLOnP9dmG4f), [386](https://yadi.sk/d/1Mta4z7ldmG5J), [ARM (RaspberryPI)](https://yadi.sk/d/sp4e8YoHdmG6u)
* Darwin (MacOS) - [amd64](https://yadi.sk/d/ljuqozwtdmG7i), [386](https://yadi.sk/d/1su-PyKcdmGAo)
* Windows - [amd64](https://yadi.sk/d/cRqNHY-VdmGJK), [386](https://yadi.sk/d/49CcRmhMdmGHC)


### Dependencies

* [Go](http://golang.org/) is an open source programming language that makes it easy to build simple, reliable, and efficient software.
* [Viper](https://github.com/spf13/viper) is a complete configuration solution for go applications.

### API keys

You should get API KEYs before an using this program, them values have to wroten to a file **$HOME/.ytapigo.json** (see the example `ytapigo_example.json`):

```javascript
{
  "APItr": "some key value",
  "APIdict": "some key value"
}
```

1. **APItr** - API KEY for [Yandex Translate](http://api.yandex.ru/key/form.xml?service=trnsl)
2. **APIdict** - API KEY for [Yandex Dictionary](http://api.yandex.ru/key/form.xml?service=dict)

It was implemented using the services:

* [Yandex Dictionary](http://api.yandex.ru/dictionary/)
* [Yandex Translate](http://api.yandex.ru/translate/)
* [Yandex Speller](http://api.yandex.ru/speller/)