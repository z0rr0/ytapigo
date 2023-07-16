package handle

import "testing"

func TestASCIIDetection(t *testing.T) {
	cases := []struct {
		text         string
		fromLanguage string
		toLanguage   string
	}{
		{fromLanguage: Ru, toLanguage: En},
		{text: "你好世界！", fromLanguage: Ru, toLanguage: En},
		{text: ",.;", fromLanguage: Ru, toLanguage: En},
		{text: "I'm", fromLanguage: En, toLanguage: Ru},
		{text: "Hello, world!", fromLanguage: En, toLanguage: Ru},
		{text: "Привет, мир!", fromLanguage: Ru, toLanguage: En},
		{text: "Привет, мир! Hello, world!", fromLanguage: En, toLanguage: Ru},
		{text: "Привет, мир всем! Hello, world!", fromLanguage: Ru, toLanguage: En},
	}

	for i, c := range cases {
		fromLanguage, toLanguage := asciiDetection(c.text)
		if fromLanguage != c.fromLanguage || toLanguage != c.toLanguage {
			t.Errorf("case %v: expected %v-%v, got %v-%v", i, c.fromLanguage, c.toLanguage, fromLanguage, toLanguage)
		}
	}
}

func TestSplitDetection(t *testing.T) {
	cases := []struct {
		direction    string
		fromLanguage string
		toLanguage   string
		withError    bool
	}{
		{direction: "ru-en", fromLanguage: Ru, toLanguage: En},
		{direction: "ru-en-kz", withError: true},
		{direction: "ru", withError: true},
		{direction: "bad", withError: true},
		{direction: "too long", withError: true},
		{direction: "no-ru", fromLanguage: "no", toLanguage: "ru"},
	}

	for i, c := range cases {
		fromLanguage, toLanguage, err := splitDetection(c.direction)

		if err != nil {
			if !c.withError {
				t.Errorf("case %v: unexpected error: %v", i, err)
			}
		} else {
			if fromLanguage != c.fromLanguage || toLanguage != c.toLanguage {
				t.Errorf("case %v: expected %v-%v, got %v-%v", i, c.fromLanguage, c.toLanguage, fromLanguage, toLanguage)
			}
		}
	}
}

func TestKnowLanguages(t *testing.T) {
	cases := []struct {
		lang     []string
		expected bool
	}{
		{lang: []string{En}, expected: true},
		{lang: []string{Ru}, expected: true},
		{lang: []string{Uk}, expected: true},
		{lang: []string{En, Ru}, expected: true},
		{lang: []string{En, Ru, "no"}, expected: false},
		{lang: []string{En, Ru, "no", Uk}, expected: false},
		{lang: []string{En, Ru, Uk}, expected: true},
		{lang: []string{"no"}, expected: false},
		{lang: []string{}, expected: false},
	}

	for i, c := range cases {
		if knowLanguages(c.lang...) != c.expected {
			t.Errorf("case %v: expected %v, got %v", i, c.expected, !c.expected)
		}
	}
}
