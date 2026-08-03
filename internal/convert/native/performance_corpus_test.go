package native

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/appautomaton/markmaton/internal/testutil"
)

func TestLargePerformanceCorpus(t *testing.T) {
	if os.Getenv("MARKMATON_RUN_LARGE_CORPUS") != "1" {
		t.Skip("set MARKMATON_RUN_LARGE_CORPUS=1 to run large performance fixtures")
	}

	tests := []struct {
		name           string
		fixture        string
		minimumOutput  int
		contentSignals []string
	}{
		{name: "mixed document", fixture: "performance/mixed.html", minimumOutput: 1_000_000, contentSignals: []string{`TEST\_e0cd13a8`, `TEST\_9cc3d91c TEST\_c2f80630`}},
		{name: "large list", fixture: "performance/large-list.html", minimumOutput: 500_000, contentSignals: []string{`1. [x](/x "x")`, `39229. [x](/x "x")`, "## x"}},
		{name: "large table", fixture: "performance/large-table.html", minimumOutput: 500_000, contentSignals: []string{"ACME INC \\#1", "ACME INC \\#5000"}},
	}

	converter := New()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := testutil.ReadFixture(t, test.fixture)
			markdown, err := converter.Convert(context.Background(), input)
			if err != nil {
				t.Fatalf("convert performance fixture: %v", err)
			}
			if len(markdown) < test.minimumOutput {
				t.Fatalf("expected at least %d output bytes, got %d", test.minimumOutput, len(markdown))
			}
			for _, signal := range test.contentSignals {
				if !strings.Contains(markdown, signal) {
					t.Fatalf("expected %q in performance output", signal)
				}
			}
		})
	}
}
