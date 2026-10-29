package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSecretFile(t *testing.T) {
	tests := map[string]string{
		"s3cret\n":      "s3cret",
		"s3cret\r\n":    "s3cret",
		"s3cret":        "s3cret",
		"s3cret\n\n":    "s3cret",
		" with spaces ": " with spaces ",
		"":              "",
	}

	dir := t.TempDir()
	for content, want := range tests {
		file := filepath.Join(dir, "secret")
		if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := ReadSecretFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("ReadSecretFile(%q) = %q, want %q", content, got, want)
		}
	}

	if _, err := ReadSecretFile(filepath.Join(dir, "missing")); err == nil {
		t.Error("expected an error for a missing file")
	}
}
