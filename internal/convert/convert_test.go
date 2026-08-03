package convert

import (
	"context"
	"errors"
	"strings"
	"sync"
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

func TestToMarkdownPreservesMeaningfulStandaloneEmbed(t *testing.T) {
	markdown, err := ToMarkdown(`<iframe src="https://www.youtube.com/embed/video-id"></iframe>`)
	if err != nil {
		t.Fatalf("convert embed: %v", err)
	}
	want := "[Embedded video](https://www.youtube.com/embed/video-id)"
	if markdown != want {
		t.Fatalf("unexpected embed output\nwant: %s\ngot:  %s", want, markdown)
	}
}

func TestToMarkdownUsesCanonicalNativeOutput(t *testing.T) {
	markdown, err := ToMarkdown(`<p>Before</p><hr><p>After</p>`)
	if err != nil {
		t.Fatalf("convert with Native converter: %v", err)
	}
	if !strings.Contains(markdown, "\n\n---\n\n") {
		t.Fatalf("expected canonical Native thematic break:\n%s", markdown)
	}
}

func TestToMarkdownContextRejectsNilContext(t *testing.T) {
	if _, err := ToMarkdownContext(nil, `<p>Hello</p>`); err == nil {
		t.Fatal("expected nil context error")
	}
}

func TestToMarkdownContextHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ToMarkdownContext(ctx, `<p>Hello</p>`)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestNativeConversionSupportsConcurrentUse(t *testing.T) {
	const workers = 24
	input := `<article><h1>Concurrent</h1><p>Hello <strong>world</strong>.</p></article>`

	var wait sync.WaitGroup
	failures := make(chan error, workers)
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			markdown, err := ToMarkdown(input)
			if err != nil {
				failures <- err
				return
			}
			if !strings.Contains(markdown, "# Concurrent") || !strings.Contains(markdown, "Hello **world**.") {
				failures <- errors.New("concurrent conversion returned unexpected output")
			}
		}()
	}
	wait.Wait()
	close(failures)
	for err := range failures {
		t.Fatal(err)
	}
}
