package native

import (
	"context"
	"errors"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

type pointerDocumentTransform struct{}

func (*pointerDocumentTransform) Name() string {
	return "pointer_transform"
}

func (*pointerDocumentTransform) Transform(context.Context, *html.Node) error {
	return nil
}

func TestDocumentTransformConvertsRetainedCustomTagExample(t *testing.T) {
	transform := mustDocumentTransform(t, "convert_video_tags", func(ctx context.Context, root *html.Node) error {
		return walkDocument(ctx, root, func(node *html.Node) {
			if node.Type != html.ElementNode || node.Data != "my_video" {
				return
			}
			href := strings.TrimSpace(nodeText(node))
			node.Data = "a"
			node.Attr = []html.Attribute{{Key: "href", Val: href}}
			replaceChildren(node, &html.Node{Type: html.TextNode, Data: "click to watch video"})
		})
	})
	converter, err := NewWithConfig(Config{DocumentTransforms: []DocumentTransform{transform}})
	if err != nil {
		t.Fatalf("create configured converter: %v", err)
	}

	markdown, err := converter.Convert(context.Background(), `<my_video>https://youtu.be/1SoMeViD</my_video>`)
	if err != nil {
		t.Fatalf("convert custom tag: %v", err)
	}
	want := "[click to watch video](https://youtu.be/1SoMeViD)"
	if markdown != want {
		t.Fatalf("unexpected custom-tag output\nwant: %s\ngot:  %s", want, markdown)
	}
}

func TestDocumentTransformConvertsRetainedCustomRuleExample(t *testing.T) {
	transform := mustDocumentTransform(t, "convert_bb_strike", func(ctx context.Context, root *html.Node) error {
		return walkDocument(ctx, root, func(node *html.Node) {
			if node.Type == html.ElementNode && node.Data == "span" && hasClass(node, "bb_strike") {
				node.Data = "del"
				replaceChildren(node, &html.Node{Type: html.TextNode, Data: strings.TrimSpace(nodeText(node))})
			}
		})
	})
	converter, err := NewWithConfig(Config{DocumentTransforms: []DocumentTransform{transform}})
	if err != nil {
		t.Fatalf("create configured converter: %v", err)
	}

	markdown, err := converter.Convert(context.Background(), `Good soundtrack <span class="bb_strike"> and cake</span>.`)
	if err != nil {
		t.Fatalf("convert custom rule: %v", err)
	}
	want := "Good soundtrack ~~and cake~~."
	if markdown != want {
		t.Fatalf("unexpected custom-rule output\nwant: %s\ngot:  %s", want, markdown)
	}
}

func TestDocumentTransformRegistrationIsValidatedAndImmutable(t *testing.T) {
	var typedNil *pointerDocumentTransform
	if _, err := NewWithConfig(Config{DocumentTransforms: []DocumentTransform{typedNil}}); err == nil {
		t.Fatal("expected a typed nil transform to fail")
	}

	first := mustDocumentTransform(t, "first", func(context.Context, *html.Node) error { return nil })
	duplicate := mustDocumentTransform(t, " FIRST ", func(context.Context, *html.Node) error { return nil })
	if _, err := NewWithConfig(Config{DocumentTransforms: []DocumentTransform{first, duplicate}}); err == nil {
		t.Fatal("expected duplicate transform names to fail")
	}
	if _, err := NewDocumentTransform("", func(context.Context, *html.Node) error { return nil }); err == nil {
		t.Fatal("expected an empty transform name to fail")
	}
	if _, err := NewDocumentTransform("nil_transform", nil); err == nil {
		t.Fatal("expected a nil transform function to fail")
	}

	marker := mustDocumentTransform(t, "mark_paragraph", func(ctx context.Context, root *html.Node) error {
		return walkDocument(ctx, root, func(node *html.Node) {
			if node.Type == html.ElementNode && node.Data == "p" {
				replaceChildren(node, &html.Node{Type: html.TextNode, Data: "configured"})
			}
		})
	})
	registrations := []DocumentTransform{marker}
	converter, err := NewWithConfig(Config{DocumentTransforms: registrations})
	if err != nil {
		t.Fatalf("create configured converter: %v", err)
	}
	registrations[0] = first

	markdown, err := converter.Convert(context.Background(), `<p>original</p>`)
	if err != nil {
		t.Fatalf("convert with immutable registration: %v", err)
	}
	if markdown != "configured" {
		t.Fatalf("expected copied transform registrations, got %q", markdown)
	}
}

func TestDocumentTransformErrorsAreNamedAndWrapped(t *testing.T) {
	sentinel := errors.New("transform failed")
	transform := mustDocumentTransform(t, "failing_transform", func(context.Context, *html.Node) error {
		return sentinel
	})
	converter, err := NewWithConfig(Config{DocumentTransforms: []DocumentTransform{transform}})
	if err != nil {
		t.Fatalf("create configured converter: %v", err)
	}

	_, err = converter.Convert(context.Background(), `<p>Hello</p>`)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped transform error, got %v", err)
	}
	if !strings.Contains(err.Error(), `document transform "failing_transform"`) {
		t.Fatalf("expected transform name in error, got %v", err)
	}
}

func mustDocumentTransform(t testing.TB, name string, apply func(context.Context, *html.Node) error) DocumentTransform {
	t.Helper()
	transform, err := NewDocumentTransform(name, apply)
	if err != nil {
		t.Fatalf("create document transform: %v", err)
	}
	return transform
}

func walkDocument(ctx context.Context, root *html.Node, visit func(*html.Node)) error {
	stack := []*html.Node{root}
	for len(stack) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		visit(current)
		for child := current.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, child)
		}
	}
	return nil
}

func nodeText(node *html.Node) string {
	var builder strings.Builder
	stack := []*html.Node{node}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
		}
		for child := current.LastChild; child != nil; child = child.PrevSibling {
			stack = append(stack, child)
		}
	}
	return builder.String()
}

func replaceChildren(parent *html.Node, children ...*html.Node) {
	for child := parent.FirstChild; child != nil; {
		next := child.NextSibling
		parent.RemoveChild(child)
		child = next
	}
	for _, child := range children {
		parent.AppendChild(child)
	}
}

func hasClass(node *html.Node, wanted string) bool {
	for _, attribute := range node.Attr {
		if attribute.Key != "class" {
			continue
		}
		for _, className := range strings.Fields(attribute.Val) {
			if className == wanted {
				return true
			}
		}
	}
	return false
}
