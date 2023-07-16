package result

import (
	"errors"
	"testing"
)

// testTranslation is a test implementation of Translation interface.
type testTranslation struct {
	value string
}

func (tt testTranslation) String() string {
	return tt.value
}

func (tt testTranslation) Exists() bool {
	return tt.value != ""
}

func TestBuild(t *testing.T) {
	testCases := []struct {
		name     string
		items    []Item
		expected []Translation
		error    string
	}{
		{name: "empty"},
		{name: "no_items", items: []Item{}},
		{
			name:     "one_item",
			items:    []Item{{Translation: testTranslation{"test"}}},
			expected: []Translation{testTranslation{"test"}},
		},
		{
			name: "tow_items",
			items: []Item{
				{Translation: testTranslation{"test2"}, Priority: 2},
				{Translation: testTranslation{"test1"}, Priority: 1},
			},
			expected: []Translation{
				testTranslation{"test1"},
				testTranslation{"test2"},
			},
		},
		{
			name: "many_items",
			items: []Item{
				{Translation: testTranslation{"test2"}, Priority: 2},
				{Translation: testTranslation{"test1"}, Priority: 1},
				{Translation: testTranslation{"test4"}, Priority: 4},
				{Translation: testTranslation{"test3"}, Priority: 3},
			},
			expected: []Translation{
				testTranslation{"test1"},
				testTranslation{"test2"},
				testTranslation{"test3"},
				testTranslation{"test4"},
			},
		},
		{
			name: "one_error",
			items: []Item{
				{Translation: testTranslation{"test2"}, Priority: 2},
				{Translation: testTranslation{"test1"}, Priority: 1, Err: errors.New("test-error-1")},
			},
			error: "test-error-1",
		},
		{
			name: "two_errors",
			items: []Item{
				{Translation: testTranslation{"test2"}, Priority: 2, Err: errors.New("test-error-2")},
				{Translation: testTranslation{"test1"}, Priority: 1, Err: errors.New("test-error-1")},
			},
			error: "test-error-2\ntest-error-1",
		},
		{
			name: "many_errors",
			items: []Item{
				{Translation: testTranslation{"test1"}, Priority: 1, Err: errors.New("test-error-1")},
				{Translation: testTranslation{"test2"}, Priority: 2},
				{Translation: testTranslation{"test3"}, Priority: 3, Err: errors.New("test-error-3")},
				{Translation: testTranslation{"test4"}, Priority: 4},
			},
			error: "test-error-1\ntest-error-3",
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			ch := make(chan Item)
			defer close(ch)

			args := testCases[i]
			go func() {
				for j := range args.items {
					ch <- args.items[j]
				}
			}()

			result, err := Build(ch, len(args.items))

			if err != nil {
				if e := err.Error(); args.error != e {
					tt.Errorf("mismatch errors, got: %q, expected: %q", e, args.error)
				}
				return
			}

			// no error
			if args.error != "" {
				tt.Errorf("expected error: %s", args.error)
				return
			}

			if len(result) != len(args.expected) {
				tt.Errorf("result: %d, expected: %d", len(result), len(args.expected))
				return
			}

			for k := range result {
				if result[k] != args.expected[k] {
					tt.Errorf("result: %#v, expected: %#v", result[k], args.expected[k])
				}
			}
		})
	}

}
