package postprocess

import "testing"

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
