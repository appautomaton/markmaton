package resolve

import (
	"strings"
	"testing"
)

func TestResolveHTMLMakesURLsAbsolute(t *testing.T) {
	html := `<p><a href="/docs">Docs</a><img src="/img/a.jpg" srcset="/img/a.jpg 1x, /img/b.jpg 2x"></p>`
	resolved, err := ResolveHTML(html, "https://example.com/guide/start")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if !strings.Contains(resolved, `href="https://example.com/docs"`) {
		t.Fatalf("expected anchor href to be absolute: %s", resolved)
	}
	if !strings.Contains(resolved, `src="https://example.com/img/b.jpg"`) {
		t.Fatalf("expected srcset to select the largest image: %s", resolved)
	}
}

func TestResolveHTMLUsesBaseHrefWhenPresent(t *testing.T) {
	html := `<base href="https://cdn.example.com/assets/"><img src="cover.png">`
	resolved, err := ResolveHTML(html, "https://example.com/page")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if !strings.Contains(resolved, `src="https://cdn.example.com/assets/cover.png"`) {
		t.Fatalf("expected image src to use base href: %s", resolved)
	}
}
