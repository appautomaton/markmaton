package metadata

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/appautomaton/markmaton/internal/model"
)

func Extract(html string) (model.Metadata, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return model.Metadata{}, err
	}

	meta := model.Metadata{
		Extras: map[string]string{},
	}

	meta.Title = normalizeText(doc.Find("title").First().Text())
	meta.Description = normalizeText(content(doc, `meta[name="description"]`))
	meta.CanonicalURL = attr(doc, `link[rel="canonical"]`, "href")
	meta.Language = attr(doc, "html[lang]", "lang")
	meta.Author = normalizeText(content(doc, `meta[name="author"]`))
	meta.OGTitle = normalizeText(content(doc, `meta[property="og:title"]`))
	meta.OGDescription = normalizeText(content(doc, `meta[property="og:description"]`))

	if meta.Title == "" {
		meta.Title = meta.OGTitle
	}

	if meta.Extras != nil && len(meta.Extras) == 0 {
		meta.Extras = nil
	}

	return meta, nil
}

func content(doc *goquery.Document, selector string) string {
	return strings.TrimSpace(attr(doc, selector, "content"))
}

func attr(doc *goquery.Document, selector, name string) string {
	value, _ := doc.Find(selector).First().Attr(name)
	return strings.TrimSpace(value)
}

func normalizeText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
