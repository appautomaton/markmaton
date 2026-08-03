package convert

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDropButtonLikeElementsRemovesButtonContent(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<article><button>Copy</button><p>Hello</p></article>`))
	if err != nil {
		t.Fatalf("document parse failed: %v", err)
	}

	removeButtonLikeElements(doc.Selection)
	if strings.Contains(doc.Text(), "Copy") {
		t.Fatalf("expected button text to be removed")
	}
	if !strings.Contains(doc.Text(), "Hello") {
		t.Fatalf("expected content text to remain")
	}
}

func TestTrimOpeningShellControlsDropsLeadingChrome(t *testing.T) {
	markdown := strings.Join([]string{
		"2378",
		"[Timeline](https://example.com/timeline)",
		"",
		"# Real title",
		"",
		"Body starts here.",
	}, "\n")

	got := trimOpeningShellControls()(markdown)
	if strings.HasPrefix(strings.TrimSpace(got), "2378") {
		t.Fatalf("expected opening shell controls to be trimmed, got:\n%s", got)
	}
	if !strings.Contains(got, "# Real title") {
		t.Fatalf("expected real title to remain, got:\n%s", got)
	}
}

func TestTrimOpeningShellControlsDropsSpacedRepositoryChrome(t *testing.T) {
	markdown := strings.Join([]string{
		"[microsoft](https://github.com/microsoft) / **[vscode](https://github.com/microsoft/vscode)** Public",
		"- [Notifications](https://github.com/login)",
		"- [Fork 39k](https://github.com/login)",
		"",
		"# Iteration Plan",
		"",
		"Body starts here.",
	}, "\n")

	got := trimOpeningShellControls()(markdown)
	if firstNonEmpty := nextNonEmptyLine(strings.Split(got, "\n"), 0); firstNonEmpty != "# Iteration Plan" {
		t.Fatalf("expected repository chrome to be trimmed, got:\n%s", got)
	}
}

func TestTrimOpeningShellControlsPreservesMeaningfulOpeningLinks(t *testing.T) {
	markdown := strings.Join([]string{
		"[Embedded video](https://www.youtube.com/watch?v=demo)",
		"",
		"Video description.",
	}, "\n")

	got := trimOpeningShellControls()(markdown)
	if got != markdown {
		t.Fatalf("expected meaningful opening link to remain\nwant:\n%s\n\ngot:\n%s", markdown, got)
	}
}

func TestTrimOpeningShellControlsDropsBrokenMetricLines(t *testing.T) {
	markdown := strings.Join([]string{
		"- [Fork\\",
		"39k](https://example.com/forks)",
		"",
		"# Real title",
		"",
		"Body starts here.",
	}, "\n")

	got := trimOpeningShellControls()(markdown)
	if strings.Contains(got, "Fork\\") || strings.Contains(got, "39k](https://example.com/forks)") {
		t.Fatalf("expected broken metric lines to be trimmed, got:\n%s", got)
	}
	if !strings.Contains(got, "# Real title") {
		t.Fatalf("expected real title to remain, got:\n%s", got)
	}
}

func TestDropStandaloneControlLinesRemovesInteractionLines(t *testing.T) {
	markdown := strings.Join([]string{
		"Follow",
		"Share",
		"Sorted by:",
		"Highest score (default)",
		"",
		"Question body",
	}, "\n")

	got := dropStandaloneControlLines()(markdown)
	for _, unwanted := range []string{"Follow", "Share", "Sorted by:", "Highest score (default)"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("expected %q to be removed, got:\n%s", unwanted, got)
		}
	}
	if !strings.Contains(got, "Question body") {
		t.Fatalf("expected body text to remain, got:\n%s", got)
	}
}

func TestDropStandaloneControlLinesRemovesControlLinks(t *testing.T) {
	markdown := strings.Join([]string{
		"[New issue](https://example.com/new)",
		"[Timeline](https://example.com/timeline)",
		"",
		"Question body",
	}, "\n")

	got := dropStandaloneControlLines()(markdown)
	for _, unwanted := range []string{"[New issue]", "[Timeline]"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("expected %q to be removed, got:\n%s", unwanted, got)
		}
	}
}

func TestDropStandaloneControlLinesRemovesCompositeControls(t *testing.T) {
	markdown := strings.Join([]string{
		"Sorted by: [Reset to default](https://example.com?sort=default)",
		"Highest score (default) Trending (recent votes count more) Date modified (newest first) Date created (oldest first)",
		"",
		"Question body",
	}, "\n")

	got := dropStandaloneControlLines()(markdown)
	for _, unwanted := range []string{"Reset to default", "Highest score", "Trending", "Date modified", "Date created"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("expected composite control %q to be removed, got:\n%s", unwanted, got)
		}
	}
	if !strings.Contains(got, "Question body") {
		t.Fatalf("expected question body to remain, got:\n%s", got)
	}
}

func TestDropStandaloneControlLinesPreservesQuotedFencesAndPunctuation(t *testing.T) {
	markdown := strings.Join([]string{
		"> Code:",
		">",
		"> ```",
		"> value",
		"> ```",
		"",
		"---",
	}, "\n")

	got := dropStandaloneControlLines()(markdown)
	if got != markdown {
		t.Fatalf("expected punctuation-only Markdown structure to remain\nwant:\n%s\n\ngot:\n%s", markdown, got)
	}
}

