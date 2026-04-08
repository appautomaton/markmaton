package cleanhtml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanRemovesLayoutNoise(t *testing.T) {
	html := loadFixture(t, "article.html")
	cleaned, err := Clean(html, true, nil, nil)
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	if strings.Contains(cleaned, "Home") {
		t.Fatalf("expected nav content to be removed")
	}
	if strings.Contains(cleaned, "Copyright Example") {
		t.Fatalf("expected footer content to be removed")
	}
	if !strings.Contains(cleaned, "Harnessing Parsers") {
		t.Fatalf("expected article content to remain")
	}
}

func TestCleanHonorsIncludeAndExcludeSelectors(t *testing.T) {
	html := loadFixture(t, "docs.html")
	cleaned, err := Clean(html, true, []string{"main"}, []string{"ul"})
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	if strings.Contains(cleaned, "Sidebar links") {
		t.Fatalf("expected sidebar content to be dropped")
	}
	if strings.Contains(cleaned, "Install the package") {
		t.Fatalf("expected excluded selector content to be removed")
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
