package native

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	defaultMaxInputBytes  = 64 << 20
	defaultMaxOutputBytes = 128 << 20
	defaultMaxNodes       = 1_000_000
	defaultMaxDepth       = 2_048
)

var (
	ErrInputLimit  = errors.New("native converter input limit exceeded")
	ErrOutputLimit = errors.New("native converter output limit exceeded")
	ErrNodeLimit   = errors.New("native converter node limit exceeded")
	ErrDepthLimit  = errors.New("native converter depth limit exceeded")
)

type Limits struct {
	MaxInputBytes  int
	MaxOutputBytes int
	MaxNodes       int
	MaxDepth       int
}

type Config struct {
	Limits             Limits
	DocumentTransforms []DocumentTransform
}

type Converter struct {
	limits             Limits
	documentTransforms []DocumentTransform
}

func New() *Converter {
	return NewWithLimits(Limits{})
}

func NewWithLimits(limits Limits) *Converter {
	return &Converter{limits: normalizeLimits(limits)}
}

func NewWithConfig(config Config) (*Converter, error) {
	transforms, err := validateDocumentTransforms(config.DocumentTransforms)
	if err != nil {
		return nil, err
	}
	return &Converter{
		limits:             normalizeLimits(config.Limits),
		documentTransforms: transforms,
	}, nil
}

func normalizeLimits(limits Limits) Limits {
	if limits.MaxInputBytes <= 0 {
		limits.MaxInputBytes = defaultMaxInputBytes
	}
	if limits.MaxOutputBytes <= 0 {
		limits.MaxOutputBytes = defaultMaxOutputBytes
	}
	if limits.MaxNodes <= 0 {
		limits.MaxNodes = defaultMaxNodes
	}
	if limits.MaxDepth <= 0 {
		limits.MaxDepth = defaultMaxDepth
	}
	return limits
}

func (c *Converter) Convert(ctx context.Context, source string) (string, error) {
	if c == nil {
		return "", errors.New("native converter is nil")
	}
	if ctx == nil {
		return "", errors.New("native converter context is nil")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if len(source) > c.limits.MaxInputBytes {
		return "", fmt.Errorf("%w: got %d bytes, maximum is %d", ErrInputLimit, len(source), c.limits.MaxInputBytes)
	}
	if strings.TrimSpace(source) == "" {
		return "", nil
	}

	parsed, err := parseDocument(ctx, source, c.limits, c.documentTransforms)
	if err != nil {
		return "", err
	}
	markdown, err := renderDocument(ctx, parsed)
	if err != nil {
		return "", err
	}
	if len(markdown) > c.limits.MaxOutputBytes {
		return "", fmt.Errorf("%w: got %d bytes, maximum is %d", ErrOutputLimit, len(markdown), c.limits.MaxOutputBytes)
	}
	return markdown, nil
}
