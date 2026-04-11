package engine

import (
	"strings"
	"testing"

	"github.com/appautomaton/markmaton/internal/model"
	"github.com/appautomaton/markmaton/internal/testutil"
)

func TestProcessMatchesGoldenFixtures(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		fixture string
		golden  string
	}{
		{
			name:    "article",
			url:     "https://example.com/articles/harnessing-parsers",
			fixture: "core/article.html",
			golden:  "core/article.md",
		},
		{
			name:    "docs",
			url:     "https://example.com/docs/setup",
			fixture: "core/docs.html",
			golden:  "core/docs.md",
		},
		{
			name:    "news",
			url:     "https://example.com/news",
			fixture: "core/news.html",
			golden:  "core/news.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadFixture(t, tc.fixture)
			expected := loadGolden(t, tc.golden)

			response, err := Process(model.Request{
				URL:  tc.url,
				HTML: fixture,
				Options: model.Options{
					OnlyMainContent: model.Bool(true),
				},
			})
			if err != nil {
				t.Fatalf("process failed: %v", err)
			}

			if strings.TrimSpace(response.Markdown) != strings.TrimSpace(expected) {
				t.Fatalf("markdown mismatch\nexpected:\n%s\n\ngot:\n%s", expected, response.Markdown)
			}
		})
	}
}

