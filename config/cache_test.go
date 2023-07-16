package config

import (
	"os"
	"path"
	"testing"
	"time"
)

func deleteFile(fileName string, t *testing.T) {
	if err := os.Remove(fileName); err != nil {
		t.Error(err)
	}
}

func TestReadCachedToken(t *testing.T) {
	expiredFile := path.Join(os.TempDir(), "ytapigo_read_cached_token_1.json")
	actualFile := path.Join(os.TempDir(), "ytapigo_read_cached_token_2.json")
	badJSONFile := path.Join(os.TempDir(), "ytapigo_read_cached_token_3.json")
	badFormatFile := path.Join(os.TempDir(), "ytapigo_read_cached_token_4.json")

	osErr := os.WriteFile(expiredFile, []byte(`{"token":"abc123","expired":"2011-02-03 04:05:06"}`), 0660)
	if osErr != nil {
		t.Fatal(osErr)
	}
	defer deleteFile(expiredFile, t)

	expiredAt := time.Now().Add(time.Hour).UTC().Format(time.DateTime)
	osErr = os.WriteFile(actualFile, []byte(`{"token":"abc124","expired":"`+expiredAt+`"}`), 0660)
	if osErr != nil {
		t.Fatal(osErr)
	}
	defer deleteFile(actualFile, t)

	osErr = os.WriteFile(badJSONFile, []byte(`{"token":"abc125`), 0660)
	if osErr != nil {
		t.Fatal(osErr)
	}
	defer deleteFile(badJSONFile, t)

	osErr = os.WriteFile(badFormatFile, []byte(`{"token":"abc126","expired":"2011-02-03/040506"}`), 0660)
	if osErr != nil {
		t.Fatal(osErr)
	}
	defer deleteFile(badFormatFile, t)

	testCases := []struct {
		fileName  string
		expected  string
		withError bool
	}{
		{},
		{fileName: actualFile + ".not-exists"},
		{fileName: expiredFile, expected: ""},
		{fileName: actualFile, expected: "abc124"},
		{fileName: badJSONFile, withError: true},
		{fileName: badFormatFile, withError: true},
	}

	for i, tc := range testCases {
		token, err := readCachedToken(tc.fileName)

		if tc.withError {
			if err == nil {
				t.Errorf("test case %d: expected error, got nil", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("test case %d: unexpected error: %v", i, err)
			continue
		}

		if token != tc.expected {
			t.Errorf("test case %d: expected %q, got %q", i, tc.expected, token)
		}
	}
}

func TestWriteCachedToken(t *testing.T) {
	testCases := []struct {
		fileName  string
		token     string
		expiresAt time.Time
		expected  string
	}{
		{},
		{fileName: "", expected: ""},
		{
			fileName:  path.Join(os.TempDir(), "write_cached_token_1.json"),
			token:     "abc123",
			expiresAt: time.Date(2011, 2, 3, 4, 5, 6, 0, time.UTC),
			expected:  `{` + "\n" + `  "token": "abc123",` + "\n" + `  "expired": "2011-02-03 04:05:06"` + "\n" + `}`,
		},
		{
			fileName:  path.Join(os.TempDir(), "write_cached_token_2.json"),
			token:     "abc123bca",
			expiresAt: time.Date(2021, 12, 11, 10, 9, 8, 0, time.UTC),
			expected:  `{` + "\n" + `  "token": "abc123bca",` + "\n" + `  "expired": "2021-12-11 10:09:08"` + "\n" + `}`,
		},
	}

	for i, tc := range testCases {
		err := writeCachedToken(tc.fileName, tc.token, tc.expiresAt)

		if err != nil {
			t.Errorf("test case %d: unexpected error: %v", i, err)
			continue
		}

		if tc.fileName == "" {
			continue
		}

		data, err := os.ReadFile(tc.fileName)
		if err != nil {
			t.Errorf("test case %d: unexpected error: %v", i, err)
		}

		if string(data) != tc.expected {
			t.Errorf("test case %d: expected %q, got %q", i, tc.expected, string(data))
		}

		deleteFile(tc.fileName, t)
	}
}
