package native

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNilConverterReturnsAnError(t *testing.T) {
	var converter *Converter
	if _, err := converter.Convert(context.Background(), `<p>Hello</p>`); err == nil {
		t.Fatal("expected a nil converter error")
	}
}

func TestConverterEnforcesOutputLimit(t *testing.T) {
	converter := NewWithLimits(Limits{MaxOutputBytes: 4})
	_, err := converter.Convert(context.Background(), `<p>12345</p>`)
	if !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("expected output limit error, got %v", err)
	}
}

func TestConverterCountsDOMNodesAndDepthPrecisely(t *testing.T) {
	input := `<p>Hello</p>`

	if _, err := NewWithLimits(Limits{MaxNodes: 5}).Convert(context.Background(), input); !errors.Is(err, ErrNodeLimit) {
		t.Fatalf("expected exact DOM node limit error, got %v", err)
	}
	if _, err := NewWithLimits(Limits{MaxNodes: 6}).Convert(context.Background(), input); err != nil {
		t.Fatalf("expected six-node DOM to fit the limit, got %v", err)
	}

	if _, err := NewWithLimits(Limits{MaxDepth: 3}).Convert(context.Background(), input); !errors.Is(err, ErrDepthLimit) {
		t.Fatalf("expected exact DOM depth limit error, got %v", err)
	}
	if _, err := NewWithLimits(Limits{MaxDepth: 4}).Convert(context.Background(), input); err != nil {
		t.Fatalf("expected depth-four DOM to fit the limit, got %v", err)
	}
}

func TestConverterHandlesMalformedHTMLWithoutPanicking(t *testing.T) {
	inputs := []string{
		`<div><li>Detached item</div>`,
		`<p><strong>Unclosed`,
		`<table><tr><td>A<tr><td>B`,
		`<blockquote><ol><li>One<li><pre><code>value`,
		`<a href="https://example.com"><h2>Linked heading</a><p>Body`,
		strings.Repeat(`<div>`, 64) + "content",
	}

	for index, input := range inputs {
		t.Run(strings.Repeat("case_", 1)+string(rune('a'+index)), func(t *testing.T) {
			markdown, err := New().Convert(context.Background(), input)
			if err != nil {
				t.Fatalf("convert malformed HTML: %v", err)
			}
			if strings.TrimSpace(markdown) == "" {
				t.Fatalf("expected readable output for malformed HTML %q", input)
			}
		})
	}
}

func FuzzConverterNeverPanics(f *testing.F) {
	for _, seed := range []string{
		`<p>Hello <strong>world</strong>.</p>`,
		`<ul><li>One<ul><li>Nested</li></ul></li></ul>`,
		`<table><tr><td colspan="2">A</td></tr></table>`,
		"<pre><code>```\nvalue</code></pre>",
		`<div><li>Malformed</div>`,
		"\x00<>&\xff",
	} {
		f.Add(seed)
	}

	converter := NewWithLimits(Limits{
		MaxInputBytes:  1 << 20,
		MaxOutputBytes: 2 << 20,
		MaxNodes:       50_000,
		MaxDepth:       512,
	})
	f.Fuzz(func(t *testing.T, input string) {
		first, firstErr := converter.Convert(context.Background(), input)
		second, secondErr := converter.Convert(context.Background(), input)
		if (firstErr == nil) != (secondErr == nil) {
			t.Fatalf("conversion error is not deterministic: %v != %v", firstErr, secondErr)
		}
		if firstErr == nil && first != second {
			t.Fatal("conversion output is not deterministic")
		}
	})
}
