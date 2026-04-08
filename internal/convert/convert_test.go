package convert

import (
	"strings"
	"testing"
)

func TestToMarkdownConvertsRichHTML(t *testing.T) {
	html := `<article><h1>Title</h1><p>Hello <strong>world</strong>.</p><pre><code class="language-python">print("hi")</code></pre></article>`
	markdown, err := ToMarkdown(html)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	expected := []string{
		"# Title",
		"Hello **world**.",
		"```python",
		`print("hi")`,
	}
	for _, part := range expected {
		if !strings.Contains(markdown, part) {
			t.Fatalf("expected markdown to contain %q, got:\n%s", part, markdown)
		}
	}
}
