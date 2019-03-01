package ytapi

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/z0rr0/ytapigo/ytapi/cloud"
)

const (
	keyFile    = "ytapigo_rsa_test.pem"
	tokenValue = "abc123"
)

type trRequestItem struct {
	params      []string
	success     bool
	spellResult string
	trResult    string
}

func getKey(name string, t *testing.T) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	f, err := os.Create(name)
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()
	privateKey := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.Encode(f, privateKey)
}

func TestNew(t *testing.T) {
	err := getKey(keyFile, t)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(keyFile); err != nil {
			t.Error(err)
		}
	}()
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{"iamToken":"%s","expiresAt":"2019-02-15T01:09:43.418711Z"}`, tokenValue)
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer authServer.Close()

	cloud.URL = authServer.URL // overwrite default auth URL
	y, err := New(".", true, true)
	if err != nil {
		t.Fatal(err)
	}
	if y.Cfg.S.Translation.IAMToken != tokenValue {
		t.Errorf("failed token: %v", y.Cfg.S.Translation.IAMToken)
	}
}

func TestYtapi_GetLanguages(t *testing.T) {
	err := getKey(keyFile, t)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(keyFile); err != nil {
			t.Error(err)
		}
	}()
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{"iamToken":"%s","expiresAt":"2019-02-15T01:09:43.418711Z"}`, tokenValue)
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer authServer.Close()
	cloud.URL = authServer.URL // overwrite default auth URL
	// translation languages server
	trLangServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"languages":[{"code":"ru","name":"Русский"}]}`
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer trLangServer.Close()
	ServiceURLs["translate_langs"] = trLangServer.URL
	// dictionary languages server
	dictLangServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `["ru-en"]`
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer dictLangServer.Close()
	ServiceURLs["dictionary_langs"] = dictLangServer.URL

	y, err := New(".", true, true)
	if err != nil {
		t.Fatal(err)
	}
	result, err := y.GetLanguages()
	if err != nil {
		t.Error(err)
	}
	eDict := "Dictionary languages:\nru-en"
	eTr := "Translation languages:\nru\nru - Русский             "
	expected := [2]string{
		fmt.Sprintf("%s\n%s\n", eDict, eTr),
		fmt.Sprintf("%s\n%s\n", eTr, eDict),
	}
	if (result != expected[0]) && (result != expected[1]) {
		t.Errorf("failed result: %q", result)
	}
}

func TestYtapi_GetTranslations(t *testing.T) {
	err := getKey(keyFile, t)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(keyFile); err != nil {
			t.Error(err)
		}
	}()
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{"iamToken":"%s","expiresAt":"2019-02-15T01:09:43.418711Z"}`, tokenValue)
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer authServer.Close()
	cloud.URL = authServer.URL // overwrite default auth URL
	// translation languages server
	trLangServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"languages":[{"code":"ru","name":"Русский"},{"code":"en","name":"English"}]}`
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer trLangServer.Close()
	ServiceURLs["translate_langs"] = trLangServer.URL
	// dictionary languages server
	dictLangServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `["en-ru"]`
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer dictLangServer.Close()
	ServiceURLs["dictionary_langs"] = dictLangServer.URL
	// speller server
	spellerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `[{"code":1,"pos":0,"row":0,"col":0,"len":5,"word": "timee","s":["time"]}]`
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer spellerServer.Close()
	ServiceURLs["spelling"] = spellerServer.URL
	// dictionary server
	dictServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `
{ "head": {},
  "def": [
     { "text": "time", "pos": "noun",
       "tr": [
          { "text": "время", "pos": "существительное",
            "syn": [{ "text": "раз" },{ "text": "тайм" }],
            "mean": [{ "text": "timing" },{ "text": "fold" },{ "text": "half"}],
            "ex" : [
               { "text": "prehistoric time",
                 "tr": [{"text": "доисторическое время"}]
               },
               { "text": "hundredth time",
                 "tr": [{"text": "сотый раз"}]
               },
               { "text": "time-slot",
                 "tr": [{ "text": "тайм-слот" }]
               }
            ]
          }
       ]
    }
  ]
}`
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer dictServer.Close()
	ServiceURLs["dictionary"] = dictServer.URL
	// translation server
	translationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{"translations":[{"text": "тест","detectedLanguageCode":"ru"}]}`
		if _, err := fmt.Fprintf(w, response); err != nil {
			t.Error(err)
		}
	}))
	defer translationServer.Close()
	ServiceURLs["translate"] = translationServer.URL

	y, err := New(".", true, true)
	if err != nil {
		t.Fatal(err)
	}
	values := []trRequestItem{
		{
			params:      []string{"time"},
			success:     true,
			spellResult: "Spelling: \n\ttimee -> [time]",
			trResult: "time()\n\tвремя (существительное)\n\tsyn: раз (), тайм ()\n\t" +
				"mean: timing, fold, half\n\texamples: \n\t\t" +
				"prehistoric time: доисторическое время\n\t\t" +
				"hundredth time: сотый раз\n\t\ttime-slot: тайм-слот",
		},
		{
			params:      []string{"en-ru time"},
			success:     true,
			spellResult: "Spelling: \n\ttimee -> [time]",
			trResult: "time()\n\tвремя (существительное)\n\tsyn: раз (), тайм ()\n\t" +
				"mean: timing, fold, half\n\texamples: \n\t\t" +
				"prehistoric time: доисторическое время\n\t\t" +
				"hundredth time: сотый раз\n\t\ttime-slot: тайм-слот",
		},
		{
			params:      []string{"англ time"},
			success:     true,
			spellResult: "Spelling: \n\ttimee -> [time]",
			trResult: "time()\n\tвремя (существительное)\n\tsyn: раз (), тайм ()\n\t" +
				"mean: timing, fold, half\n\texamples: \n\t\t" +
				"prehistoric time: доисторическое время\n\t\t" +
				"hundredth time: сотый раз\n\t\ttime-slot: тайм-слот",
		},
		{
			params:      []string{"time", "is", "running"},
			success:     true,
			spellResult: "Spelling: \n\ttimee -> [time]",
			trResult:    "тест",
		},
		{
			params:      []string{"time is running"},
			success:     true,
			spellResult: "Spelling: \n\ttimee -> [time]",
			trResult:    "тест",
		},
		{
			params:      []string{"en-ru", "time", "is", "running"},
			success:     true,
			spellResult: "Spelling: \n\ttimee -> [time]",
			trResult:    "тест",
		},
		{
			params:      []string{"en-ru", "time is running"},
			success:     true,
			spellResult: "Spelling: \n\ttimee -> [time]",
			trResult:    "тест",
		},
		{
			params:  []string{},
			success: false,
		},
	}
	for i, v := range values {
		sp, tr, err := y.GetTranslations(v.params)
		if err != nil {
			if v.success {
				t.Errorf("failed: %v", err)
				continue
			}
		} else {
			if !v.success {
				t.Errorf("[%d] error expected\n%q\n%q", i, sp, tr)
				continue
			}
		}
		// success result
		if sp != v.spellResult {
			t.Errorf("[%d] spelling: %q", i, sp)
		}
		if tr != v.trResult {
			t.Errorf("[%d] translation: %q", i, tr)
		}
	}
}
