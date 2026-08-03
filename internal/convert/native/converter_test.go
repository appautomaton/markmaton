package native

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestConverterConvertsCoreBlocksAndInlines(t *testing.T) {
	input := `<article>
		<h1>Native Title</h1>
		<p>Hello <strong>world</strong> and <em>friends</em>. <a href="https://example.com/docs" title="Docs">Read docs</a>.</p>
		<p><img src="https://example.com/cover.png" alt="Cover"></p>
	</article>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	want := strings.Join([]string{
		"# Native Title",
		"",
		`Hello **world** and _friends_. [Read docs](https://example.com/docs "Docs").`,
		"",
		`![Cover](https://example.com/cover.png)`,
	}, "\n")
	if got != want {
		t.Fatalf("unexpected markdown\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterPreservesUnicodeAndNestedLists(t *testing.T) {
	input := `<ol start="3"><li>你好 <strong>世界</strong><ul><li>嵌套项目</li></ul></li><li>完成</li></ol>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	for _, expected := range []string{
		"3. 你好 **世界**",
		"   - 嵌套项目",
		"4. 完成",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in:\n%s", expected, got)
		}
	}
}

func TestConverterNormalizesNestedFormattingAndBoundaryWhitespace(t *testing.T) {
	input := `<p>Some<strong> Text. </strong>Content</p>
	<p><strong><strong>Text</strong></strong></p>
	<p><em><i>Double</i>Italic</em></p>
	<p>首付<em><i>19,8</i>万</em> / 月供<em>6339元X24</em></p>
	<p><em>Content </em>and no space afterward.</p>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	want := strings.Join([]string{
		"Some **Text.** Content",
		"",
		"**Text**",
		"",
		"_DoubleItalic_",
		"",
		"首付*19,8万* / 月供*6339元X24*",
		"",
		"_Content_ and no space afterward.",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected formatting output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterRendersSafeLinksImagesAndLiteralHTMLText(t *testing.T) {
	input := `<p><a href="https://example.com/a b" title=" A title "> Link </a><a href="https://example.com/image"><img src="https://example.com/a.png" alt=" A [cover] "></a></p>
	<p>The &lt;img&gt; tag is literal.</p>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	for _, expected := range []string{
		`[Link](https://example.com/a%20b "A title") [![A \[cover\]](https://example.com/a.png)](https://example.com/image)`,
		`The \<img\> tag is literal.`,
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in:\n%s", expected, got)
		}
	}
}

func TestConverterPreservesMeaningfulIframeLinksWithoutTrackingNoise(t *testing.T) {
	input := `<p><iframe src="https://www.youtube.com/embed/video-id"></iframe></p>
	<p><iframe src="https://widgets.example.com/map" title="Live map"></iframe></p>
	<p><iframe src="about:blank" title="Tracking frame"></iframe></p>
	<p><iframe src="https://widgets.example.com/untitled"></iframe></p>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert iframes: %v", err)
	}
	want := "[Embedded video](https://www.youtube.com/embed/video-id)\n\n[Live map](https://widgets.example.com/map)"
	if got != want {
		t.Fatalf("unexpected iframe output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterPreservesSafeMediaAndDropsUnsafeDestinations(t *testing.T) {
	input := `<p><a href="javascript:alert(1)">Run</a> <a href="mailto:reader@example.com">Mail</a></p>
	<p><img src="data:image/svg+xml,bad" alt="Unsafe image"></p>
	<p><video src="https://cdn.example.com/demo.mp4" title="Product demo"></video></p>
	<p><audio><source src="https://cdn.example.com/episode.mp3"></audio></p>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert safe media: %v", err)
	}
	want := "Run [Mail](mailto:reader@example.com)\n\nUnsafe image\n\n[Product demo](https://cdn.example.com/demo.mp4)\n\n[Audio](https://cdn.example.com/episode.mp3)"
	if got != want {
		t.Fatalf("unexpected safe-media output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterSplitsFormattingAcrossHardBreaks(t *testing.T) {
	input := `<p><em>First <br> Second</em> and <strong>Third<br>Fourth</strong></p>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	want := "_First_ \\\n _Second_ and **Third**\\\n**Fourth**"
	if got != want {
		t.Fatalf("unexpected hard-break formatting\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterAvoidsRedundantSameTargetImageLinks(t *testing.T) {
	input := `<p><a href="https://example.com/image.png"><img src="https://example.com/image.png" alt="Preview"></a></p>
	<p><a href="https://example.com/image.png" title="Open image"><img src="https://example.com/image.png" alt="Preview"></a></p>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert linked images: %v", err)
	}
	want := "![Preview](https://example.com/image.png)\n\n[![Preview](https://example.com/image.png)](https://example.com/image.png \"Open image\")"
	if got != want {
		t.Fatalf("unexpected linked-image output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterChoosesSafeInlineAndBlockCodeFences(t *testing.T) {
	input := `<p>Use <code>value ` + "```" + ` here</code> and <code>  spaced  value  </code>.</p><pre><code class="language-go">fmt.Println("` + "```" + `")</code></pre>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	if !strings.Contains(got, "````value ``` here````") {
		t.Fatalf("expected four-backtick inline fence, got:\n%s", got)
	}
	if !strings.Contains(got, "`   spaced  value   `") {
		t.Fatalf("expected inline code spacing to be preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "````go\n") || !strings.Contains(got, "\n````") {
		t.Fatalf("expected safe fenced code block, got:\n%s", got)
	}
}

func TestConverterStripsCodeGuttersAndRendersTables(t *testing.T) {
	input := `<pre><code class="language-js"><table><tr><td class="gutter">1</td><td><div>const x = 1;</div><div>const y = 2;</div></td></tr></table></code></pre>
	<pre><code class="line-numbers language-css">body { color: red; }</code></pre>
	<table><thead><tr><th>Name</th><th>Value</th></tr></thead><tbody><tr><td>A</td><td>1</td></tr></tbody></table>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	if strings.Contains(got, "\n1\nconst") {
		t.Fatalf("expected code gutter to be removed, got:\n%s", got)
	}
	for _, expected := range []string{
		"```js",
		"const x = 1;",
		"const y = 2;",
		"```css",
		"body { color: red; }",
		"| Name | Value |",
		"| --- | --- |",
		"| A | 1 |",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in:\n%s", expected, got)
		}
	}
}

func TestConverterEscapesAccidentalMarkdownStructure(t *testing.T) {
	input := `<p>1. first<br>2. second</p><p>- not a list</p><p>---</p><p>~~not deleted~~ and ~single~</p>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	for _, expected := range []string{
		"1\\. first\\\n2\\. second",
		"\\- not a list",
		"\\---",
		`\~\~not deleted\~\~ and ~single~`,
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in:\n%s", expected, got)
		}
	}
}

func TestConverterHonorsCancellationAndLimits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := New().Convert(ctx, "<p>Hello</p>"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}

	limited := NewWithLimits(Limits{MaxInputBytes: 4, MaxNodes: 100, MaxDepth: 10})
	if _, err := limited.Convert(context.Background(), "<p>Hello</p>"); !errors.Is(err, ErrInputLimit) {
		t.Fatalf("expected input limit error, got %v", err)
	}
}
