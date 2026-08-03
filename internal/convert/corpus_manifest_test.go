package convert

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/appautomaton/markmaton/internal/testutil"
)

type referenceExpectation string

const (
	referenceExact     referenceExpectation = "exact"
	referenceSemantic  referenceExpectation = "semantic"
	referenceDivergent referenceExpectation = "intentional_divergence"
)

type conversionCorpusCase struct {
	ID                              string
	Expectation                     referenceExpectation
	BaseURL                         string
	AllowReferenceContentDivergence bool
	Reason                          string
}

var conversionCorpus = []conversionCorpusCase{
	{ID: "commonmark/blockquote", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native follows HTML whitespace semantics for direct quote text while preserving nested block structure."},
	{ID: "commonmark/strong", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native preserves source adjacency and normalizes nested strong elements without reproducing reference whitespace insertion."},
	{ID: "commonmark/line-break", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native emits explicit Markdown hard breaks instead of converting every HTML break into a separate paragraph."},
	{ID: "commonmark/heading", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native uses ATX headings without reproducing duplicate output around headings that contain block elements."},
	{ID: "commonmark/thematic-break", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Markmaton canonicalizes thematic breaks to three hyphens instead of spaced asterisks."},
	{ID: "commonmark/image", Expectation: referenceDivergent, BaseURL: "http://example.com/", AllowReferenceContentDivergence: true, Reason: "Native normalizes unsafe destination whitespace and preserves explicit hard breaks without duplicating linked images."},
	{ID: "commonmark/emphasis", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native preserves source adjacency, collapses redundant nesting, and splits emphasis safely across hard breaks."},
	{ID: "commonmark/link", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native preserves source adjacency and block structure inside links while sanitizing destinations deterministically."},
	{ID: "commonmark/list", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native emits contiguous semantic numbering without reference numbering gaps caused by empty or non-list children."},
	{ID: "commonmark/nested-list", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native normalizes malformed nested-list wrappers while preserving every readable item."},
	{ID: "commonmark/paragraph", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native collapses source formatting whitespace according to HTML paragraph semantics."},
	{ID: "commonmark/code", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native chooses collision-safe fences, preserves code whitespace, and uses explicit hard breaks for adjacent inline code."},
	{ID: "commonmark/superscript", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native preserves superscript text without depending on raw HTML passthrough."},
	{ID: "gfm/task-list", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native emits compact valid GFM task lists without blank lines between every item."},
	{ID: "gfm/strikethrough", Expectation: referenceExact, BaseURL: "http://example.com/", Reason: "GitHub strikethrough is part of Markmaton's canonical GFM output."},
	{ID: "gfm/table", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native uses GFM pipe tables for simple data and readable text fallback for spanning, nested, or block-rich tables."},
	{ID: "real-world/go-blog", Expectation: referenceSemantic, BaseURL: "http://blog.golang.org/", Reason: "The page is a broad semantic regression fixture; site chrome and formatting details are not a byte-level external reference contract."},
	{ID: "real-world/go-home", Expectation: referenceSemantic, BaseURL: "http://golang.org/", Reason: "The page is a broad semantic regression fixture rather than a stable byte-level external reference contract."},
	{ID: "real-world/github-about", Expectation: referenceSemantic, BaseURL: "http://example.com/", Reason: "Repository chrome is validated by retained content and link semantics instead of external-reference whitespace."},
	{ID: "real-world/linked-heading", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native preserves heading structure when invalid HTML places a heading inside a link."},
	{ID: "real-world/navigation-list", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native normalizes malformed navigation lists without reproducing empty or disconnected markers."},
	{ID: "real-world/highlighted-code", Expectation: referenceDivergent, BaseURL: "http://example.com/", AllowReferenceContentDivergence: true, Reason: "Native detects highlighter languages and chooses safe fences without reproducing code-wrapper artifacts."},
	{ID: "real-world/inline-price-emphasis", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native uses an intraword-safe emphasis delimiter without inserting spaces absent from the HTML."},
	{ID: "real-world/literal-brackets", Expectation: referenceExact, BaseURL: "http://example.com/", Reason: "Literal square brackets must be escaped deterministically."},
	{ID: "real-world/source-whitespace", Expectation: referenceDivergent, BaseURL: "http://example.com/", Reason: "Native follows HTML whitespace semantics and keeps inline links inline instead of reflecting source indentation."},
	{ID: "real-world/mixed-markdown", Expectation: referenceSemantic, BaseURL: "http://example.com/", Reason: "The mixed-feature document is validated as an aggregate semantic fixture."},
	{ID: "real-world/social-embed", Expectation: referenceSemantic, BaseURL: "http://example.com/", Reason: "Social embed markup is validated for retained readable content rather than provider-specific whitespace."},
}

var performanceCorpusFiles = []string{
	"large-list.html",
	"large-table.html",
	"mixed.html",
}

func TestConversionCorpusManifestCoversEveryFixtureAndGolden(t *testing.T) {
	fixtureRoot := conversionFixtureRoot(t)
	goldenRoot := conversionGoldenRoot(t)
	actualFixtures := relativeFiles(t, fixtureRoot, ".html")
	actualGoldens := relativeFiles(t, goldenRoot, ".md")
	declaredFixtures := make([]string, 0, len(conversionCorpus))
	declaredGoldens := make([]string, 0, len(conversionCorpus)*2)
	seen := make(map[string]struct{}, len(conversionCorpus))

	for _, fixture := range conversionCorpus {
		if fixture.ID == "" || fixture.BaseURL == "" || fixture.Reason == "" {
			t.Fatalf("conversion corpus entries require an ID, base URL, and English reason: %+v", fixture)
		}
		if !validReferenceExpectation(fixture.Expectation) {
			t.Fatalf("conversion corpus entry %q has invalid expectation %q", fixture.ID, fixture.Expectation)
		}
		if fixture.AllowReferenceContentDivergence && fixture.Expectation != referenceDivergent {
			t.Fatalf("conversion corpus entry %q may allow reference-content divergence only for an intentional divergence", fixture.ID)
		}
		if _, exists := seen[fixture.ID]; exists {
			t.Fatalf("conversion corpus entry %q is declared more than once", fixture.ID)
		}
		seen[fixture.ID] = struct{}{}
		declaredFixtures = append(declaredFixtures, fixture.ID+".html")
		declaredGoldens = append(declaredGoldens, filepath.ToSlash(filepath.Join(fixture.ID, "expected.md")))
		if fixture.Expectation != referenceExact {
			declaredGoldens = append(declaredGoldens, filepath.ToSlash(filepath.Join(fixture.ID, "reference.md")))
		}
	}

	sort.Strings(declaredFixtures)
	sort.Strings(declaredGoldens)
	assertStringSlicesEqual(t, "conversion fixture manifest", actualFixtures, declaredFixtures)
	assertStringSlicesEqual(t, "conversion golden manifest", actualGoldens, declaredGoldens)
}

func TestPerformanceCorpusManifestCoversEveryFixture(t *testing.T) {
	actual := relativeFiles(t, performanceFixtureRoot(t), ".html")
	declared := append([]string(nil), performanceCorpusFiles...)
	sort.Strings(declared)
	assertStringSlicesEqual(t, "performance fixture manifest", actual, declared)
}

func TestIntegratedCorpusProvenanceIsDocumented(t *testing.T) {
	path := filepath.Join(testutil.RepoRoot(t), "testdata", "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read testdata provenance: %v", err)
	}
	text := string(data)
	for _, required := range []string{"fixtures/compatibility", "golden/compatibility", "fixtures/performance", "html-to-markdown"} {
		if !strings.Contains(text, required) {
			t.Fatalf("testdata provenance must mention %q", required)
		}
	}
}

func validReferenceExpectation(value referenceExpectation) bool {
	switch value {
	case referenceExact, referenceSemantic, referenceDivergent:
		return true
	default:
		return false
	}
}

func conversionFixtureRoot(t testing.TB) string {
	t.Helper()
	return filepath.Join(testutil.RepoRoot(t), "testdata", "fixtures", "compatibility")
}

func conversionGoldenRoot(t testing.TB) string {
	t.Helper()
	return filepath.Join(testutil.RepoRoot(t), "testdata", "golden", "compatibility")
}

func performanceFixtureRoot(t testing.TB) string {
	t.Helper()
	return filepath.Join(testutil.RepoRoot(t), "testdata", "fixtures", "performance")
}

func relativeFiles(t testing.TB, root, suffix string) []string {
	t.Helper()
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		t.Fatalf("walk corpus files: %v", err)
	}
	sort.Strings(paths)
	return paths
}

func assertStringSlicesEqual(t testing.TB, label string, got, want []string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("%s mismatch\ngot:\n%s\n\nwant:\n%s", label, strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
