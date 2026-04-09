package convert

import (
	md "github.com/firecrawl/html-to-markdown"
	"github.com/firecrawl/html-to-markdown/plugin"
)

func DefaultPluginRegistrations() []PluginRegistration {
	return []PluginRegistration{
		newPluginRegistration("github_flavored", plugin.GitHubFlavored()),
		newPluginRegistration("robust_code_block", plugin.RobustCodeBlock()),
	}
}

func newPluginRegistration(name string, plug md.Plugin) PluginRegistration {
	return PluginRegistration{
		Name:   name,
		Plugin: plug,
	}
}
