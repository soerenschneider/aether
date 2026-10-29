package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEmailConfig_secretFiles(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		file := filepath.Join(dir, name)
		if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return file
	}

	conf := EmailConfig{
		UsernameFile:   write("username", "aether\n"),
		PasswordFile:   write("password", "s3cret\n"),
		FromFile:       write("from", "aether@example.org\n"),
		RecipientsFile: write("recipients", "a@example.org, b@example.org\nc@example.org\n"),
	}

	if got, err := conf.GetUsername(); err != nil || got != "aether" {
		t.Errorf("GetUsername() = %q, %v", got, err)
	}
	if got, err := conf.GetPassword(); err != nil || got != "s3cret" {
		t.Errorf("GetPassword() = %q, %v", got, err)
	}
	if got, err := conf.GetFrom(); err != nil || got != "aether@example.org" {
		t.Errorf("GetFrom() = %q, %v", got, err)
	}
	want := []string{"a@example.org", "b@example.org", "c@example.org"}
	if got, err := conf.GetRecipients(); err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("GetRecipients() = %q, %v", got, err)
	}

	conf.RecipientsFile = write("empty", "\n")
	if _, err := conf.GetRecipients(); err == nil {
		t.Error("expected an error for a file without recipients")
	}
}
