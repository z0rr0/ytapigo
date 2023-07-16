// Copyright (c) 2023, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dictionary implements dictionary translation.
package dictionary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/z0rr0/ytapigo/cloud"
	"github.com/z0rr0/ytapigo/config"
)

// TranslationURL is a URL for translation request.
// Documentation https://yandex.com/dev/dictionary/doc/dg/reference/lookup.html
const TranslationURL = "https://dictionary.yandex.net/api/v1/dicservice.json/lookup"

// TextPosGen is common struct for base data.
type TextPosGen struct {
	Text string `json:"text"`
	Pos  string `json:"pos"`
	Gen  string `json:"gen"`
	Fr   int    `json:"fr"`
}

// Example is an internal type of Response.
type Example struct {
	Pos  string              `json:"pos"`
	Text string              `json:"text"`
	Tr   []map[string]string `json:"tr"`
}

// TrItem is an internal type of Response.
type TrItem struct {
	TextPosGen
	Syn  []TextPosGen        `json:"syn"`
	Mean []map[string]string `json:"mean"`
	Ex   []Example           `json:"ex"`
}

// Article is an internal type of Response.
type Article struct {
	Pos  string   `json:"pos"`
	Text string   `json:"text"`
	Ts   string   `json:"ts"`
	Tr   []TrItem `json:"tr"`
}

// Response is a type of translation dictionary (from API response).
type Response struct {
	Head map[string]string `json:"head"`
	Def  []Article         `json:"def"`
}

// Exists is an implementation of Exists() method for Response.
func (r *Response) Exists() bool {
	return r.String() != ""
}

// String is an implementation of String() method for Response.
// It returns a pretty formatted string.
func (r *Response) String() string {
	var (
		result, arResult, syn, mean, ex, extr []string
		txtResult, txtSyn, txtMean, txtEx     string
	)
	result = make([]string, len(r.Def))
	for i, def := range r.Def {
		ts := ""
		if def.Ts != "" {
			ts = fmt.Sprintf(" [%v] ", def.Ts)
		}
		txtResult = fmt.Sprintf("%v%v(%v)", def.Text, ts, def.Pos)
		arResult = make([]string, len(def.Tr))
		for j, tr := range def.Tr {
			syn, mean, ex = make([]string, len(tr.Syn)), make([]string, len(tr.Mean)), make([]string, len(tr.Ex))
			txtSyn, txtMean, txtEx = "", "", ""
			for k, s := range tr.Syn {
				syn[k] = fmt.Sprintf("%v (%v)", s.Text, s.Pos)
			}
			for k, v := range tr.Mean {
				mean[k] = v["text"]
			}
			for k, v := range tr.Ex {
				extr = make([]string, len(v.Tr))
				for t, trv := range v.Tr {
					extr[t] = trv["text"]
				}
				ex[k] = fmt.Sprintf("%v: %v", v.Text, strings.Join(extr, ", "))
			}
			if len(syn) > 0 {
				txtSyn = fmt.Sprintf("\n\tsyn: %v", strings.Join(syn, ", "))
			}
			if len(mean) > 0 {
				txtMean = fmt.Sprintf("\n\tmean: %v", strings.Join(mean, ", "))
			}
			if len(ex) > 0 {
				txtEx = fmt.Sprintf("\n\texamples: \n\t\t%v", strings.Join(ex, "\n\t\t"))
			}

			arResult[j] = fmt.Sprintf("\t%v (%v)%v%v%v", tr.Text, tr.Pos, txtSyn, txtMean, txtEx)
		}
		result[i] = fmt.Sprintf("%v\n%v", txtResult, strings.Join(arResult, "\n"))
	}
	return strings.Join(result, "\n")
}

// Request is a type of translation request items.
type Request struct {
	Key                string
	Text               string
	TargetLanguageCode string
	SourceLanguageCode string
}

func (r *Request) reader() io.Reader {
	lang := fmt.Sprintf("%s-%s", r.SourceLanguageCode, r.TargetLanguageCode)
	params := url.Values{"lang": {lang}, "text": {r.Text}, "key": {r.Key}}

	return strings.NewReader(params.Encode())
}

// Translate returns translated dictionary article.
func Translate(ctx context.Context, client *http.Client, cfg *config.Config, r *Request) (*Response, error) {
	body, err := cloud.Request(ctx, client, r.reader(), cfg.GetURL(TranslationURL), "", cfg.UserAgent, false, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to translate: %w", err)
	}

	response := &Response{}
	if err = json.Unmarshal(body, response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal translation response: %w", err)
	}

	return response, nil
}
