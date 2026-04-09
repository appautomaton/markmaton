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
}

var defaultGlobalExcludeSelectors = []string{
	"dialog",
	"template",
	"[hidden]",
	"[aria-hidden='true']",
	"[role='banner']",
	"[role='dialog']",
	"[role='navigation']",
	"[role='contentinfo']",
	"[role='complementary']",
	"[role='tablist']",
	"[role='tab']",
	"[role='toolbar']",
	"[role='tooltip']",
	"[role='progressbar']",
	"[aria-modal='true']",
	"[role='alert']",
	"[role='status']",
	"[role='search']",
	"a[href*='#skip']",
	"a[aria-label*='Skip']",
	"a[aria-label*='skip']",
	".sr-only",
	".visually-hidden",
	".cookie",
	".modal",
	".popup",
	".advert",
	".ads",
	"[class*='sr-only']",
	"[class*='visually-hidden']",
	".skip-link",
	".skip-to-content",
	"[class*='skip-nav']",
	"[class*='skip-link']",
	"[class*='cookie']",
	"[class*='modal']",
	"[class*='popup']",
	"[class*='overlay']",
	"[class*='popover']",
	"[class*='tooltip']",
	"[class*='loading']",
	"[class*='spinner']",
	"[class*='skeleton']",
	"[class*='shimmer']",
}

var preferredMainContentRootSelectors = []string{
	"article",
	"[role='article']",
	"[itemprop='articleBody']",
	"main",
	"[role='main']",
	"#main",
	".main-content",
	".main",
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

	if onlyMainContent && len(includeSelectors) == 0 {
		doc, err = focusOnPreferredMainRoot(doc)
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
		removeSelector(doc, selector)
	}

	removeSelectors(doc, defaultGlobalExcludeSelectors)

	if onlyMainContent {
		removeSelectors(doc, defaultMainContentExcludeSelectors)
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

func focusOnPreferredMainRoot(doc *goquery.Document) (*goquery.Document, error) {
	for _, selector := range preferredMainContentRootSelectors {
		matchCount := 0
		bestHTML := ""
		bestLength := 0
		doc.Find(selector).Each(func(_ int, selection *goquery.Selection) {
			text := strings.Join(strings.Fields(selection.Text()), " ")
			if strings.TrimSpace(text) == "" {
				return
			}
			matchCount++
			textLength := len([]rune(text))
			if textLength <= bestLength {
				return
			}
			if outer, outerErr := goquery.OuterHtml(selection); outerErr == nil {
				bestHTML = outer
				bestLength = textLength
			}
		})

		if bestHTML == "" {
			continue
		}

		// When a page has many sibling articles or cards, a single article root is
		// often too narrow. In that case, keep looking for a broader main container.
		if matchCount > 1 && strings.Contains(selector, "article") {
			continue
		}

		focused, err := goquery.NewDocumentFromReader(strings.NewReader(bestHTML))
		if err != nil {
			return nil, err
		}
		return focused, nil
	}

	return doc, nil
}

func removeSelectors(doc *goquery.Document, selectors []string) {
	for _, selector := range selectors {
		removeSelector(doc, selector)
	}
}

func removeSelector(doc *goquery.Document, selector string) {
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		s.Remove()
	})
}
