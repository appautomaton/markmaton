package safeurl

import (
	"net/url"
	"strings"
)

// Link reports whether value is safe to expose as a Markdown or HTML link.
// Relative references are allowed because the resolver may absolutize them
// later when a trustworthy page URL is available.
func Link(value string) bool {
	scheme, ok := parsedScheme(value)
	if !ok {
		return false
	}
	switch scheme {
	case "", "http", "https", "mailto", "tel":
		return true
	default:
		return false
	}
}

// Resource reports whether value is safe to expose as an embedded resource.
func Resource(value string) bool {
	scheme, ok := parsedScheme(value)
	if !ok {
		return false
	}
	switch scheme {
	case "", "http", "https":
		return true
	default:
		return false
	}
}

// WebBase reports whether value may be used as a base URL for resolving
// document-relative references.
func WebBase(value string) bool {
	value = strings.TrimSpace(value)
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return true
	default:
		return false
	}
}

func parsedScheme(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", false
	}
	return strings.ToLower(parsed.Scheme), true
}
