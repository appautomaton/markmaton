package links

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Extract(html string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	var out []string
	doc.Find("a[href]").Each(func(_ int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if !exists {
			return
		}
		href = strings.TrimSpace(href)
		if href == "" {
			return
		}
		if _, exists := seen[href]; exists {
			return
		}
		seen[href] = struct{}{}
		out = append(out, href)
	})

	return out, nil
}
