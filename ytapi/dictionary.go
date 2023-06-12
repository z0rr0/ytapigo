// Copyright (c) 2023, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ytapi implements console text translation
// method using Yandex web services.
package ytapi

import (
	"fmt"
	"sort"
	"strings"
)

// TextPosGen is common struct for base data.
type TextPosGen struct {
	Text string `json:"text"`
	Pos  string `json:"pos"`
	Gen  string `json:"gen"`
	Fr   int    `json:"fr"`
}

// DictExample is an internal type of DictionaryResponse.
type DictExample struct {
	Pos  string              `json:"pos"`
	Text string              `json:"text"`
	Tr   []map[string]string `json:"tr"`
}

// JSONTrDictItem is an internal type of DictionaryResponse.
type JSONTrDictItem struct {
	TextPosGen
	Syn  []TextPosGen        `json:"syn"`
	Mean []map[string]string `json:"mean"`
	Ex   []DictExample       `json:"ex"`
}

// DictArticle is an internal type of DictionaryResponse.
type DictArticle struct {
	Pos  string           `json:"pos"`
	Text string           `json:"text"`
	Ts   string           `json:"ts"`
	Tr   []JSONTrDictItem `json:"tr"`
}

// DictionaryResponse is a type of a translation dictionary (from API response).
type DictionaryResponse struct {
	Head map[string]string `json:"head"`
	Def  []DictArticle     `json:"def"`
}

// DictionaryLanguages is a  list of dictionary's languages.
// It is sorted in ascending order.
type DictionaryLanguages []string

// Exists is an implementation of Exists() method for DictionaryResponse.
func (d *DictionaryResponse) Exists() bool {
	return d.String() != ""
}

// String is an implementation of String() method for DictionaryResponse.
// It returns a pretty formatted string.
func (d *DictionaryResponse) String() string {
	var (
		result, arResult, syn, mean, ex, extr []string
		txtResult, txtSyn, txtMean, txtEx     string
	)
	result = make([]string, len(d.Def))
	for i, def := range d.Def {
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

// String is an implementation of String() method for DictionaryLanguages pointer (LangChecker interface).
func (lch *DictionaryLanguages) String() string {
	return fmt.Sprintf("Dictionary languages:\n%v", strings.Join(*lch, ", "))
}

// Len returns a list of dictionary languages.
func (lch *DictionaryLanguages) Len() int {
	return len(*lch)
}

// Sort sorts dictionary languages.
func (lch *DictionaryLanguages) Sort() {
	//sort.Sort(lch)
	sort.Slice(*lch, func(i, j int) bool {
		return (*lch)[i] < (*lch)[j]
	})
}

// Contains is an implementation of Contains() method for DictionaryLanguages pointer (LangChecker interface).
func (lch *DictionaryLanguages) Contains(s string) bool {
	var data []string = *lch
	if len(data) == 0 {
		return false
	}
	i := sort.Search(lch.Len(), func(i int) bool { return data[i] >= s })
	return i < lch.Len() && data[i] == s
}

// Description is an implementation of Description() method for
// DictionaryLanguages pointer (LangChecker interface).
func (lch *DictionaryLanguages) Description() string {
	return fmt.Sprintf("Length=%v\n%v", len(*lch), lch.String())
}
