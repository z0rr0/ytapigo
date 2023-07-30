YtAPIGo
=======

[![GoDoc](https://godoc.org/github.com/z0rr0/ytapigo?status.svg)](https://godoc.org/github.com/z0rr0/ytapigo)
![Go](https://github.com/z0rr0/ytapigo/workflows/Go/badge.svg)
![Version](https://img.shields.io/github/tag/z0rr0/ytapigo.svg)
![License](https://img.shields.io/github/license/z0rr0/ytapigo.svg)

It is a program to translate and check spelling using the console,
it's based on [Yandex Translate API](https://cloud.yandex.com/en/docs/translate/).
By default, UTF-8 encoding is used.

A spell check is supported only for English, Russian and Ukrainian languages.

### Usage

Build binary file **yg**: 

```
make build

./yg -h
Usage of ./yg:
  -c string
        configuration file (default "<USER_CONFIG_DIR>/ytapigo/config.json")
  -d    debug mode
  -g string
        translation languages direction (empty - auto en/ru, ru/en, "auto" - detected lang to ru)
  -r    reset cache
  -t duration
        timeout for requests (default 5s)
  -v    print version
```


Usage:

```
./yg -g en-fr Hello dear fried!  
Spelling: 
        fried -> [friend friends fred]
Bonjour chère fried!


./yg lion
lion [ˈlaɪən] (noun)
        лев (noun)
        syn: львица (noun), львенок (noun)
        mean: lev, lioness, cub
lion [ˈlaɪən] (adjective)
        львиный (adjective)
        
./yg лев 
лев(noun)
        lion (noun)
        Lev (noun)
        syn: Leo (noun)
        mean: лео        
```

### API keys

API keys are required for using Yandex Translate API.

Users should get API keys before *ytapigo* using (see links below).
By default, configuration file will be searched in [user config directory](https://golang.org/pkg/os/#UserConfigDir).
(example [cfg.example.json](https://github.com/z0rr0/ytapigo/blob/master/cfg.example.json)).

```json
{
  "user_agent": "ytapigo/3.0",
  "proxy_url": "proxy like https://user:password@host:port",
  "dictionary": "API dictionary key",
  "auth_cache": "path to local token credentials JSON cache file, no cache if empty",
  "debug": true,
  "translation": {
    "folder_id": "API translation folder ID",
    "key_id": "API key ID",
    "service_account_id": "API service account ID",
    "key_file": "path to local auth PEM file"
  }
}
```

1. **translation** - documentation [Yandex Translate](https://cloud.yandex.com/en/docs/translate/)
2. **dictionary** - documentation [Yandex Dictionary](https://tech.yandex.com/dictionary/)

Also it uses [Yandex Speller](http://api.yandex.ru/speller/).

## License

This source code is governed by a BSD license 
that can be found in the [LICENSE](https://github.com/z0rr0/ytapigo/blob/master/LICENSE) file.
