package main

import (
	"strings"
	"testing"
)

func TestReadRequestDefaultsToMainContentWhenOptionOmitted(t *testing.T) {
	request, err := readRequest(strings.NewReader(`{"html":"<p>Hello</p>"}`))
	if err != nil {
		t.Fatalf("readRequest failed: %v", err)
	}

	if !request.Options.UseOnlyMainContent() {
		t.Fatalf("expected omitted only_main_content to default to true")
	}
}

func TestReadRequestHonorsExplicitFalseOnlyMainContent(t *testing.T) {
	request, err := readRequest(strings.NewReader(`{"html":"<p>Hello</p>","options":{"only_main_content":false}}`))
	if err != nil {
		t.Fatalf("readRequest failed: %v", err)
	}

	if request.Options.UseOnlyMainContent() {
		t.Fatalf("expected explicit only_main_content=false to be preserved")
	}
}
