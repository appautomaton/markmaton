package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appautomaton/markmaton/internal/model"
)

func TestProcessMatchesGoldenFixtures(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		fixture string
		golden  string
	}{
		{
			name:    "article",
			url:     "https://example.com/articles/harnessing-parsers",
			fixture: "article.html",
			golden:  "article.md",
		},
		{
			name:    "docs",
			url:     "https://example.com/docs/setup",
			fixture: "docs.html",
			golden:  "docs.md",
		},
		{
			name:    "news",
			url:     "https://example.com/news",
			fixture: "news.html",
			golden:  "news.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			expected := loadGolden(t, tc.golden)

			response, err := Process(model.Request{
				URL:  tc.url,
				HTML: fixture,
				Options: model.Options{
					OnlyMainContent: true,
				},
			})
			if err != nil {
				t.Fatalf("process failed: %v", err)
			}

			if strings.TrimSpace(response.Markdown) != strings.TrimSpace(expected) {
				t.Fatalf("markdown mismatch\nexpected:\n%s\n\ngot:\n%s", expected, response.Markdown)
			}
		})
	}
}

func TestProcessFallsBackToFullContent(t *testing.T) {
	html := `<html><body><aside><p>This is the actual content.</p><p>It lives in an aside.</p></aside></body></html>`
	response, err := Process(model.Request{
		URL:  "https://example.com/weird-layout",
		HTML: html,
		Options: model.Options{
			OnlyMainContent: true,
		},
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	if !response.Quality.FallbackUsed {
		t.Fatalf("expected fallback to be used")
	}
	if response.Markdown == "" {
		t.Fatalf("expected markdown after fallback")
	}
}

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "fixtures", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return string(data)
}

func loadGolden(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "golden", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	return string(data)
}
