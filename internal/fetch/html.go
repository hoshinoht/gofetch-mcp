package fetch

import (
	"bytes"
	"net/url"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/PuerkitoBio/goquery"
)

// mainContentSelectors are tried in order; the first non-empty match is
// treated as the article body.
var mainContentSelectors = []string{"article", "main", "[role=main]", "#content", ".content", "body"}

// strippedSelectors is page chrome removed before conversion.
const strippedSelectors = "script, style, nav, header, footer, aside, form, noscript, iframe, svg, [role=navigation], [role=banner], [aria-hidden=true]"

func convertHTML(body []byte, pageURL string) (markdown, title string, err error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	title = strings.TrimSpace(doc.Find("title").First().Text())

	var sel *goquery.Selection
	for _, selector := range mainContentSelectors {
		if s := doc.Find(selector).First(); s.Length() > 0 {
			sel = s
			break
		}
	}
	if sel == nil {
		sel = doc.Selection
	}
	sel.Find(strippedSelectors).Remove()

	fragment, err := goquery.OuterHtml(sel)
	if err != nil {
		return "", "", err
	}

	opts := []converter.ConvertOptionFunc{}
	if domain := domainOf(pageURL); domain != "" {
		opts = append(opts, converter.WithDomain(domain))
	}

	md, err := htmltomarkdown.ConvertString(fragment, opts...)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(md), title, nil
}

func domainOf(pageURL string) string {
	u, err := url.Parse(pageURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}
