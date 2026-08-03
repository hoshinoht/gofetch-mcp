package fetch

import (
	"strings"
	"testing"
)

func TestConvertHTMLPicksMainContent(t *testing.T) {
	html := `<html><head><title>Example Domain</title></head><body>
<nav>Site nav that should vanish</nav>
<article><h1>Heading</h1><p>Body text with a <a href="/rel">relative link</a>.</p></article>
<footer>Footer junk</footer></body></html>`

	md, title, err := convertHTML([]byte(html), "https://example.com/page")
	if err != nil {
		t.Fatalf("convertHTML: %v", err)
	}
	if title != "Example Domain" {
		t.Errorf("title = %q, want Example Domain", title)
	}
	if !strings.Contains(md, "# Heading") {
		t.Errorf("markdown missing heading: %q", md)
	}
	if strings.Contains(md, "Site nav") || strings.Contains(md, "Footer junk") {
		t.Errorf("chrome not stripped: %q", md)
	}
	if !strings.Contains(md, "https://example.com/rel") {
		t.Errorf("relative link not absolutized: %q", md)
	}
}

func TestFocusContent(t *testing.T) {
	md := "intro block\n\nabout cats here\n\nafter cats\n\nunrelated one\n\nunrelated two\n\nunrelated three"
	got, ok := focusContent(md, "cats")
	if !ok {
		t.Fatal("expected a focus match")
	}
	if !strings.Contains(got, "about cats here") || !strings.Contains(got, "after cats") || !strings.Contains(got, "intro block") {
		t.Errorf("context window wrong: %q", got)
	}
	if strings.Contains(got, "unrelated three") {
		t.Errorf("distant block should be elided: %q", got)
	}
	if !strings.Contains(got, "[…]") {
		t.Errorf("gap marker missing: %q", got)
	}

	if _, ok := focusContent(md, "zebras"); ok {
		t.Error("no-match focus should report ok=false")
	}
}

func TestPaginate(t *testing.T) {
	md := strings.Repeat("é", 100)

	p := paginate(md, 40, 0)
	if p.total != 100 || !p.truncated || p.nextOffset != 40 || len([]rune(p.content)) != 40 {
		t.Errorf("first page wrong: %+v", p)
	}

	p = paginate(md, 40, 80)
	if p.truncated || p.nextOffset != 0 || len([]rune(p.content)) != 20 {
		t.Errorf("last page wrong: %+v", p)
	}

	p = paginate(md, 40, 500)
	if p.content != "" || p.truncated {
		t.Errorf("past-end page wrong: %+v", p)
	}
}

func TestSniffKind(t *testing.T) {
	if got := sniffKind([]byte("%PDF-1.7 ..."), "text/html"); got != "pdf" {
		t.Errorf("magic bytes: got %q", got)
	}
	if got := sniffKind([]byte("<html>"), "application/pdf"); got != "pdf" {
		t.Errorf("content type: got %q", got)
	}
	if got := sniffKind([]byte("<html>"), "text/html; charset=utf-8"); got != "html" {
		t.Errorf("html: got %q", got)
	}
}
