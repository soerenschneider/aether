package carddav

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/emersion/go-webdav/carddav"
	"github.com/soerenschneider/aether/internal/templates"
)

type emptyAddressBook struct{}

func (emptyAddressBook) FindAddressBookHomeSet(_ context.Context, _ string) (string, error) {
	return "/home/", nil
}

func (emptyAddressBook) FindAddressBooks(_ context.Context, _ string) ([]carddav.AddressBook, error) {
	return []carddav.AddressBook{{Path: "/home/contacts/"}}, nil
}

func (emptyAddressBook) QueryAddressBook(_ context.Context, _ string, _ *carddav.AddressBookQuery) ([]carddav.AddressObject, error) {
	return nil, nil
}

func TestWithTemplateFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "contacts.html")
	if err := os.WriteFile(file, []byte(`custom:{{ len .Cards }}`), 0o600); err != nil {
		t.Fatal(err)
	}

	tmpl := templates.TemplateData{DefaultTemplate: []byte("default"), SimpleTemplate: []byte("simple")}
	ds, err := New("http://localhost", tmpl, WithTemplateFile(file), WithClient(emptyAddressBook{}))
	if err != nil {
		t.Fatal(err)
	}

	data, err := ds.GetData(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data.RenderedDefaultTemplate); got != "custom:0" {
		t.Errorf("expected the custom template to be used, got %q", got)
	}

	if _, err := New("http://localhost", tmpl, WithTemplateFile(filepath.Join(t.TempDir(), "missing.html"))); err == nil {
		t.Error("expected an error for a missing template file")
	}
}