func TestProcessFallsBackToFullContent(t *testing.T) {
	html := `<html><body><aside><p>This is the actual content.</p><p>It lives in an aside.</p></aside></body></html>`
	response, err := Process(model.Request{
		URL:  "https://example.com/weird-layout",
		HTML: html,
		Options: model.Options{
			OnlyMainContent: model.Bool(true),
		},
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	if !response.Quality.FallbackUsed {
		t.Fatalf("expected fallback to be used")
	}
	if response.Markdown == "" {
		t.Fatalf("expected markdown after fallback")
	}
}

func TestProcessHonorsExplicitFullContentMode(t *testing.T) {
	html := `
	<html>
	  <body>
	    <header><nav><a href="/home">Home</a></nav></header>
	    <article><h1>Main Story</h1><p>Primary article body.</p></article>
	    <section><h2>Visible page chrome</h2><p>Keep this in full-content mode.</p></section>
	  </body>
	</html>`

	response, err := Process(model.Request{
		URL:  "https://example.com/story",
		HTML: html,
		Options: model.Options{
			OnlyMainContent: model.Bool(false),
		},
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	if response.Quality.UsedMainContent {
		t.Fatalf("did not expect explicit full-content mode to report used_main_content=true")
	}
	if response.Quality.FallbackUsed {
		t.Fatalf("did not expect explicit full-content mode to trigger fallback")
	}
	if !strings.Contains(response.Markdown, "Visible page chrome") {
		t.Fatalf("expected full-content mode to preserve additional visible content")
	}
}

func TestProcessUsesRequestURLAsCanonicalFallback(t *testing.T) {
	response, err := Process(model.Request{
		URL:  "https://example.com/no-canonical",
		HTML: `<html><head><title>No Canonical</title></head><body><main><p>Hello</p></main></body></html>`,
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	if response.Metadata.CanonicalURL != "https://example.com/no-canonical" {
		t.Fatalf("expected canonical fallback to use request URL, got %q", response.Metadata.CanonicalURL)
	}
}

func TestProcessPreservesCareersLandingSignals(t *testing.T) {
	response, err := Process(model.Request{
		URL:  "https://openai.com/careers/",
		HTML: loadFixture(t, "regression/careers_landing.html"),
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	if response.Quality.FallbackUsed {
		t.Fatalf("did not expect fallback for careers landing fixture")
	}
	if !strings.Contains(response.Markdown, "Develop safe, beneficial AI systems") {
		t.Fatalf("expected careers hero copy to remain")
	}
	if !strings.Contains(response.Markdown, "[View open roles](https://openai.com/careers/search/)") {
		t.Fatalf("expected primary careers CTA to remain")
	}
}

func TestProcessCleansGenericCardListControls(t *testing.T) {
	response, err := Process(model.Request{
		URL:  "https://openai.com/news/engineering/",
		HTML: loadFixture(t, "regression/card_grid.html"),
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	for _, unwanted := range []string{
		"Filter",
		"Sort",
		"Switch cards to show Media",
		"Switch cards to hide Media",
	} {
		if strings.Contains(response.Markdown, unwanted) {
			t.Fatalf("expected %q to be removed from card grid markdown", unwanted)
		}
	}

	for _, expected := range []string{
		"Engineering",
		"From model to agent: Equipping the Responses API with a computer environment",
		"Mar 11, 2026",
		"Beyond rate limits: scaling access to Codex and Sora",
		"Feb 13, 2026",
	} {
		if !strings.Contains(response.Markdown, expected) {
			t.Fatalf("expected %q to remain in card grid markdown", expected)
		}
	}
}

func TestProcessPreservesJobDetailSignals(t *testing.T) {
	response, err := Process(model.Request{
		URL:  "https://jobs.ashbyhq.com/openai/example/application",
		HTML: loadFixture(t, "regression/job_detail.html"),
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	if response.Quality.FallbackUsed {
		t.Fatalf("did not expect fallback for job detail fixture")
	}
	for _, expected := range []string{
		"Abuse Investigator",
		"San Francisco; Remote - US",
		"$288K – $425K • Offers Equity",
	} {
		if !strings.Contains(response.Markdown, expected) {
			t.Fatalf("expected %q to remain in job detail markdown", expected)
		}
	}
}

func TestProcessSuppressesShellHeavyBlogChrome(t *testing.T) {
	response, err := Process(model.Request{
		URL:  "https://developers.openai.com/blog",
		HTML: loadFixture(t, "regression/openai_blog_shell.html"),
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}
	if response.Quality.FallbackUsed {
		t.Fatalf("did not expect fallback for shell-heavy blog fixture")
	}

	for _, unwanted := range []string{"Search the blog", "Search docs", "Primary navigation", "{{ className }}"} {
		if strings.Contains(response.Markdown, unwanted) {
			t.Fatalf("expected %q to be removed from blog markdown", unwanted)
		}
	}
	if !strings.Contains(response.Markdown, "Launch notes for the Responses API") {
		t.Fatalf("expected blog content to remain")
	}
}

func TestProcessSuppressesShellHeavyRepoChrome(t *testing.T) {
	response, err := Process(model.Request{
		URL:  "https://github.com/zellij-org/zellij",
		HTML: loadFixture(t, "regression/github_repo_shell.html"),
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}
	if response.Quality.FallbackUsed {
		t.Fatalf("did not expect fallback for shell-heavy repo fixture")
	}

	for _, unwanted := range []string{
		"Skip to content",
		"You signed in with another tab or window",
		"Dismiss alert",
		"{{ className }}",
		"Uh oh!",
		"Please reload this page",
		"Repository files navigation",
	} {
		if strings.Contains(response.Markdown, unwanted) {
			t.Fatalf("expected %q to be removed from repo markdown", unwanted)
		}
	}
	if !strings.Contains(response.Markdown, "A terminal workspace with batteries included.") {
		t.Fatalf("expected repository summary to remain")
	}
}

func TestProcessRetainsDiscussionFixtureCoreSignals(t *testing.T) {
	response, err := Process(model.Request{
		URL:  "https://stackoverflow.com/questions/1732348/regex-match-open-tags-except-xhtml-self-contained-tags",
		HTML: loadFixture(t, "regression/stackoverflow_question_thread.html"),
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	for _, unwanted := range []string{
		"Collectives™ on Stack Overflow",
		"Find centralized, trusted content",
		"Knowledge at work",
		"[Share](",
		"Improve this question",
		"Reset to default",
	} {
		if strings.Contains(response.Markdown, unwanted) {
			t.Fatalf("expected %q to be removed from discussion fixture", unwanted)
		}
	}

	for _, expected := range []string{
		"I need to match all of these opening tags:",
		"36 Answers",
		"You can't parse \\[X\\]HTML with regex.",
	} {
		if !strings.Contains(response.Markdown, expected) {
			t.Fatalf("expected %q to remain in discussion fixture markdown", expected)
		}
	}

	firstNonEmpty := firstNonEmptyLine(response.Markdown)
	if firstNonEmpty != "I need to match all of these opening tags:" {
		t.Fatalf("expected discussion fixture to open on question body, got %q", firstNonEmpty)
	}
}

func TestProcessRetainsTimelineFixtureCoreSignals(t *testing.T) {
	response, err := Process(model.Request{
		URL:  "https://github.com/microsoft/vscode/issues/286040",
		HTML: loadFixture(t, "regression/github_issue_timeline.html"),
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	for _, unwanted := range []string{
		"Skip to content",
		"You signed in with another tab or window",
		"Dismiss alert",
		"[New issue](",
	} {
		if strings.Contains(response.Markdown, unwanted) {
			t.Fatalf("expected %q to be removed from timeline fixture", unwanted)
		}
	}

	for _, expected := range []string{
		"Iteration Plan for January 2026",
		"# Iteration Plan for January 2026\\#286040",
		"Closed",
	} {
		if !strings.Contains(response.Markdown, expected) {
			t.Fatalf("expected %q to remain in timeline fixture markdown", expected)
		}
	}

	firstNonEmpty := firstNonEmptyLine(response.Markdown)
	if firstNonEmpty != "# Iteration Plan for January 2026\\#286040" {
		t.Fatalf("expected timeline fixture to open on the issue title, got %q", firstNonEmpty)
	}
	if strings.Contains(response.Markdown, "[Iteration Plan for January 2026](https://github.com/microsoft/vscode/issues/286040#top)#286040") {
		t.Fatalf("expected redundant title echo to be removed from timeline fixture")
	}
}

func loadFixture(t *testing.T, name string) string {
	return testutil.ReadFixture(t, name)
}

func loadGolden(t *testing.T, name string) string {
	return testutil.ReadGolden(t, name)
}

func firstNonEmptyLine(markdown string) string {
	for _, line := range strings.Split(markdown, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
