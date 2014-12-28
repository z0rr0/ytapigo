YtapiGo
=======

It is a programm to translate and check spelling using the console, it based on [Yandex Translate API](http://api.yandex.ru/translate/). By default **en-ru** direction and UTF-8 encoding are used.

### API keys

You should get API KEYs before an using this program, them values have to wroten to a file **$HOME/.ytapigo.json** (see the example `ytapigo_example.json`):

```
{
  "apitr": "some key value",
  "apidict": "some key value"
}
```


1. **apitr** - API KEY for [Yandex Translate](http://api.yandex.ru/key/form.xml?service=trnsl)
2. **apidict** - API KEY for [Yandex Dictionary](http://api.yandex.ru/key/form.xml?service=dict)

It was implemented using the services:

* [Yandex Dictionary](http://api.yandex.ru/dictionary/)
* [Yandex Translate](http://api.yandex.ru/translate/)
* [Yandex Speller](http://api.yandex.ru/speller/)