func TestDropStandaloneControlLinesRemovesPaginationLinks(t *testing.T) {
	markdown := strings.Join([]string{
		"[2](https://example.com/page-2) [Next](https://example.com/page-2)",
		"",
		"Question body",
	}, "\n")

	got := dropStandaloneControlLines()(markdown)
	if strings.Contains(got, "[2](https://example.com/page-2)") || strings.Contains(got, "[Next](https://example.com/page-2)") {
		t.Fatalf("expected pagination controls to be removed, got:\n%s", got)
	}
	if !strings.Contains(got, "Question body") {
		t.Fatalf("expected body text to remain, got:\n%s", got)
	}
}

func TestDropStandaloneMetricLinesRemovesIsolatedCountsNearControls(t *testing.T) {
	markdown := strings.Join([]string{
		"## 36 Answers 36",
		"",
		"1",
		"",
		"[2](https://example.com/page-2) [Next](https://example.com/page-2)",
		"",
		"First answer body",
	}, "\n")

	got := dropStandaloneMetricLines()(markdown)
	if strings.Contains(got, "\n1\n") {
		t.Fatalf("expected isolated metric line to be removed, got:\n%s", got)
	}
	if !strings.Contains(got, "First answer body") {
		t.Fatalf("expected answer body to remain, got:\n%s", got)
	}
}

func TestCollapseAdjacentDuplicateLines(t *testing.T) {
	markdown := strings.Join([]string{
		"Closed",
		"Closed",
		"",
		"Body",
		"Body",
	}, "\n")

	got := collapseAdjacentDuplicateLines()(markdown)
	if strings.Count(got, "Closed") != 1 {
		t.Fatalf("expected duplicate status line to collapse, got:\n%s", got)
	}
	if strings.Count(got, "Body") != 1 {
		t.Fatalf("expected duplicate body line to collapse, got:\n%s", got)
	}
}

func TestCollapseAdjacentDuplicateLinesCollapsesSpacedShortDuplicates(t *testing.T) {
	markdown := strings.Join([]string{
		"# Iteration Plan for January 2026\\#286040",
		"",
		"[Iteration Plan for January 2026](https://example.com/top)#286040",
		"",
		"Closed",
		"",
		"Closed",
	}, "\n")

	got := collapseAdjacentDuplicateLines()(markdown)
	if strings.Count(got, "Iteration Plan for January 2026") != 1 {
		t.Fatalf("expected heading echo to collapse, got:\n%s", got)
	}
	if strings.Count(got, "Closed") != 1 {
		t.Fatalf("expected spaced duplicate status to collapse, got:\n%s", got)
	}
}

func TestDropRedundantOpeningHeadingEchoesRemovesLinkedTitleEcho(t *testing.T) {
	markdown := strings.Join([]string{
		"# Iteration Plan for January 2026\\#286040",
		"",
		"Closed",
		"",
		"[Iteration Plan for January 2026](https://example.com/top)#286040",
		"",
		"Body starts here.",
	}, "\n")

	got := dropRedundantOpeningHeadingEchoes()(markdown)
	if strings.Contains(got, "[Iteration Plan for January 2026](https://example.com/top)#286040") {
		t.Fatalf("expected redundant linked title echo to be removed, got:\n%s", got)
	}
	if !strings.Contains(got, "Body starts here.") {
		t.Fatalf("expected body text to remain, got:\n%s", got)
	}
}
