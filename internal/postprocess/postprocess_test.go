package postprocess

import (
	"strings"
	"testing"
)

func TestMarkdownNormalizesBlankLines(t *testing.T) {
	input := "Line 1\n\n\n\nLine 2\n"
	output := Markdown(input)
	expected := "Line 1\n\nLine 2"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestMarkdownKeepsFenceSpacing(t *testing.T) {
	input := "```python\nprint('hi')\n\n\nprint('bye')\n```"
	output := Markdown(input)
	expected := "```python\nprint('hi')\n\n\nprint('bye')\n```"
	if output != expected {
		t.Fatalf("expected fenced code block to stay intact")
	}
}

func TestMarkdownSplitsAdjacentCardLinks(t *testing.T) {
	input := "[![One](https://example.com/one.jpg)](https://example.com/one)[![Two](https://example.com/two.jpg)](https://example.com/two)"
	output := Markdown(input)
	expected := "[![One](https://example.com/one.jpg)](https://example.com/one)\n\n[![Two](https://example.com/two.jpg)](https://example.com/two)"
	if output != expected {
		t.Fatalf("expected adjacent card links to be split, got %q", output)
	}
}

func TestMarkdownRemovesDuplicateStandaloneCardImage(t *testing.T) {
	input := "![One](https://example.com/one.jpg)\n\n[![One](https://example.com/one.jpg)\\\n\\\nFeatured **One**](https://example.com/one)"
	output := Markdown(input)
	expected := "[![One](https://example.com/one.jpg)\\\n\\\nFeatured **One**](https://example.com/one)"
	if output != expected {
		t.Fatalf("expected duplicate standalone image to be removed, got %q", output)
	}
}

func TestMarkdownRemovesGenericListControls(t *testing.T) {
	input := "## Engineering\n\nFilterSortSwitch cards to show MediaSwitch cards to hide Media\n\n[Card](https://example.com/card)"
	output := Markdown(input)

	for _, unwanted := range []string{
		"Filter",
		"Sort",
		"Switch cards to show Media",
		"Switch cards to hide Media",
	} {
		if output == unwanted || containsLine(output, unwanted) {
			t.Fatalf("expected generic list control %q to be removed, got %q", unwanted, output)
		}
	}
	if !containsLine(output, "## Engineering") || !containsLine(output, "[Card](https://example.com/card)") {
		t.Fatalf("expected real content to remain, got %q", output)
	}
}

func TestMarkdownSplitsLabelDateCollisions(t *testing.T) {
	input := "[From model to agent\\\n\\\nEngineeringMar 11, 2026](https://example.com/post)"
	output := Markdown(input)

	expected := "[From model to agent\\\n\\\nEngineering\\\n\\\nMar 11, 2026](https://example.com/post)"
	if output != expected {
		t.Fatalf("expected date collision to be split, got %q", output)
	}
}

func containsLine(markdown, line string) bool {
	for _, current := range strings.Split(markdown, "\n") {
		if current == line {
			return true
		}
	}
	return false
}
