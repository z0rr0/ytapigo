YtapiGo
=======

[![GoDoc](https://godoc.org/github.com/z0rr0/ytapigo?status.svg)](https://godoc.org/github.com/z0rr0/ytapigo) [![Build Status](https://travis-ci.org/z0rr0/ytapigo.svg?branch=master)](https://travis-ci.org/z0rr0/ytapigo)

It is a program to translate and check spelling using the console, it based on [Yandex Translate API](https://cloud.yandex.ru/docs/translate/). By default UTF-8 encoding is used.

It's a clone of the project [Ytapi](http://z0rr0.github.io/ytapi/) but on the [Go programming language](http://golang.org/). This is created as a package/library, but it can be used as a separate program (see main.go.example), the [documentation](http://godoc.org/github.com/z0rr0/ytapigo) contains details about all methods and variables.

A spell check is supported only for English, Russian and Ukrainian languages.

### Usage


See example in [main.go.example](https://github.com/z0rr0/ytapigo/blob/master/main.go.example).

The latest versions of binary files are available in [Releases](https://github.com/z0rr0/ytapigo/releases)

Usage:

```
chmod u+x ytapigo
./ytapigo en-fr Hello dear fried!
# output: Bonjour chers frit!

./ytapigo en-ru lion
lion [ˈlaɪən] ()
        лев (noun)
        syn: львица (noun), львенок (noun)
        mean: lev, lioness, cub
        examples:
                sea lion: морской лев
lion [ˈlaɪən] ()
        львиный (adjective)
        examples:
                lion's share: львиная доля
```

### Dependencies

Standard [Go library](http://golang.org/pkg/).

### API keys

Users should get API KEYs before an using this program, these values have to be written to a file **$HOME/.ytapigo.json** (see the example `ytapigo_example.json`). **APIlangs** is a set of [available translate directions](https://tech.yandex.ru/translate/doc/dg/concepts/langs-docpage/), each one can have a list of possible user's aliases.

```
{
  "services": {
    "translation": {
      "folder": "folder_id",
      "key_id": "key id",
      "service_account_id": "service account id",
      "key_file": "rsa private file"
    },
    "dictionary": "token"
  },
  "languages": {
    "default": "en-ru",
    "aliases": {
      "en-ru": ["enru", "en", "англ"],
      "ru-en": ["ruen", "ru", "ру"]
    }
  },
  "timeout": 10
}
```

1. **translation** - documentation [Yandex Translate](https://cloud.yandex.ru/docs/translate/)
2. **dictionary** - documentation [Yandex Dictionary](https://tech.yandex.com/dictionary/)

Also it uses [Yandex Speller](http://api.yandex.ru/speller/).

## License

This source code is governed by a BSD license that can be found in the [LICENSE](https://github.com/z0rr0/ytapigo/blob/master/LICENSE) file.
