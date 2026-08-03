package convert

import (
	"context"
	"strings"
	"testing"
)

func TestSharedPoliciesApplyToNativeConversion(t *testing.T) {
	input := `<article><button>Copy</button><span role="button">Share</span><p>Follow</p><p>Body text.</p></article>`

	markdown, err := ToMarkdownContext(context.Background(), input)
	if err != nil {
		t.Fatalf("convert with Native converter: %v", err)
	}
	for _, unwanted := range []string{"Copy", "Share", "Follow"} {
		if strings.Contains(markdown, unwanted) {
			t.Fatalf("expected shared policies to remove %q from Native output:\n%s", unwanted, markdown)
		}
	}
	if !strings.Contains(markdown, "Body text.") {
		t.Fatalf("expected body content in Native output:\n%s", markdown)
	}
}

func TestDefaultPolicyNamesRemainDeterministic(t *testing.T) {
	leftHTML := defaultHTMLPolicyRegistrations()
	rightHTML := defaultHTMLPolicyRegistrations()
	if policyNames(leftHTML) != policyNames(rightHTML) {
		t.Fatalf("HTML policy names are not deterministic: %q != %q", policyNames(leftHTML), policyNames(rightHTML))
	}

	leftMarkdown := defaultMarkdownPolicyRegistrations()
	rightMarkdown := defaultMarkdownPolicyRegistrations()
	if markdownPolicyNames(leftMarkdown) != markdownPolicyNames(rightMarkdown) {
		t.Fatalf("Markdown policy names are not deterministic: %q != %q", markdownPolicyNames(leftMarkdown), markdownPolicyNames(rightMarkdown))
	}
}

func policyNames(registrations []htmlPolicyRegistration) string {
	names := make([]string, 0, len(registrations))
	for _, registration := range registrations {
		names = append(names, registration.Name)
	}
	return strings.Join(names, ",")
}

func markdownPolicyNames(registrations []markdownPolicyRegistration) string {
	names := make([]string, 0, len(registrations))
	for _, registration := range registrations {
		names = append(names, registration.Name)
	}
	return strings.Join(names, ",")
}
