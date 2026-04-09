package cleanhtml

import (
	"strings"
	"testing"

	"github.com/appautomaton/markmaton/internal/testutil"
)

func TestCleanRemovesLayoutNoise(t *testing.T) {
	html := testutil.ReadFixture(t, "core/article.html")
	cleaned, err := Clean(html, true, nil, nil)
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	if strings.Contains(cleaned, "Home") {
		t.Fatalf("expected nav content to be removed")
	}
	if strings.Contains(cleaned, "Copyright Example") {
		t.Fatalf("expected footer content to be removed")
	}
	if !strings.Contains(cleaned, "Harnessing Parsers") {
		t.Fatalf("expected article content to remain")
	}
}

func TestCleanHonorsIncludeAndExcludeSelectors(t *testing.T) {
	html := testutil.ReadFixture(t, "core/docs.html")
	cleaned, err := Clean(html, true, []string{"main"}, []string{"ul"})
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	if strings.Contains(cleaned, "Sidebar links") {
		t.Fatalf("expected sidebar content to be dropped")
	}
	if strings.Contains(cleaned, "Install the package") {
		t.Fatalf("expected excluded selector content to be removed")
	}
}

func TestCleanFocusesOnPreferredMainRootAndDropsDialogChrome(t *testing.T) {
	html := testutil.ReadFixture(t, "regression/openai_blog_shell.html")
	cleaned, err := Clean(html, true, nil, nil)
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	if strings.Contains(cleaned, `href="/"`) || strings.Contains(cleaned, `href="/api"`) {
		t.Fatalf("expected header navigation to be removed")
	}
	if strings.Contains(cleaned, "Search the blog") || strings.Contains(cleaned, "Search docs") {
		t.Fatalf("expected dialog/search chrome to be removed")
	}
	if strings.Contains(cleaned, "Primary navigation") {
		t.Fatalf("expected shell navigation copy to be removed")
	}
	if !strings.Contains(cleaned, "Launch notes for the Responses API") {
		t.Fatalf("expected blog article content to remain")
	}
}

func TestCleanDropsHiddenAlertsTemplatesAndScreenReaderChrome(t *testing.T) {
	html := testutil.ReadFixture(t, "regression/github_repo_shell.html")
	cleaned, err := Clean(html, true, nil, nil)
	if err != nil {
		t.Fatalf("clean failed: %v", err)
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
		if strings.Contains(cleaned, unwanted) {
			t.Fatalf("expected %q to be removed", unwanted)
		}
	}
	if !strings.Contains(cleaned, "A terminal workspace with batteries included.") {
		t.Fatalf("expected repository summary to remain")
	}
}

func TestCleanDropsGlobalShellChromeInFullContentMode(t *testing.T) {
	html := testutil.ReadFixture(t, "regression/github_repo_shell.html")
	cleaned, err := Clean(html, false, nil, nil)
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	for _, unwanted := range []string{
		"Skip to content",
		"You signed in with another tab or window",
		"Dismiss alert",
		"{{ className }}",
		"Uh oh!",
		"Please reload this page",
	} {
		if strings.Contains(cleaned, unwanted) {
			t.Fatalf("expected %q to be removed in full-content mode", unwanted)
		}
	}
	if !strings.Contains(cleaned, "Pull requests") {
		t.Fatalf("expected full-content mode to preserve visible page chrome")
	}
}

func TestCleanPrefersArticleRootOverBroaderMainShell(t *testing.T) {
	html := `
	<html>
	  <body>
	    <main>
	      <section><h2>Top Stories</h2><p>Promo shell</p></section>
	      <article>
	        <h1>Actual Story</h1>
	        <p>The article body should win over the broader main wrapper.</p>
	      </article>
	    </main>
	  </body>
	</html>`

	cleaned, err := Clean(html, true, nil, nil)
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	if strings.Contains(cleaned, "Top Stories") {
		t.Fatalf("expected broader main shell content to be dropped when article root is available")
	}
	if !strings.Contains(cleaned, "Actual Story") {
		t.Fatalf("expected article content to remain")
	}
}

func TestCleanDropsTabsToolbarSkipLinksAndLoadingShell(t *testing.T) {
	html := `
	<html>
	  <body>
	    <a href="#skipToMainContent" aria-label="Skip to main content">Skip to main content</a>
	    <div role="tablist"><button role="tab">Article</button><button role="tab">Talk</button></div>
	    <div role="toolbar"><button>Notifications</button></div>
	    <div class="cookie-banner">Cookie settings</div>
	    <div class="loading-state">Loading...</div>
	    <main><h1>Real content</h1><p>Keep this.</p></main>
	  </body>
	</html>`

	cleaned, err := Clean(html, true, nil, nil)
	if err != nil {
		t.Fatalf("clean failed: %v", err)
	}

	for _, unwanted := range []string{
		"Skip to main content",
		"Article",
		"Talk",
		"Notifications",
		"Cookie settings",
		"Loading...",
	} {
		if strings.Contains(cleaned, unwanted) {
			t.Fatalf("expected %q to be removed", unwanted)
		}
	}
	if !strings.Contains(cleaned, "Real content") {
		t.Fatalf("expected main content to remain")
	}
}
