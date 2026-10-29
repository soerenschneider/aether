package serve

import (
	"strings"
	"testing"
)

const testPage = `<!DOCTYPE html>
<html><head><title>aether</title><style>.card { display: grid; }</style><script src="static/htmx.min.js"></script></head>
<body>
<div id="app" hx-get="" hx-trigger="every 60s">
  <div class="status"><span class="live"></span><span class="updated">Updated 10:05</span><button class="refresh">↻</button></div>
  <nav class="now"><a class="tile" href="calendar"><div class="tile-value">Standup</div></a><a class="tile" href="#taskwarrior"><div class="tile-value">2 due</div></a></nav>
  <section class="card" id="calendar">
    <a class="head-link" href="calendar">Month →</a>
    <a href="https://google.com/maps/?q=Berlin" target="_blank" rel="noopener">Berlin</a>
    <details class="more"><summary>All details</summary><table></table></details>
  </section>
</div>
<script>document.body.classList.add("x")</script>
</body></html>`

const testCSS = `.tile { display: inline-block; width: 30%; } .card { border: 1px solid #e4e7eb; }`

func TestPrepareEmail(t *testing.T) {
	email, err := PrepareEmail([]byte(testPage), testCSS, "https://aether.example.org/home/")
	if err != nil {
		t.Fatal(err)
	}

	for _, unwanted := range []string{"<script", "htmx.min.js", "hx-get", "hx-trigger", "<button", `class="live"`, "<details", "display: grid"} {
		if strings.Contains(email, unwanted) {
			t.Errorf("expected %q to be removed: %s", unwanted, email)
		}
	}

	for _, wanted := range []string{
		// styles are inlined
		`style="display:inline-block;width:30%"`,
		`style="border:1px solid #e4e7eb"`,
		// relative links point to the dashboard
		`href="https://aether.example.org/home/calendar"`,
		`href="https://aether.example.org/home/#taskwarrior"`,
		// absolute links are kept
		`href="https://google.com/maps/?q=Berlin"`,
		"Updated 10:05",
	} {
		if !strings.Contains(email, wanted) {
			t.Errorf("expected %q in the email: %s", wanted, email)
		}
	}
}

func TestPrepareEmail_withoutDashboardURL(t *testing.T) {
	email, err := PrepareEmail([]byte(testPage), testCSS, "")
	if err != nil {
		t.Fatal(err)
	}

	// links that can't be followed from an email are removed, their content is kept
	if strings.Contains(email, `href="calendar"`) || strings.Contains(email, `href="#taskwarrior"`) {
		t.Errorf("expected relative links to be removed: %s", email)
	}
	if !strings.Contains(email, `<div class="tile" style="display:inline-block;width:30%"><div class="tile-value">Standup</div></div>`) {
		t.Errorf("expected the tile to be kept without link: %s", email)
	}
	if !strings.Contains(email, `<span class="head-link">Month →</span>`) {
		t.Errorf("expected the link text to be kept: %s", email)
	}
	if !strings.Contains(email, `href="https://google.com/maps/?q=Berlin"`) {
		t.Errorf("expected absolute links to be kept: %s", email)
	}
}

func TestPrepareEmail_invalidDashboardURL(t *testing.T) {
	if _, err := PrepareEmail([]byte(testPage), testCSS, "aether.example.org"); err == nil {
		t.Error("expected an error for a dashboard url without scheme")
	}
}
