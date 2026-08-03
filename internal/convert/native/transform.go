package native

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/net/html"
)

// DocumentTransform performs a deterministic DOM rewrite before the Native
// parser builds its intermediate representation. Implementations must be safe
// for concurrent use and must honor context cancellation.
type DocumentTransform interface {
	Name() string
	Transform(ctx context.Context, root *html.Node) error
}

type documentTransform struct {
	name      string
	transform func(context.Context, *html.Node) error
}

func (t documentTransform) Name() string {
	return t.name
}

func (t documentTransform) Transform(ctx context.Context, root *html.Node) error {
	return t.transform(ctx, root)
}

// NewDocumentTransform creates an immutable named DOM transform.
func NewDocumentTransform(name string, transform func(context.Context, *html.Node) error) (DocumentTransform, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("native document transform name is empty")
	}
	if transform == nil {
		return nil, fmt.Errorf("native document transform %q is nil", name)
	}
	return documentTransform{name: name, transform: transform}, nil
}

func validateDocumentTransforms(transforms []DocumentTransform) ([]DocumentTransform, error) {
	validated := make([]DocumentTransform, 0, len(transforms))
	seen := make(map[string]struct{}, len(transforms))
	for _, transform := range transforms {
		if isNilDocumentTransform(transform) {
			return nil, errors.New("native document transform is nil")
		}
		name := strings.TrimSpace(transform.Name())
		if name == "" {
			return nil, errors.New("native document transform name is empty")
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("native document transform %q is registered more than once", name)
		}
		seen[key] = struct{}{}
		validated = append(validated, transform)
	}
	return validated, nil
}

func isNilDocumentTransform(transform DocumentTransform) bool {
	if transform == nil {
		return true
	}
	reflected := reflect.ValueOf(transform)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func applyDocumentTransforms(ctx context.Context, root *html.Node, transforms []DocumentTransform) error {
	for _, transform := range transforms {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := transform.Transform(ctx, root); err != nil {
			return fmt.Errorf("native document transform %q: %w", transform.Name(), err)
		}
	}
	return nil
}
