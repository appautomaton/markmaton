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

func TestExtractDropsUnsafeLinks(t *testing.T) {
	html := `<a href="javascript:alert(1)">Run</a><a href="data:text/html,hello">Data</a><a href="mailto:reader@example.com">Mail</a>`
	links, err := Extract(html)
	if err != nil {
		t.Fatalf("extract links: %v", err)
	}
	if len(links) != 1 || links[0] != "mailto:reader@example.com" {
		t.Fatalf("unexpected safe links: %v", links)
	}
}
