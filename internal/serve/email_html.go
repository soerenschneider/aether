package serve

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/vanng822/go-premailer/premailer"
)

// PrepareEmail turns the rendered dashboard into html that email clients can display. Email clients support little
// CSS and often drop <style> elements, so the dashboard's styles are replaced by the given, simpler css, which is
// inlined into the elements. Parts that need a browser are removed. Relative links point to the dashboard at
// dashboardURL, or are removed if dashboardURL is empty, as they can't be followed from an email.
func PrepareEmail(page []byte, css, dashboardURL string) (string, error) {
	var base *url.URL
	if dashboardURL != "" {
		var err error
		base, err = url.Parse(dashboardURL)
		if err != nil || !base.IsAbs() {
			return "", fmt.Errorf("invalid dashboard url %q", dashboardURL)
		}
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(page))
	if err != nil {
		return "", err
	}

	// interactive parts and the full weather table, which only make sense in a browser
	doc.Find("script, noscript, button, .live, details.more, link").Remove()
	doc.Find("style").Remove()
	doc.Find("head").AppendHtml("<style>" + css + "</style>")

	doc.Find("*").Each(func(_ int, s *goquery.Selection) {
		node := s.Nodes[0]
		attrs := node.Attr[:0]
		for _, attr := range node.Attr {
			if !strings.HasPrefix(attr.Key, "hx-") {
				attrs = append(attrs, attr)
			}
		}
		node.Attr = attrs
	})

	doc.Find("a[href]").Each(func(_ int, link *goquery.Selection) {
		href, _ := link.Attr("href")
		target, err := url.Parse(href)
		if err == nil && target.IsAbs() {
			return
		}
		if err == nil && base != nil {
			link.SetAttr("href", base.ResolveReference(target).String())
			return
		}

		// nowhere to link to: keep the content, but not the link
		link.RemoveAttr("href")
		link.RemoveAttr("target")
		link.RemoveAttr("rel")
		if link.HasClass("tile") || link.HasClass("glance-item") {
			link.Nodes[0].Data = "div"
		} else {
			link.Nodes[0].Data = "span"
		}
	})

	html, err := doc.Html()
	if err != nil {
		return "", err
	}

	inliner, err := premailer.NewPremailerFromString(html, premailer.NewOptions())
	if err != nil {
		return "", err
	}
	return inliner.Transform()
}
