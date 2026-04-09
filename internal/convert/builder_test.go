package convert

import (
	"strings"
	"testing"

	md "github.com/firecrawl/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

func TestDefaultBuilderExposesDefaultPluginNames(t *testing.T) {
	builder := DefaultBuilder("")

	got := strings.Join(builder.PluginNames(), ",")
	wantParts := []string{"github_flavored", "robust_code_block"}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("expected default builder to include plugin %q, got %q", want, got)
		}
	}
}

func TestBuilderAppliesBeforeAndAfterHooks(t *testing.T) {
	builder := NewBuilder("").
		WithBeforeHooks(newBeforeHookRegistration("drop-button", func(selec *goquery.Selection) {
			selec.Find("button").Each(func(_ int, s *goquery.Selection) {
				s.Remove()
			})
		})).
		WithAfterHooks(newAfterHookRegistration("append-sentinel", func(markdown string) string {
			return markdown + "\n\nHOOKED"
		}))

	converter := builder.Build()
	markdown, err := converter.ConvertString(`<article><button>Ignore me</button><p>Hello world.</p></article>`)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	if strings.Contains(markdown, "Ignore me") {
		t.Fatalf("expected before hook to remove button content, got:\n%s", markdown)
	}
	if !strings.Contains(markdown, "HOOKED") {
		t.Fatalf("expected after hook marker, got:\n%s", markdown)
	}
}

func TestBuilderAppliesCustomRulesDeterministically(t *testing.T) {
	rule := md.Rule{
		Filter: []string{"mark"},
		Replacement: func(content string, _ *goquery.Selection, _ *md.Options) *string {
			return md.String("==" + strings.TrimSpace(content) + "==")
		},
	}

	builder := NewBuilder("").WithRules(rule)
	converter := builder.Build()
	markdown, err := converter.ConvertString(`<article><p>before <mark>important</mark> after</p></article>`)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	if !strings.Contains(markdown, "==important==") {
		t.Fatalf("expected custom rule output, got:\n%s", markdown)
	}
}
