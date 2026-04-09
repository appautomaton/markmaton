package metadata

import (
	"testing"

	"github.com/appautomaton/markmaton/internal/testutil"
)

func TestExtractMetadata(t *testing.T) {
	data := testutil.ReadFixture(t, "core/article.html")

	meta, err := Extract(data)
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

func TestExtractNormalizesTitleWhitespace(t *testing.T) {
	meta, err := Extract(`<html><head><title>Abuse Investigator  @ OpenAI</title></head><body></body></html>`)
	if err != nil {
		t.Fatalf("extract metadata: %v", err)
	}

	if meta.Title != "Abuse Investigator @ OpenAI" {
		t.Fatalf("unexpected normalized title: %q", meta.Title)
	}
}
