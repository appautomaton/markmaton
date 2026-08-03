package native

import (
	"context"
	"strings"
	"testing"
)

func TestConverterPreservesBlocksInsideCustomElementWrappers(t *testing.T) {
	input := `<react-app><issue-view><h1>Issue Title</h1><section><p>Issue body.</p></section></issue-view></react-app>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert custom element wrappers: %v", err)
	}
	want := "# Issue Title\n\nIssue body."
	if got != want {
		t.Fatalf("unexpected custom-wrapper output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterSeparatesDefinitionListTermsAndDescriptions(t *testing.T) {
	input := `<dl><dt>Term</dt><dd>Definition</dd><dt>Other</dt><dd>Second definition</dd></dl>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert definition list: %v", err)
	}
	want := "Term\n\nDefinition\n\nOther\n\nSecond definition"
	if got != want {
		t.Fatalf("unexpected definition-list output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterPreservesBlockStructureInsideLinks(t *testing.T) {
	input := `<a href="https://example.com/card"><h2>Card<br>Title</h2><p>Card body.</p></a>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert linked blocks: %v", err)
	}
	want := "## [Card Title](https://example.com/card)\n\n[Card body.](https://example.com/card)"
	if got != want {
		t.Fatalf("unexpected linked-block output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterPreservesFencedCodeInsideBlockquotes(t *testing.T) {
	input := `<blockquote><p>Code:</p><pre><code>return 1 &lt; 2;</code></pre></blockquote>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert quoted code: %v", err)
	}
	want := "> Code:\n>\n> ```\n> return 1 < 2;\n> ```"
	if got != want {
		t.Fatalf("unexpected quoted-code output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterNormalizesMalformedNestedListsWithoutEmptyMarkers(t *testing.T) {
	input := `<ul>
	<li>One</li>
	<ul><li>One point one</li><li>One point two</li></ul>
	<li><ul><li>Wrapper nested item</li></ul></li>
	</ul>
	<div><li>Detached item</li></div>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert malformed lists: %v", err)
	}
	for _, expected := range []string{
		"- One\n  - One point one\n  - One point two",
		"  - Wrapper nested item",
		"Detached item",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in:\n%s", expected, got)
		}
	}
	for _, line := range strings.Split(got, "\n") {
		if strings.TrimSpace(line) == "-" {
			t.Fatalf("unexpected empty list marker in:\n%s", got)
		}
	}
}

func TestConverterRendersReversedAndExplicitlyNumberedLists(t *testing.T) {
	input := `<ol reversed><li>Three</li><li value="7">Seven</li><li>Six</li></ol>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert numbered list: %v", err)
	}
	want := "3. Three\n7. Seven\n6. Six"
	if got != want {
		t.Fatalf("unexpected numbered-list output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterRendersListItemsContainingParagraphsAndCode(t *testing.T) {
	input := `<ol start="3"><li><p>First paragraph.</p><p>Second paragraph.</p><pre><code class="language-go">fmt.Println("ok")</code></pre><ul><li>Nested</li></ul></li></ol>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert rich list item: %v", err)
	}
	for _, expected := range []string{
		"3. First paragraph.",
		"   Second paragraph.",
		"   ```go",
		"   fmt.Println(\"ok\")",
		"   - Nested",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in:\n%s", expected, got)
		}
	}
}

func TestConverterPreservesNestedSyntaxHighlighterContent(t *testing.T) {
	inputs := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name: "nested divs",
			html: `<pre><code class="lang-json"><div class="outer"><div class="inner">{
  "status": "success"
}</div></div></code></pre>`,
			expected: []string{"```json", `"status"`, `"success"`},
		},
		{
			name:     "token spans",
			html:     `<pre><code class="language-go"><div class="highlight"><span class="keyword">func</span> <span class="function">main</span>() {}</div></code></pre>`,
			expected: []string{"```go", "func main()"},
		},
		{
			name:     "multiple token lines",
			html:     `<pre><code><div class="token-line">const x = 1;</div><div class="token-line">const y = 2;</div></code></pre>`,
			expected: []string{"const x = 1;\nconst y = 2;"},
		},
	}

	for _, input := range inputs {
		t.Run(input.name, func(t *testing.T) {
			got, err := New().Convert(context.Background(), input.html)
			if err != nil {
				t.Fatalf("convert syntax highlighter: %v", err)
			}
			for _, expected := range input.expected {
				if !strings.Contains(got, expected) {
					t.Fatalf("expected %q in:\n%s", expected, got)
				}
			}
		})
	}
}

func TestConverterDetectsCodeLanguagesFromHighlighterContainers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "data language ancestor", input: `<div data-language="js"><pre><code>const value = 1;</code></pre></div>`, expected: "```js"},
		{name: "brush class", input: `<pre class="brush: html notranslate"><code>&lt;p&gt;Hello&lt;/p&gt;</code></pre>`, expected: "```html"},
		{name: "language class", input: `<pre><code class="language-python">print("hello")</code></pre>`, expected: "```python"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := New().Convert(context.Background(), test.input)
			if err != nil {
				t.Fatalf("convert code block: %v", err)
			}
			if !strings.Contains(got, test.expected) {
				t.Fatalf("expected %q in:\n%s", test.expected, got)
			}
		})
	}
}
