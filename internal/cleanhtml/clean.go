package cleanhtml

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var defaultMainContentExcludeSelectors = []string{
	"header",
	"footer",
	"nav",
	"aside",
	".header",
	".footer",
	".sidebar",
	".nav",
	".navbar",
	".menu",
	".breadcrumb",
	".breadcrumbs",
	".social",
	".share",
	".cookie",
	".modal",
	".popup",
	".advert",
	".ads",
}

func Clean(html string, onlyMainContent bool, includeSelectors, excludeSelectors []string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	if len(includeSelectors) > 0 {
		var selected []string
		for _, selector := range includeSelectors {
			selector = strings.TrimSpace(selector)
			if selector == "" {
				continue
			}
			doc.Find(selector).Each(func(_ int, selection *goquery.Selection) {
				if outer, outerErr := goquery.OuterHtml(selection); outerErr == nil {
					selected = append(selected, outer)
				}
			})
		}
		doc, err = goquery.NewDocumentFromReader(strings.NewReader(strings.Join(selected, "\n")))
		if err != nil {
			return "", err
		}
	}

	doc.Find("head, meta, noscript, style, script").Each(func(_ int, s *goquery.Selection) {
		s.Remove()
	})

	for _, selector := range excludeSelectors {
		selector = strings.TrimSpace(selector)
		if selector == "" {
			continue
		}
		doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
			s.Remove()
		})
	}

	if onlyMainContent {
		for _, selector := range defaultMainContentExcludeSelectors {
			doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
				s.Remove()
			})
		}
	}

	if body := doc.Find("body").First(); body.Length() > 0 {
		cleaned, err := body.Html()
		if err != nil {
			return "", err
		}
		return cleaned, nil
	}

	return doc.Html()
}
