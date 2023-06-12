// Copyright (c) 2023, Alexander Zaitsev <me@axv.email>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ytapi implements console text translation
// method using Yandex web services.
package ytapi

import (
	"fmt"
	"strings"
)

// SpellerItem is a type of spell check (from JSON API response).
type SpellerItem struct {
	Word string   `json:"word"`
	S    []string `json:"s"`
	Code float64  `json:"code"`
	Pos  float64  `json:"pos"`
	Row  float64  `json:"row"`
	Col  float64  `json:"col"`
	Len  float64  `json:"len"`
}

// SpellerResponse is an array of spelling results.
type SpellerResponse []SpellerItem

// Exists is an implementation of Exists() method for SpellerResponse.
func (s *SpellerResponse) Exists() bool {
	return len(*s) > 0
}

// String is an implementation of String() method for SpellerResponse.
func (s *SpellerResponse) String() string {
	n := len(*s)
	if n == 0 {
		return ""
	}
	items := make([]string, n)
	for i, v := range *s {
		if v.Exists() {
			items[i] = v.String()
		}
	}
	return fmt.Sprintf("Spelling: \n\t%v", strings.Join(items, "\n\t"))
}

// Exists is an implementation of Exists() method for SpellerItem.
// (Translator interface).
func (si *SpellerItem) Exists() bool {
	return (len(si.Word) > 0) || (len(si.S) > 0)
}

// String is an implementation of String() method for SpellerItem.
func (si *SpellerItem) String() string {
	return fmt.Sprintf("%v -> %v", si.Word, si.S)
}
