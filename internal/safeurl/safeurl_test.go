package safeurl

import "testing"

func TestLinkAndResourceSchemes(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		link     bool
		resource bool
	}{
		{name: "absolute HTTPS", value: "https://example.com/path", link: true, resource: true},
		{name: "relative", value: "../path", link: true, resource: true},
		{name: "fragment", value: "#section", link: true, resource: true},
		{name: "email", value: "mailto:reader@example.com", link: true, resource: false},
		{name: "telephone", value: "tel:+12025550123", link: true, resource: false},
		{name: "JavaScript", value: "javascript:alert(1)", link: false, resource: false},
		{name: "data", value: "data:text/html,hello", link: false, resource: false},
		{name: "VBScript", value: "vbscript:msgbox(1)", link: false, resource: false},
		{name: "blank", value: "  ", link: false, resource: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Link(test.value); got != test.link {
				t.Fatalf("Link(%q) = %v, want %v", test.value, got, test.link)
			}
			if got := Resource(test.value); got != test.resource {
				t.Fatalf("Resource(%q) = %v, want %v", test.value, got, test.resource)
			}
		})
	}
}

func TestWebBaseRequiresHTTPHost(t *testing.T) {
	for _, value := range []string{"https://example.com/base/", "http://example.com"} {
		if !WebBase(value) {
			t.Fatalf("expected %q to be a web base", value)
		}
	}
	for _, value := range []string{"", "/relative", "javascript:alert(1)", "file:///tmp/page.html"} {
		if WebBase(value) {
			t.Fatalf("did not expect %q to be a web base", value)
		}
	}
}
