package convert

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/appautomaton/markmaton/internal/postprocess"
	"github.com/appautomaton/markmaton/internal/resolve"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"golang.org/x/net/html"
)

func TestConversionCorpus(t *testing.T) {
	for _, fixture := range conversionCorpus {
		fixture := fixture
		t.Run(strings.ReplaceAll(fixture.ID, "/", "_"), func(t *testing.T) {
			source := readConversionFixture(t, fixture.ID)
			resolved, err := resolve.ResolveHTML(source, fixture.BaseURL)
			if err != nil {
				t.Fatalf("resolve conversion fixture URLs: %v", err)
			}

			output := convertCorpusFixture(t, resolved)
			assertConversionGolden(t, fixture, output)

			switch fixture.Expectation {
			case referenceExact:
				return
			case referenceSemantic:
				reference := readConversionGolden(t, fixture.ID, "reference.md")
				if report := compareSemanticMarkdown(t, reference, output); report != "" {
					t.Fatalf("Native differs semantically from the retained reference: %s\n%s", fixture.Reason, report)
				}
			case referenceDivergent:
				reference := readConversionGolden(t, fixture.ID, "reference.md")
				if output == reference {
					t.Fatalf("expected Native to differ from the retained reference: %s", fixture.Reason)
				}
				if !fixture.AllowReferenceContentDivergence {
					if report := compareContentMarkdown(t, reference, output); report != "" {
						t.Fatalf("Native lost retained reference content: %s\n%s", fixture.Reason, report)
					}
				}
			default:
				t.Fatalf("unsupported reference expectation %q", fixture.Expectation)
			}
		})
	}
}

func normalizeCorpusMarkdown(markdown string) string {
	return strings.TrimSpace(strings.ReplaceAll(markdown, "\r\n", "\n"))
}

func TestSemanticComparisonDetectsContentAndStructureLoss(t *testing.T) {
	t.Run("structure", func(t *testing.T) {
		reference := "# Heading\n\n- Item"
		native := "Heading\n\nItem"
		if report := compareSemanticMarkdown(t, reference, native); !strings.Contains(report, "block structure") {
			t.Fatalf("expected structural mismatch, got %q", report)
		}
		if report := compareContentMarkdown(t, reference, native); report != "" {
			t.Fatalf("did not expect content-only mismatch, got %q", report)
		}
	})

	t.Run("link destination", func(t *testing.T) {
		reference := "[Documentation](https://example.com/docs)"
		native := "Documentation"
		if report := compareContentMarkdown(t, reference, native); !strings.Contains(report, "links similarity") {
			t.Fatalf("expected link mismatch, got %q", report)
		}
	})

	t.Run("missing text", func(t *testing.T) {
		reference := "alpha beta gamma delta epsilon zeta eta theta iota kappa"
		native := "alpha beta gamma"
		if report := compareContentMarkdown(t, reference, native); !strings.Contains(report, "text tokens similarity") {
			t.Fatalf("expected text mismatch, got %q", report)
		}
	})
}

func convertCorpusFixture(t testing.TB, source string) string {
	t.Helper()
	markdown, err := ToMarkdownContext(context.Background(), source)
	if err != nil {
		t.Fatalf("convert corpus fixture with Native: %v", err)
	}
	return postprocess.Markdown(markdown)
}

func assertConversionGolden(t testing.TB, fixture conversionCorpusCase, markdown string) {
	t.Helper()
	path := filepath.Join(conversionGoldenRoot(t), filepath.FromSlash(fixture.ID), "expected.md")
	if os.Getenv("MARKMATON_UPDATE_CONVERSION_GOLDENS") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create conversion golden directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(markdown+"\n"), 0o644); err != nil {
			t.Fatalf("update conversion golden: %v", err)
		}
	}

	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read conversion golden %q: %v; set MARKMATON_UPDATE_CONVERSION_GOLDENS=1 to create it", path, err)
	}
	want := normalizeCorpusMarkdown(string(expected))
	if want != markdown {
		t.Fatalf("Native conversion golden mismatch: %s\n%s", fixture.Reason, formatLineDifference(want, markdown))
	}
}

type semanticSummary struct {
	textTokens     map[string]struct{}
	links          map[string]struct{}
	images         map[string]struct{}
	blockStructure map[string]struct{}
}

type semanticCheck struct {
	name      string
	reference map[string]struct{}
	native    map[string]struct{}
	threshold float64
}

func compareSemanticMarkdown(t testing.TB, reference, native string) string {
	t.Helper()
	return compareMarkdownSemantics(t, reference, native, true)
}

func compareContentMarkdown(t testing.TB, reference, native string) string {
	t.Helper()
	return compareMarkdownSemantics(t, reference, native, false)
}

func compareMarkdownSemantics(t testing.TB, reference, native string, includeStructure bool) string {
	t.Helper()
	referenceSummary := summarizeMarkdownSemantics(t, reference)
	nativeSummary := summarizeMarkdownSemantics(t, native)

	checks := []semanticCheck{
		{name: "text tokens", reference: referenceSummary.textTokens, native: nativeSummary.textTokens, threshold: 0.90},
		{name: "links", reference: referenceSummary.links, native: nativeSummary.links, threshold: 0.80},
		{name: "images", reference: referenceSummary.images, native: nativeSummary.images, threshold: 0.80},
	}
	if includeStructure {
		checks = append(checks, semanticCheck{name: "block structure", reference: referenceSummary.blockStructure, native: nativeSummary.blockStructure, threshold: 0.75})
	}

	var reports []string
	for _, check := range checks {
		similarity := setSimilarity(check.reference, check.native)
		if similarity >= check.threshold {
			continue
		}
		reports = append(reports, fmt.Sprintf(
			"%s similarity %.2f is below %.2f\nmissing from Native: %s\nextra in Native: %s",
			check.name,
			similarity,
			check.threshold,
			strings.Join(setDifference(check.reference, check.native, 20), ", "),
			strings.Join(setDifference(check.native, check.reference, 20), ", "),
		))
	}
	return strings.Join(reports, "\n")
}

