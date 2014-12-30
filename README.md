YtapiGo
=======

It is a programm to translate and check spelling using the console, it based on [Yandex Translate API](http://api.yandex.ru/translate/). By default **en-ru** direction and UTF-8 encoding are used.

### Usage

```go
package main

import (
    "github.com/z0rr0/ytapigo"
    // "os"
)

func main() {
    var params []string
    // first string parameter ignored

    // ru-en direction (default)
    params = []string{"", "en", "Hi All!"}
    // params = []string{"", "Hi All!"}
    ytapigo.GetTr(params)
    // Привет Всем!!!

    // en-ru direction
    params = []string{"", "ru", "Привет Всем!"}
    ytapigo.GetTr(params)
    // Hi All!

    // get translation article
    params = []string{"", "noun"}

    ytapigo.GetTr(params)
    // noun [naʊn] (noun)
    //     существительное (существительное)
    //     examples:
    //             collective noun: собирательное существительное
    // noun [naʊn] (adjective)
    //     именной (прилагательное)

    // use params from stdin
    // ytapigo.GetTr(os.Args)

}
```

Download binary file - [ytapigo](https://yadi.sk/d/ysOtugQVdiS6x)

### API keys

You should get API KEYs before an using this program, them values have to wroten to a file **$HOME/.ytapigo.json** (see the example `ytapigo_example.json`):

```javascript
{
  "APItr": "key value",
  "APIdict": "key value",
  "Debug": false
}
```

1. **APItr** - API KEY for [Yandex Translate](http://api.yandex.ru/key/form.xml?service=trnsl)
2. **APIdict** - API KEY for [Yandex Dictionary](http://api.yandex.ru/key/form.xml?service=dict)

It was implemented using the services:

* [Yandex Dictionary](http://api.yandex.ru/dictionary/)
* [Yandex Translate](http://api.yandex.ru/translate/)
* [Yandex Speller](http://api.yandex.ru/speller/)