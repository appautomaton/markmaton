package resolve

import (
	"strings"
	"testing"
)

func TestResolveHTMLMakesURLsAbsolute(t *testing.T) {
	html := `<p><a href="/docs">Docs</a><img src="/img/a.jpg" srcset="/img/a.jpg 1x, /img/b.jpg 2x"><iframe src="/embed/demo" title="Demo"></iframe><video src="/media/demo.mp4"></video><audio><source src="/media/demo.mp3"></audio></p>`
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
	if !strings.Contains(resolved, `src="https://example.com/embed/demo"`) {
		t.Fatalf("expected iframe src to be absolute: %s", resolved)
	}
	if !strings.Contains(resolved, `src="https://example.com/media/demo.mp4"`) {
		t.Fatalf("expected video src to be absolute: %s", resolved)
	}
	if !strings.Contains(resolved, `src="https://example.com/media/demo.mp3"`) {
		t.Fatalf("expected audio source src to be absolute: %s", resolved)
	}
}

func TestResolveHTMLDropsUnsafeDestinationsAndBaseURLs(t *testing.T) {
	html := `<base href="javascript:alert(1)">
	<a href="javascript:alert(1)">Run</a>
	<a href="data:text/html,hello">Data</a>
	<a href="mailto:reader@example.com">Mail</a>
	<img src="data:image/svg+xml,bad" alt="Unsafe image">
	<img src="/fallback.jpg" srcset="/small.jpg 1x, javascript:alert(1) 2x" alt="Safe candidate">
	<iframe src="vbscript:msgbox(1)" title="Unsafe frame"></iframe>
	<video src="/media/demo.mp4"></video>`
	resolved, err := ResolveHTML(html, "https://example.com/articles/page")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	for _, unsafe := range []string{"javascript:", "data:text", "data:image", "vbscript:"} {
		if strings.Contains(strings.ToLower(resolved), unsafe) {
			t.Fatalf("expected unsafe destination %q to be removed: %s", unsafe, resolved)
		}
	}
	if !strings.Contains(resolved, `href="mailto:reader@example.com"`) {
		t.Fatalf("expected safe mail link to remain: %s", resolved)
	}
	if !strings.Contains(resolved, `src="https://example.com/media/demo.mp4"`) {
		t.Fatalf("expected relative media source to use the trusted page URL: %s", resolved)
	}
	if !strings.Contains(resolved, `src="https://example.com/small.jpg"`) {
		t.Fatalf("expected the largest safe srcset candidate to be selected: %s", resolved)
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
