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

func TestDefaultBuilderNamesAreDeterministic(t *testing.T) {
	left := DefaultBuilder("")
	right := DefaultBuilder("")

	if strings.Join(left.PluginNames(), ",") != strings.Join(right.PluginNames(), ",") {
		t.Fatalf("expected default plugin names to be deterministic")
	}
	if strings.Join(left.BeforeHookNames(), ",") != strings.Join(right.BeforeHookNames(), ",") {
		t.Fatalf("expected default before hook names to be deterministic")
	}
	if strings.Join(left.AfterHookNames(), ",") != strings.Join(right.AfterHookNames(), ",") {
		t.Fatalf("expected default after hook names to be deterministic")
	}
}
