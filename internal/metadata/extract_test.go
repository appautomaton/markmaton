package metadata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractMetadata(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "fixtures", "article.html")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	meta, err := Extract(string(data))
	if err != nil {
		t.Fatalf("extract metadata: %v", err)
	}

	if meta.Title != "Harnessing Parsers" {
		t.Fatalf("unexpected title: %q", meta.Title)
	}
	if meta.Description != "A practical note on parser design." {
		t.Fatalf("unexpected description: %q", meta.Description)
	}
	if meta.CanonicalURL != "https://example.com/articles/harnessing-parsers" {
		t.Fatalf("unexpected canonical URL: %q", meta.CanonicalURL)
	}
}
