// Copyright (c) 2019, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ytapigo implements console text translation
// method using Yandex web services.
//
package ytapigo

import (
	"fmt"
	"sort"
	"strings"
)

// TranslateResponse is a type of translation response.
type TranslateResponse struct {
	Code float64  `json:"code"`
	Lang string   `json:"lang"`
	Text []string `json:"text"`
}

// String is an implementation of String() method for TranslateResponse pointer.
func (t *TranslateResponse) String() string {
	if len(t.Text) == 0 {
		return ""
	}
	return t.Text[0]
}

// Exists is an implementation of Exists() method for TranslateResponse pointer.
func (t *TranslateResponse) Exists() bool {
	return t.String() != ""
}

// TranslateLanguages is a list of translation's languages (from JSON response).
// "Dirs" field is an array that sorted in ascending order.
type TranslateLanguages struct {
	Dirs  []string          `json:"dirs"`
	Langs map[string]string `json:"langs"`
}

// String is an implementation of String() method for TranslateLanguages
// pointer (LangChecker interface).
func (ltr *TranslateLanguages) String() string {
	return fmt.Sprintf("%v", strings.Join(ltr.Dirs, ", "))
}

// Contains is an implementation of Contains() method for TranslateLanguages
// pointer (LangChecker interface).
func (ltr *TranslateLanguages) Contains(s string) bool {
	if len(ltr.Dirs) == 0 {
		return false
	}
	if !sort.StringsAreSorted(ltr.Dirs) {
		sort.Strings(ltr.Dirs)
	}
	if i := sort.SearchStrings(ltr.Dirs, s); i < len(ltr.Dirs) && ltr.Dirs[i] == s {
		return true
	}
	return false
}

// Description is an implementation of Description() method
// for TranslateLanguages pointer (LangChecker interface).
func (ltr *TranslateLanguages) Description() string {
	const n int = 3
	var (
		collen, counter int
	)
	counter = len(ltr.Langs)
	i, descstr := 0, make([]string, counter)
	for k, v := range ltr.Langs {
		if len(v) > 0 {
			descstr[i] = fmt.Sprintf("%v - %v", k, v)
			i++
		}
	}
	sort.Strings(descstr)

	if (counter % n) != 0 {
		collen = counter/n + 1
	} else {
		collen = counter / n
	}
	output := make([]string, collen)
	for j := 0; j < collen; j++ {
		switch {
		case j+2*collen < counter:
			output[j] = fmt.Sprintf("%-25v %-25v %-25v", descstr[j], descstr[j+collen], descstr[j+2*collen])
		case j+collen < counter:
			output[j] = fmt.Sprintf("%-25v %-25v", descstr[j], descstr[j+collen])
		default:
			output[j] = fmt.Sprintf("%-25v", descstr[j])
		}
	}
	return strings.Join(output, "\n")
}
