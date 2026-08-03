package native

import (
	"context"
	"strings"
	"testing"
)

func TestConverterRendersGitHubTaskLists(t *testing.T) {
	input := `<ul>
	<li><input type="checkbox" checked>Checked<ul><li><input type="checkbox">Nested</li></ul></li>
	<li><input type="checkbox">Unchecked</li>
	</ul>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert task list: %v", err)
	}
	want := "- [x] Checked\n  - [ ] Nested\n- [ ] Unchecked"
	if got != want {
		t.Fatalf("unexpected task-list output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterRendersAlignedCaptionedAndRaggedTables(t *testing.T) {
	input := `<table>
	<caption>Inventory</caption>
	<thead><tr><th align="left">Name</th><th style="text-align: center">Value</th><th align="right">Total</th></tr></thead>
	<tbody><tr><td>A</td><td>1<br>2</td><td>3</td></tr><tr><td>B</td></tr></tbody>
	</table>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert aligned table: %v", err)
	}
	want := strings.Join([]string{
		"Inventory",
		"",
		"| Name | Value | Total |",
		"| :-- | :-: | --: |",
		"| A | 1<br>2 | 3 |",
		"| B |  |  |",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected aligned-table output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterRendersHeaderlessTableWithSyntheticHeader(t *testing.T) {
	input := `<table><tr><td>A</td><td>B</td></tr><tr><td>C</td></tr></table>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert headerless table: %v", err)
	}
	want := "|  |  |\n| --- | --- |\n| A | B |\n| C |  |"
	if got != want {
		t.Fatalf("unexpected headerless-table output\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestConverterFallsBackForSpanningAndNestedTablesWithoutLosingText(t *testing.T) {
	input := `<table><caption>Complex report</caption><tr><td colspan="2">Combined</td><td rowspan="2"><table><tr><td>Nested value</td></tr></table></td></tr><tr><td>Final</td></tr></table>`

	got, err := New().Convert(context.Background(), input)
	if err != nil {
		t.Fatalf("convert complex table: %v", err)
	}
	for _, expected := range []string{"Complex report", "Combined", "Nested value", "Final"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in fallback output:\n%s", expected, got)
		}
	}
	if strings.Contains(got, "| ---") {
		t.Fatalf("complex tables must use readable fallback output:\n%s", got)
	}
}
