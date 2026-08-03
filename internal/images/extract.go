package images

import (
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/appautomaton/markmaton/internal/safeurl"
)

func Extract(html string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	var out []string
	doc.Find("img[src], picture source[src]").Each(func(_ int, selection *goquery.Selection) {
		src, exists := selection.Attr("src")
		if !exists {
			return
		}
		src = strings.TrimSpace(src)
		if !safeurl.Resource(src) {
			return
		}
		if _, exists := seen[src]; exists {
			return
		}
		seen[src] = struct{}{}
		out = append(out, src)
	})

	return out, nil
}
