package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/appautomaton/markmaton/internal/model"
	"github.com/appautomaton/markmaton/internal/testutil"
)

type processGoldenFixture struct {
	name    string
	url     string
	fixture string
	golden  string
}

var processGoldenFixtures = []processGoldenFixture{
	{name: "article", url: "https://example.com/articles/harnessing-parsers", fixture: "core/article.html", golden: "core/article.md"},
	{name: "docs", url: "https://example.com/docs/setup", fixture: "core/docs.html", golden: "core/docs.md"},
	{name: "news", url: "https://example.com/news", fixture: "core/news.html", golden: "core/news.md"},
}

var processExactRegressionFixtures = []processGoldenFixture{
	{name: "embedded media", url: "https://example.com/articles/media", fixture: "regression/embedded_media.html", golden: "regression/embedded_media.md"},
	{name: "native semantic structures", url: "https://example.com/issues/42", fixture: "regression/native_semantic_structures.html", golden: "regression/native_semantic_structures.md"},
	{name: "safe destinations", url: "https://example.com/articles/safety", fixture: "regression/safe_destinations.html", golden: "regression/safe_destinations.md"},
}

type processRegressionFixture struct {
	name          string
	url           string
	fixture       string
	expected      []string
	unwanted      []string
	firstNonEmpty string
	allowFallback bool
}

var processRegressionFixtures = []processRegressionFixture{
	{
		name:     "careers landing",
		url:      "https://openai.com/careers/",
		fixture:  "regression/careers_landing.html",
		expected: []string{"Develop safe, beneficial AI systems", "[View open roles](https://openai.com/careers/search/)"},
	},
	{
		name:     "card grid",
		url:      "https://openai.com/news/engineering/",
		fixture:  "regression/card_grid.html",
		expected: []string{"Engineering", "From model to agent: Equipping the Responses API with a computer environment", "Mar 11, 2026", "Beyond rate limits: scaling access to Codex and Sora", "Feb 13, 2026"},
		unwanted: []string{"Filter", "Sort", "Switch cards to show Media", "Switch cards to hide Media"},
	},
	{
		name:     "job detail",
		url:      "https://jobs.ashbyhq.com/openai/example/application",
		fixture:  "regression/job_detail.html",
		expected: []string{"Abuse Investigator", "San Francisco; Remote - US", "$288K – $425K • Offers Equity"},
	},
	{
		name:     "blog shell",
		url:      "https://developers.openai.com/blog",
		fixture:  "regression/openai_blog_shell.html",
		expected: []string{"Launch notes for the Responses API"},
		unwanted: []string{"Search the blog", "Search docs", "Primary navigation", "{{ className }}"},
	},
	{
		name:     "repository shell",
		url:      "https://github.com/zellij-org/zellij",
		fixture:  "regression/github_repo_shell.html",
		expected: []string{"A terminal workspace with batteries included."},
		unwanted: []string{"Skip to content", "You signed in with another tab or window", "Dismiss alert", "{{ className }}", "Uh oh!", "Please reload this page", "Repository files navigation"},
	},
	{
		name:          "discussion thread",
		url:           "https://stackoverflow.com/questions/1732348/regex-match-open-tags-except-xhtml-self-contained-tags",
		fixture:       "regression/stackoverflow_question_thread.html",
		expected:      []string{"I need to match all of these opening tags:", "36 Answers", "You can't parse \\[X\\]HTML with regex."},
		unwanted:      []string{"Collectives™ on Stack Overflow", "Find centralized, trusted content", "Knowledge at work", "[Share](", "Improve this question", "Reset to default"},
		firstNonEmpty: "I need to match all of these opening tags:",
		allowFallback: true,
	},
	{
		name:          "issue timeline",
		url:           "https://github.com/microsoft/vscode/issues/286040",
		fixture:       "regression/github_issue_timeline.html",
		expected:      []string{"Iteration Plan for January 2026", "# Iteration Plan for January 2026 \\#286040", "This plan captures our work in **January 2026**.", "## Plan Items", "Closed"},
		unwanted:      []string{"Skip to content", "You signed in with another tab or window", "Dismiss alert", "[New issue](", "[Iteration Plan for January 2026](https://github.com/microsoft/vscode/issues/286040#top)#286040"},
		firstNonEmpty: "# Iteration Plan for January 2026 \\#286040",
		allowFallback: true,
	},
}

