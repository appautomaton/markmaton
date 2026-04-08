package links

import "testing"

func TestExtractDeduplicatesLinks(t *testing.T) {
	html := `<a href="https://example.com/docs">Docs</a><a href="https://example.com/docs">Docs again</a>`
	links, err := Extract(html)
	if err != nil {
		t.Fatalf("extract links: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected one unique link, got %d", len(links))
	}
}
