package cli

import (
	"testing"

	"github.com/sirrobot01/unifydoc/internal/plugin"
)

// TestRegisterPluginsCompliance verifies that every built-in plugin registers
// cleanly and satisfies the basic Plugin contract (non-empty identity).
func TestRegisterPluginsCompliance(t *testing.T) {
	manager := plugin.NewManager()
	if err := registerPlugins(manager); err != nil {
		t.Fatalf("registerPlugins() failed: %v", err)
	}

	// All 9 built-in plugins from the PRD should be registered.
	wantPlugins := []string{
		"openapi", "grpc", "websocket", "webhook",
		"events", "mcp", "api", "asyncapi", "custom",
	}

	metas := manager.List()
	if len(metas) < len(wantPlugins) {
		t.Errorf("registered %d plugins, want at least %d", len(metas), len(wantPlugins))
	}

	for _, name := range wantPlugins {
		t.Run(name, func(t *testing.T) {
			p, err := manager.Get(name)
			if err != nil {
				t.Fatalf("plugin %q not registered: %v", name, err)
			}
			if p.Name() != name {
				t.Errorf("Name() = %q, want %q", p.Name(), name)
			}
			if p.Version() == "" {
				t.Errorf("plugin %q has empty Version()", name)
			}
		})
	}
}

// TestPluginsParseExamples runs each protocol plugin against its example spec
// to guard the end-to-end parse path from regressions.
func TestPluginsParseExamples(t *testing.T) {
	manager := plugin.NewManager()
	if err := registerPlugins(manager); err != nil {
		t.Fatalf("registerPlugins() failed: %v", err)
	}

	cases := []struct {
		plugin string
		spec   string
	}{
		{"openapi", "../../examples/openapi/petstore.yaml"},
		{"websocket", "../../examples/websocket/chat.yaml"},
		{"grpc", "../../examples/grpc/user.proto"},
	}

	for _, tc := range cases {
		t.Run(tc.plugin, func(t *testing.T) {
			p, err := manager.Get(tc.plugin)
			if err != nil {
				t.Fatalf("plugin %q not found: %v", tc.plugin, err)
			}

			data := readExample(t, tc.spec)
			result, err := p.Parse(data)
			if err != nil {
				t.Fatalf("Parse() failed for %s: %v", tc.spec, err)
			}
			if result == nil {
				t.Fatal("Parse() returned nil IR")
			}
			if result.Protocol != tc.plugin {
				t.Errorf("Protocol = %q, want %q", result.Protocol, tc.plugin)
			}
			if len(result.Resources) == 0 {
				t.Errorf("expected at least one resource parsed from %s", tc.spec)
			}
		})
	}
}
