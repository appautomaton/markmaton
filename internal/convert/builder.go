package convert

import md "github.com/firecrawl/html-to-markdown"

type BeforeHookRegistration struct {
	Name string
	Hook md.BeforeHook
}

type AfterHookRegistration struct {
	Name string
	Hook md.Afterhook
}

type PluginRegistration struct {
	Name   string
	Plugin md.Plugin
}

type Builder struct {
	domain           string
	enableCommonMark bool
	options          *md.Options
	beforeHooks      []BeforeHookRegistration
	afterHooks       []AfterHookRegistration
	plugins          []PluginRegistration
	rules            []md.Rule
	keepTags         []string
	removeTags       []string
}

func NewBuilder(domain string) *Builder {
	return &Builder{
		domain:           domain,
		enableCommonMark: true,
		beforeHooks:      DefaultBeforeHookRegistrations(),
		afterHooks:       DefaultAfterHookRegistrations(),
		plugins:          DefaultPluginRegistrations(),
	}
}

func DefaultBuilder(domain string) *Builder {
	return NewBuilder(domain)
}

func (b *Builder) WithOptions(options *md.Options) *Builder {
	if options == nil {
		b.options = nil
		return b
	}

	copied := *options
	b.options = &copied
	return b
}

func (b *Builder) WithBeforeHooks(hooks ...BeforeHookRegistration) *Builder {
	b.beforeHooks = append(b.beforeHooks, hooks...)
	return b
}

func (b *Builder) WithAfterHooks(hooks ...AfterHookRegistration) *Builder {
	b.afterHooks = append(b.afterHooks, hooks...)
	return b
}

func (b *Builder) WithPlugins(plugins ...PluginRegistration) *Builder {
	b.plugins = append(b.plugins, plugins...)
	return b
}

func (b *Builder) WithRules(rules ...md.Rule) *Builder {
	b.rules = append(b.rules, rules...)
	return b
}

func (b *Builder) Keep(tags ...string) *Builder {
	b.keepTags = append(b.keepTags, tags...)
	return b
}

func (b *Builder) Remove(tags ...string) *Builder {
	b.removeTags = append(b.removeTags, tags...)
	return b
}

func (b *Builder) PluginNames() []string {
	names := make([]string, 0, len(b.plugins))
	for _, registration := range b.plugins {
		names = append(names, registration.Name)
	}
	return names
}

func (b *Builder) BeforeHookNames() []string {
	names := make([]string, 0, len(b.beforeHooks))
	for _, registration := range b.beforeHooks {
		names = append(names, registration.Name)
	}
	return names
}

func (b *Builder) AfterHookNames() []string {
	names := make([]string, 0, len(b.afterHooks))
	for _, registration := range b.afterHooks {
		names = append(names, registration.Name)
	}
	return names
}

func (b *Builder) Build() *md.Converter {
	converter := md.NewConverter(b.domain, b.enableCommonMark, cloneOptions(b.options))

	for _, registration := range b.beforeHooks {
		converter.Before(registration.Hook)
	}
	for _, registration := range b.afterHooks {
		converter.After(registration.Hook)
	}
	for _, registration := range b.plugins {
		converter.Use(registration.Plugin)
	}
	if len(b.rules) > 0 {
		converter.AddRules(b.rules...)
	}
	if len(b.keepTags) > 0 {
		converter.Keep(b.keepTags...)
	}
	if len(b.removeTags) > 0 {
		converter.Remove(b.removeTags...)
	}

	return converter
}

func cloneOptions(options *md.Options) *md.Options {
	if options == nil {
		return nil
	}
	copied := *options
	return &copied
}
