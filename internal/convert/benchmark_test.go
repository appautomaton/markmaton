package convert

import (
	"testing"

	"github.com/appautomaton/markmaton/internal/testutil"
)

var benchmarkMarkdown string

func BenchmarkConversion(b *testing.B) {
	fixtures := []struct {
		name   string
		source func(testing.TB) string
	}{
		{
			name: "small_document",
			source: func(testing.TB) string {
				return `<article><h1>Title</h1><p>Hello <strong>world</strong>. Read <a href="https://example.com/docs">the docs</a>.</p><ul><li>One</li><li>Two</li></ul></article>`
			},
		},
		{
			name: "real_world_page",
			source: func(tb testing.TB) string {
				return testutil.ReadFixture(tb, "compatibility/real-world/go-blog.html")
			},
		},
		{
			name: "large_mixed_html",
			source: func(tb testing.TB) string {
				return testutil.ReadFixture(tb, "performance/mixed.html")
			},
		},
		{
			name: "forty_thousand_item_list",
			source: func(tb testing.TB) string {
				return testutil.ReadFixture(tb, "performance/large-list.html")
			},
		},
		{
			name: "five_thousand_by_twelve_table",
			source: func(tb testing.TB) string {
				return testutil.ReadFixture(tb, "performance/large-table.html")
			},
		},
	}

	for _, fixture := range fixtures {
		fixture := fixture
		b.Run(fixture.name, func(b *testing.B) {
			source := fixture.source(b)
			b.ReportAllocs()
			b.SetBytes(int64(len(source)))
			b.ResetTimer()
			for iteration := 0; iteration < b.N; iteration++ {
				markdown, err := ToMarkdown(source)
				if err != nil {
					b.Fatalf("convert benchmark fixture: %v", err)
				}
				benchmarkMarkdown = markdown
			}
			b.StopTimer()
			b.ReportMetric(float64(len(benchmarkMarkdown)), "output_B")
		})
	}
}