func summarizeMarkdownSemantics(t testing.TB, markdown string) semanticSummary {
	t.Helper()
	converter := goldmark.New(goldmark.WithExtensions(extension.GFM))
	var rendered bytes.Buffer
	if err := converter.Convert([]byte(markdown), &rendered); err != nil {
		t.Fatalf("render Markdown semantics: %v", err)
	}
	root, err := html.Parse(strings.NewReader(rendered.String()))
	if err != nil {
		t.Fatalf("parse rendered Markdown semantics: %v", err)
	}

	summary := semanticSummary{
		textTokens:     make(map[string]struct{}),
		links:          make(map[string]struct{}),
		images:         make(map[string]struct{}),
		blockStructure: make(map[string]struct{}),
	}
	stack := []*html.Node{root}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if current.Type == html.TextNode {
			for _, token := range semanticTokens(current.Data) {
				summary.textTokens[token] = struct{}{}
			}
		}
		if current.Type == html.ElementNode {
			switch current.Data {
			case "a":
				if href := htmlAttribute(current, "href"); href != "" {
					summary.links[normalizeSemanticURL(href)] = struct{}{}
				}
			case "img":
				if src := htmlAttribute(current, "src"); src != "" {
					summary.images[normalizeSemanticURL(src)] = struct{}{}
				}
			}
			if isSemanticBlockElement(current.Data) {
				if text := strings.Join(semanticTokens(descendantText(current)), " "); text != "" {
					summary.blockStructure[current.Data+":"+text] = struct{}{}
				}
			}
		}
		for child := current.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, child)
		}
	}
	return summary
}

func semanticTokens(value string) []string {
	var normalized strings.Builder
	var previous rune
	for _, current := range value {
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			if unicode.IsUpper(current) && (unicode.IsLower(previous) || unicode.IsDigit(previous)) {
				normalized.WriteByte(' ')
			}
			normalized.WriteRune(unicode.ToLower(current))
			previous = current
			continue
		}
		normalized.WriteByte(' ')
		previous = 0
	}
	return strings.Fields(normalized.String())
}

func htmlAttribute(node *html.Node, name string) string {
	for _, attribute := range node.Attr {
		if attribute.Key == name {
			return attribute.Val
		}
	}
	return ""
}

func normalizeSemanticURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return value
	}
	if parsed.Host != "" && parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String()
}

func setSimilarity(left, right map[string]struct{}) float64 {
	if len(left) == 0 && len(right) == 0 {
		return 1
	}
	intersection := 0
	for value := range left {
		if _, ok := right[value]; ok {
			intersection++
		}
	}
	union := len(left) + len(right) - intersection
	if union == 0 {
		return 1
	}
	return float64(intersection) / float64(union)
}

func setDifference(left, right map[string]struct{}, limit int) []string {
	values := make([]string, 0)
	for value := range left {
		if _, ok := right[value]; !ok {
			values = append(values, value)
		}
	}
	sort.Strings(values)
	if len(values) > limit {
		values = append(values[:limit], "...")
	}
	return values
}

func isSemanticBlockElement(name string) bool {
	switch name {
	case "h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "li", "pre", "th", "td":
		return true
	default:
		return false
	}
}

func descendantText(node *html.Node) string {
	var builder strings.Builder
	stack := []*html.Node{node}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
			builder.WriteByte(' ')
		}
		for child := current.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, child)
		}
	}
	return builder.String()
}

func formatLineDifference(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	limit := len(wantLines)
	if len(gotLines) > limit {
		limit = len(gotLines)
	}

	var builder strings.Builder
	builder.WriteString("--- expected\n+++ actual\n")
	shown := 0
	for index := 0; index < limit && shown < 40; index++ {
		var wantLine, gotLine string
		if index < len(wantLines) {
			wantLine = wantLines[index]
		}
		if index < len(gotLines) {
			gotLine = gotLines[index]
		}
		if wantLine == gotLine {
			continue
		}
		fmt.Fprintf(&builder, "@@ line %d @@\n- %s\n+ %s\n", index+1, wantLine, gotLine)
		shown++
	}
	if shown == 0 && want != got {
		builder.WriteString("outputs differ outside line-oriented normalization\n")
	}
	if shown == 40 {
		builder.WriteString("... difference output truncated ...\n")
	}
	return builder.String()
}

func readConversionFixture(t testing.TB, id string) string {
	t.Helper()
	path := filepath.Join(conversionFixtureRoot(t), filepath.FromSlash(id)+".html")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read conversion fixture %q: %v", path, err)
	}
	return string(data)
}

func readConversionGolden(t testing.TB, id, name string) string {
	t.Helper()
	path := filepath.Join(conversionGoldenRoot(t), filepath.FromSlash(id), name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read conversion golden %q: %v", path, err)
	}
	return normalizeCorpusMarkdown(string(data))
}
