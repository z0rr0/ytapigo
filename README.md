YtapiGo
=======

It is a programm to translate and check spelling using the console, it based on [Yandex Translate API](http://api.yandex.ru/translate/). By default **en-ru** direction and UTF-8 encoding are used.

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

* [Linux x86_64](https://yadi.sk/d/GMRRkcMidjTDK)
* [RaspberryPI](https://yadi.sk/d/5Aq5XwcJdjRud)

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