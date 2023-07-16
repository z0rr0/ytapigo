package result

import (
	"errors"
	"fmt"
	"sort"
)

// Languages is an interface for languages set.
type Languages interface {
	String() string
	Contains(string, string) bool
}

// Translation is an interface for Translation of translation or spelling check requests.
type Translation interface {
	String() string
	Exists() bool
}

// Item is a translation item with priority and error.
type Item struct {
	Translation Translation
	Priority    uint8
	Err         error
}

// Append adds item to slice or joins error.
func (item *Item) Append(err error, items *[]Item) error {
	if item.Err != nil {
		err = errors.Join(err, item.Err)
	} else {
		*items = append(*items, *item)
	}
	return err
}

// orderedTranslations converts response items to ordered translations.
func orderedTranslations(items []Item) []Translation {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Priority < items[j].Priority
	})

	translations := make([]Translation, 0, len(items))
	for i := range items {
		translations = append(translations, items[i].Translation)
	}

	return translations
}

// Build builds results from ch-channel to a slice.
func Build(ch <-chan Item, count int) ([]Translation, error) {
	var (
		err       error
		responses = make([]Item, 0, count)
	)

	for i := 0; i < count; i++ {
		item := <-ch
		err = item.Append(err, &responses)
	}

	if err != nil {
		return nil, err
	}

	return orderedTranslations(responses), nil
}

// Show prints results to stdout.
func Show(translations []Translation) {
	for _, t := range translations {
		if t.Exists() {
			fmt.Println(t)
		}
	}
}
