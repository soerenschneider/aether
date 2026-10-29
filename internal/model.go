package internal

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
)

var ErrTemplate = errors.New("template error")

type Data struct {
	Summary                    []string
	RenderedDefaultTemplate    []byte
	RenderedSimplifiedTemplate []byte
	// RenderedTile is a compact summary shown at the top of the page, empty if the datasource has nothing to show
	RenderedTile []byte
}

// Page is a page of its own that a datasource offers next to the dashboard, e.g. a month view of the calendar.
type Page struct {
	Title   string
	Content []byte
}

type PageProvider interface {
	// PageName is the path of the page relative to the dashboard, empty if the datasource offers no page
	PageName() string
	RenderPage(r *http.Request) (*Page, error)
}

// RenderTile renders the optional "tile" block of a template, see Data.RenderedTile.
func RenderTile(tmpl *template.Template, data any) ([]byte, error) {
	if tmpl == nil || tmpl.Lookup("tile") == nil {
		return nil, nil
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "tile", data); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
}