func TestProcessMatchesGoldenFixtures(t *testing.T) {
	fixtures := append(append([]processGoldenFixture(nil), processGoldenFixtures...), processExactRegressionFixtures...)
	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			html := loadFixture(t, fixture.fixture)
			expected := loadGolden(t, fixture.golden)

			response, err := Process(model.Request{
				URL:  fixture.url,
				HTML: html,
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

func TestProcessRegressionCorpus(t *testing.T) {
	for _, fixture := range processRegressionFixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			response, err := Process(model.Request{
				URL:  fixture.url,
				HTML: loadFixture(t, fixture.fixture),
			})
			if err != nil {
				t.Fatalf("process regression fixture: %v", err)
			}
			if response.Quality.FallbackUsed && !fixture.allowFallback {
				t.Fatal("unexpected fallback for regression fixture")
			}
			for _, expected := range fixture.expected {
				if !strings.Contains(response.Markdown, expected) {
					t.Fatalf("expected %q to remain in regression output", expected)
				}
			}
			for _, unwanted := range fixture.unwanted {
				if strings.Contains(response.Markdown, unwanted) {
					t.Fatalf("expected %q to be removed from regression output", unwanted)
				}
			}
			if fixture.firstNonEmpty != "" {
				if got := firstNonEmptyLine(response.Markdown); got != fixture.firstNonEmpty {
					t.Fatalf("expected first output line %q, got %q", fixture.firstNonEmpty, got)
				}
			}
		})
	}
}

func TestProcessFixtureCorpusIsFullyClassified(t *testing.T) {
	coreFixtures := make([]string, 0, len(processGoldenFixtures))
	coreGoldens := make([]string, 0, len(processGoldenFixtures))
	for _, fixture := range processGoldenFixtures {
		coreFixtures = append(coreFixtures, filepath.Base(fixture.fixture))
		coreGoldens = append(coreGoldens, filepath.Base(fixture.golden))
	}

	regressionFixtures := make([]string, 0, len(processRegressionFixtures)+len(processExactRegressionFixtures))
	for _, fixture := range processRegressionFixtures {
		regressionFixtures = append(regressionFixtures, filepath.Base(fixture.fixture))
	}
	regressionGoldens := make([]string, 0, len(processExactRegressionFixtures))
	for _, fixture := range processExactRegressionFixtures {
		regressionFixtures = append(regressionFixtures, filepath.Base(fixture.fixture))
		regressionGoldens = append(regressionGoldens, filepath.Base(fixture.golden))
	}

	assertCorpusFiles(t, filepath.Join(testutil.RepoRoot(t), "testdata", "fixtures", "core"), ".html", coreFixtures)
	assertCorpusFiles(t, filepath.Join(testutil.RepoRoot(t), "testdata", "golden", "core"), ".md", coreGoldens)
	assertCorpusFiles(t, filepath.Join(testutil.RepoRoot(t), "testdata", "fixtures", "regression"), ".html", regressionFixtures)
	assertCorpusFiles(t, filepath.Join(testutil.RepoRoot(t), "testdata", "golden", "regression"), ".md", regressionGoldens)
}

func assertCorpusFiles(t testing.TB, directory, extension string, expected []string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read fixture corpus %q: %v", directory, err)
	}
	actual := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == extension {
			actual = append(actual, entry.Name())
		}
	}
	sort.Strings(actual)
	sort.Strings(expected)
	if strings.Join(actual, "\n") != strings.Join(expected, "\n") {
		t.Fatalf("fixture corpus mismatch in %s\nactual:\n%s\n\nexpected:\n%s", directory, strings.Join(actual, "\n"), strings.Join(expected, "\n"))
	}
}

func TestProcessFallbackPreservesCanonicalMarkdown(t *testing.T) {
	response, err := Process(model.Request{
		URL:  "https://example.com/native-fallback",
		HTML: `<html><body><aside><p>This is the actual content.</p><hr><p>It lives in an aside.</p></aside></body></html>`,
		Options: model.Options{
			OnlyMainContent: model.Bool(true),
		},
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}
	if !response.Quality.FallbackUsed {
		t.Fatal("expected fallback to be used")
	}
	if !strings.Contains(response.Markdown, "---") || strings.Contains(response.Markdown, "* * *") {
		t.Fatalf("expected fallback to preserve Native thematic-break output:\n%s", response.Markdown)
	}
}

func TestProcessContextHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ProcessContext(ctx, model.Request{HTML: `<p>Hello</p>`})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestProcessUsesNativeConversion(t *testing.T) {
	response, err := Process(model.Request{
		URL:  "https://example.com/native",
		HTML: `<html><body><article><h1>Native Engine</h1><p>Hello <strong>world</strong>.</p></article></body></html>`,
		Options: model.Options{
			OnlyMainContent: model.Bool(false),
		},
	})
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}
	if response.Markdown != "# Native Engine\n\nHello **world**." {
		t.Fatalf("unexpected native markdown:\n%s", response.Markdown)
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
