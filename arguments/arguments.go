package arguments

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Text parses and builds text from parameters.
// It returns result text and count of words.
func Text(params []string) (string, uint) {
	var (
		builder strings.Builder
		count   uint
	)

	for _, p := range params {
		for _, word := range strings.Split(p, " ") {
			if w := strings.Trim(word, " \t\n\r"); len(w) > 0 {
				builder.WriteString(w)
				builder.WriteString(" ")
				count++
			}
		}
	}

	return strings.TrimSuffix(builder.String(), " "), count
}

// TextWithDictionary parses and builds text from parameters.
// It returns result text and true if it is a dictionary request.
func TextWithDictionary(params []string) (string, bool) {
	text, count := Text(params)

	if text == "" {
		// not found any words for translation
		return "", false
	}

	return text, count == 1
}

// Read reads text from reader.
// Basically, it reads text from stdin.
func Read(reader io.Reader) (string, error) {
	var b strings.Builder
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		if s := strings.TrimSpace(scanner.Text()); s != "" {
			b.WriteString(s)
			b.WriteString(" ")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.TrimSpace(b.String()), nil
}

// Build checks parameters and builds text from stdin if it is empty.
func Build(params []string, reader io.Reader) ([]string, error) {
	filledParams := make([]string, 0, len(params))

	for _, p := range params {
		if p = strings.TrimSpace(p); p != "" {
			filledParams = append(filledParams, p)
		}
	}

	if len(filledParams) > 0 {
		// there are explicit parameters for command line
		return filledParams, nil
	}

	// try to read text from stdin
	result, err := Read(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read text from stdin: %w", err)
	}

	if result == "" {
		return nil, fmt.Errorf("text is empty")
	}

	return []string{result}, nil
}
