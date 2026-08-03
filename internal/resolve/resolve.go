package resolve

import (
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/appautomaton/markmaton/internal/safeurl"
)

type imageSource struct {
	URL  string
	Size float64
}

func ResolveHTML(html string, pageURL string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	baseURL := effectiveBaseURL(doc, pageURL)

	doc.Find("img[srcset], source[srcset]").Each(func(_ int, selection *goquery.Selection) {
		srcset, exists := selection.Attr("srcset")
		if !exists {
			return
		}
		selection.RemoveAttr("srcset")
		picked, ok := pickLargestSource(srcset)
		if !ok || !safeurl.Resource(picked) {
			return
		}
		if abs, ok := absolutize(picked, baseURL); ok {
			selection.SetAttr("src", abs)
		} else {
			selection.SetAttr("src", strings.TrimSpace(picked))
		}
	})

	doc.Find("a[href]").Each(func(_ int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if !exists {
			return
		}
		if !safeurl.Link(href) {
			selection.RemoveAttr("href")
			return
		}
		if abs, ok := absolutize(href, baseURL); ok {
			selection.SetAttr("href", abs)
		} else {
			selection.SetAttr("href", strings.TrimSpace(href))
		}
	})

	doc.Find("img[src], source[src], iframe[src], video[src], audio[src]").Each(func(_ int, selection *goquery.Selection) {
		src, exists := selection.Attr("src")
		if !exists {
			return
		}
		if !safeurl.Resource(src) {
			selection.RemoveAttr("src")
			return
		}
		if abs, ok := absolutize(src, baseURL); ok {
			selection.SetAttr("src", abs)
		} else {
			selection.SetAttr("src", strings.TrimSpace(src))
		}
	})

	if body := doc.Find("body").First(); body.Length() > 0 {
		return body.Html()
	}

	return doc.Html()
}

func effectiveBaseURL(doc *goquery.Document, pageURL string) *url.URL {
	pageURL = strings.TrimSpace(pageURL)
	if pageURL == "" {
		return nil
	}

	parsedPageURL, err := url.Parse(pageURL)
	if err != nil || !safeurl.WebBase(parsedPageURL.String()) {
		return nil
	}

	if baseHref, exists := doc.Find("base[href]").First().Attr("href"); exists {
		if parsedBase, err := parsedPageURL.Parse(baseHref); err == nil && safeurl.WebBase(parsedBase.String()) {
			return parsedBase
		}
	}

	return parsedPageURL
}

func absolutize(value string, base *url.URL) (string, bool) {
	if base == nil {
		return value, false
	}
	if strings.TrimSpace(value) == "" {
		return value, false
	}

	parsed, err := base.Parse(value)
	if err != nil {
		return value, false
	}

	return parsed.String(), true
}

func pickLargestSource(srcset string) (string, bool) {
	var sources []imageSource

	for _, part := range strings.Split(srcset, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		tokens := strings.Fields(part)
		if len(tokens) == 0 {
			continue
		}

		if !safeurl.Resource(tokens[0]) {
			continue
		}
		candidate := imageSource{URL: tokens[0], Size: 1}
		if len(tokens) > 1 {
			descriptor := tokens[len(tokens)-1]
			if strings.HasSuffix(descriptor, "x") || strings.HasSuffix(descriptor, "w") {
				if size, err := strconv.ParseFloat(descriptor[:len(descriptor)-1], 64); err == nil {
					candidate.Size = size
				}
			}
		}

		sources = append(sources, candidate)
	}

	if len(sources) == 0 {
		return "", false
	}

	sort.SliceStable(sources, func(i, j int) bool {
		return sources[i].Size > sources[j].Size
	})

	return sources[0].URL, true
}
