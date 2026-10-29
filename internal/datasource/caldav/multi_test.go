package caldav

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/emersion/go-webdav/caldav"
	"github.com/soerenschneider/aether/internal/templates"
)

type staticClient struct {
	entries []Entry
	err     error
}

func (s staticClient) FetchData(_ context.Context) ([]Entry, error) {
	return s.entries, s.err
}

func TestMultiClient_FetchData(t *testing.T) {
	shared := entry("dinner", at(1, 19, 0), at(1, 21, 0))
	client, err := NewMultiClient(
		NamedCalendar{Name: "Private", Client: staticClient{entries: []Entry{entry("review", at(0, 9, 0), at(0, 10, 0)), shared}}},
		NamedCalendar{Name: "Family", Client: staticClient{entries: []Entry{entry("pick up kids", at(0, 16, 0), at(0, 17, 0)), shared}}},
	)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := client.FetchData(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]struct {
		calendar string
		color    int
	}{
		"review":       {"Private", 0},
		"dinner":       {"Private, Family", 0},
		"pick up kids": {"Family", 1},
	}
	if len(entries) != len(want) {
		t.Fatalf("expected %d entries, got %d: %v", len(want), len(entries), entries)
	}
	for _, e := range entries {
		w, ok := want[e.Summary]
		if !ok || e.Calendar != w.calendar || e.CalendarColor != w.color {
			t.Errorf("entry %q: got calendar %q color %d, want %+v", e.Summary, e.Calendar, e.CalendarColor, w)
		}
	}
}

func TestMultiClient_FetchData_singleCalendarIsNotLabelled(t *testing.T) {
	client, _ := NewMultiClient(NamedCalendar{Name: "", Client: staticClient{entries: []Entry{entry("review", at(0, 9, 0), at(0, 10, 0))}}})
	entries, err := client.FetchData(context.Background())
	if err != nil || len(entries) != 1 || entries[0].Calendar != "" {
		t.Errorf("unexpected result: %v, %v", entries, err)
	}
}

func TestMultiClient_FetchData_failures(t *testing.T) {
	boom := errors.New("boom")
	client, _ := NewMultiClient(
		NamedCalendar{Name: "Private", Client: staticClient{entries: []Entry{entry("review", at(0, 9, 0), at(0, 10, 0))}}},
		NamedCalendar{Name: "Family", Client: staticClient{err: boom}},
	)

	entries, err := client.FetchData(context.Background())
	var partial *PartialError
	if !errors.As(err, &partial) {
		t.Fatalf("expected a partial error, got %v", err)
	}
	if len(partial.Calendars) != 1 || partial.Calendars[0] != "Family" || !errors.Is(err, boom) {
		t.Errorf("unexpected partial error: %v", partial)
	}
	if len(entries) != 1 {
		t.Errorf("expected the entries of the working calendar, got %v", entries)
	}

	client, _ = NewMultiClient(
		NamedCalendar{Name: "Private", Client: staticClient{err: boom}},
		NamedCalendar{Name: "Family", Client: staticClient{err: boom}},
	)
	_, err = client.FetchData(context.Background())
	if err == nil || errors.As(err, &partial) {
		t.Errorf("expected a regular error if all calendars fail, got %v", err)
	}
}

func Test_selectCalendar(t *testing.T) {
	calendars := []caldav.Calendar{
		{Name: "Personal", Path: "/remote.php/dav/calendars/soeren/personal/"},
		{Name: "Family", Path: "/remote.php/dav/calendars/soeren/family-shared/"},
	}

	tests := []struct {
		wanted   string
		wantPath string
		wantErr  bool
	}{
		{"", calendars[0].Path, false},
		{"family", calendars[1].Path, false},
		{"/remote.php/dav/calendars/soeren/family-shared/", calendars[1].Path, false},
		{"family-shared", calendars[1].Path, false},
		{"work", "", true},
	}

	for _, tt := range tests {
		got, err := selectCalendar(calendars, tt.wanted)
		if (err != nil) != tt.wantErr {
			t.Errorf("selectCalendar(%q) error = %v, wantErr %v", tt.wanted, err, tt.wantErr)
			continue
		}
		if got.Path != tt.wantPath {
			t.Errorf("selectCalendar(%q) = %q, want %q", tt.wanted, got.Path, tt.wantPath)
		}
	}

	if _, err := selectCalendar(nil, ""); err == nil {
		t.Error("expected an error without calendars")
	}
}

func TestCaldavDatasource_GetData_incomplete(t *testing.T) {
	client, _ := NewMultiClient(
		NamedCalendar{Name: "Private", Client: staticClient{entries: []Entry{entry("review", time.Now().Add(time.Hour), time.Now().Add(2*time.Hour))}}},
		NamedCalendar{Name: "Family", Client: staticClient{err: errors.New("boom")}},
	)
	ds, err := New(client, templates.TemplateData{
		DefaultTemplate: []byte(`{{ range .Failed }}failed:{{ . }} {{ end }}{{ range .Entries }}{{ .Summary }}{{ end }}`),
		SimpleTemplate:  []byte(`{{ len .Entries }}`),
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := ds.GetData(context.Background())
	if err != nil {
		t.Fatalf("expected the working calendar to be shown, got %v", err)
	}
	if !data.Incomplete {
		t.Error("expected the data to be marked as incomplete")
	}
	if got := string(data.RenderedDefaultTemplate); got != "failed:Family review" {
		t.Errorf("unexpected rendering %q", got)
	}
}
