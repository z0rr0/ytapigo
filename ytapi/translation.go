// Copyright (c) 2019, Alexander Zaitsev <me@axv.email>. All rights reserved.
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

// TranslateResponseItem is an item for translation request.
type TranslateResponseItem struct {
	Text                 string `json:"text"`
	DetectedLanguageCode string `json:"detectedLanguageCode"`
}

// TranslateResponse is a type of translation response.
type TranslateResponse struct {
	Translations []TranslateResponseItem `json:"translations"`
}

// String is an implementation of String() method for TranslateResponse pointer.
func (t *TranslateResponse) String() string {
	if len(t.Translations) == 0 {
		return ""
	}
	texts := make([]string, len(t.Translations))
	for i := range t.Translations {
		texts[i] = t.Translations[i].Text
	}
	return strings.Join(texts, "\n")
}

// Exists is an implementation of Exists() method for TranslateResponse pointer.
func (t *TranslateResponse) Exists() bool {
	return t.String() != ""
}

// TranslateLanguage is language info item.
type TranslateLanguage struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// TranslateLanguages is a list of translation's languages (from JSON response).
type TranslateLanguages struct {
	Languages []TranslateLanguage `json:"languages"`
}

func (tl *TranslateLanguage) String() string {
	return fmt.Sprintf("%v - %v", tl.Code, tl.Name)
}

// String is an implementation of String() method for TranslateLanguages
// pointer (LangChecker interface).
func (ltr *TranslateLanguages) String() string {
	codes := make([]string, ltr.Len())
	for i, v := range ltr.Languages {
		codes[i] = v.Code
	}
	return fmt.Sprintf("Translation languages:\n%v\n%v", strings.Join(codes, ", "), ltr.Description())
}

func (ltr TranslateLanguages) Len() int {
	return len(ltr.Languages)
}

func (ltr TranslateLanguages) Swap(i, j int) {
	ltr.Languages[i], ltr.Languages[j] = ltr.Languages[j], ltr.Languages[i]
}

func (ltr TranslateLanguages) Less(i, j int) bool {
	return ltr.Languages[i].Code < ltr.Languages[j].Code
}

// Sort sorts languages by code.
func (ltr *TranslateLanguages) Sort() {
	sort.Sort(ltr)
}

func (ltr *TranslateLanguages) containsLanguage(s string) bool {
	if len(ltr.Languages) == 0 {
		return false
	}
	i := sort.Search(ltr.Len(), func(i int) bool { return ltr.Languages[i].Code >= s })
	return i < ltr.Len() && ltr.Languages[i].Code == s
}

// Contains is an implementation of Contains() method for TranslateLanguages
// pointer (LangChecker interface).
func (ltr *TranslateLanguages) Contains(s string) bool {
	languages := strings.SplitN(s, "-", 2)
	if len(languages) < 2 {
		return false
	}
	return ltr.containsLanguage(languages[0]) && ltr.containsLanguage(languages[1])
}

// Description is an implementation of Description() method
// for TranslateLanguages pointer (LangChecker interface).
func (ltr *TranslateLanguages) Description() string {
	const n int = 3
	var colLen int

	counter := ltr.Len()
	if (counter % n) != 0 {
		colLen = counter/n + 1
	} else {
		colLen = counter / n
	}
	output := make([]string, colLen)
	for j := 0; j < colLen; j++ {
		switch {
		case j+2*colLen < counter:
			output[j] = fmt.Sprintf(
				"%-25v %-25v %-25v",
				ltr.Languages[j].String(), ltr.Languages[j+colLen].String(), ltr.Languages[j+2*colLen].String(),
			)
		case j+colLen < counter:
			output[j] = fmt.Sprintf("%-25v %-25v", ltr.Languages[j].String(), ltr.Languages[j+colLen].String())
		default:
			output[j] = fmt.Sprintf("%-25v", ltr.Languages[j].String())
		}
	}
	return strings.Join(output, "\n")
}
