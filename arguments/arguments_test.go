package arguments

import (
	"slices"
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name       string
		params     []string
		fromReader string
		expected   []string
		error      string
	}{
		{name: "empty", error: "text is empty"},
		{name: "space_params", error: "text is empty", params: []string{"  ", ""}},
		{name: "with_param", params: []string{"Hello"}, expected: []string{"Hello"}},
		{name: "strip_params", params: []string{"  Hello  ", "  world  "}, expected: []string{"Hello", "world"}},
		{name: "from_reader", fromReader: " Hello world  ", expected: []string{"Hello world"}},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			text, err := Build(tc.params, strings.NewReader(tc.fromReader))

			if err != nil {
				if e := err.Error(); e != tc.error {
					t.Errorf("expected error %q, got %q", tc.error, e)
				}
			} else {
				if slices.Compare(text, tc.expected) != 0 {
					t.Errorf("expected text %#v, got %#v", tc.expected, text)
				}
			}
		})
	}
}

func TestTextWithDictionary(t *testing.T) {
	tests := []struct {
		name         string
		params       []string
		expected     string
		isDictionary bool
	}{
		{name: "empty", params: []string{}, expected: ""},
		{
			name:     "two_words",
			params:   []string{"Hello", "world"},
			expected: "Hello world",
		},
		{
			name:     "some_separate_text",
			params:   []string{"This", "is", "a", "test"},
			expected: "This is a test",
		},
		{
			name:     "some_text_with_spaces",
			params:   []string{" This", "  is\n", " \ta ", " test   "},
			expected: "This is a test",
		},
		{
			name:         "dictionary",
			params:       []string{"  ", "test"},
			expected:     "test",
			isDictionary: true,
		},
		{
			name:         "dictionary_with_spaces",
			params:       []string{"  ", "  book ", "   \n"},
			expected:     "book",
			isDictionary: true,
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tests[i].name, func(t *testing.T) {
			text, isDictionary := TextWithDictionary(tc.params)

			if text != tc.expected {
				t.Errorf("expected text %q, but got %q", tc.expected, text)
			}

			if isDictionary != tc.isDictionary {
				t.Errorf("expected isDictionary %v, but got %v", tc.isDictionary, isDictionary)
			}
		})
	}
}

func BenchmarkBuildText(b *testing.B) {
	params := []string{"This is a sample sentence", "This is another sentence", "And this is yet another one"}
	expected := "This is a sample sentence This is another sentence And this is yet another one"

	for n := 0; n < b.N; n++ {
		result, _ := Text(params)

		if result != expected {
			b.Errorf("expected text %q, but got %q", expected, result)
		}
	}
}

func FuzzBuildText(f *testing.F) {
	testCases := []string{
		"",
		" :: ",
		"Hello world",
		"This is a sample sentence",
		"This is a sample sentence::This is another sentence",
		"This is a sample sentence::This is another sentence::And this is yet another one",
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, args string) {
		params := strings.Split(args, "::")
		text, count := Text(params)

		if count != 0 {
			if text == "" {
				t.Errorf("expected non-empty text count=%d, but got %q", count, text)
			}
		} else {
			if text != "" {
				t.Errorf("expected empty text, but got %q", text)
			}
		}
	})
}
