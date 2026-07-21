package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sirrobot01/unifydoc/internal/config"
	"gopkg.in/yaml.v3"
)

func TestConvertPostman(t *testing.T) {
	data := []byte(`{
	  "info": {"name": "Shop API"},
	  "item": [
	    {"name": "List products", "request": {"method": "get", "url": {"path": ["products"]}, "description": "all products"}},
	    {"name": "Orders", "item": [
	      {"name": "Create order", "request": {"method": "POST", "url": "https://api.shop.com/orders?x=1", "description": {"content": "make order"}}}
	    ]}
	  ]
	}`)

	out, title, err := convertPostman(data)
	if err != nil {
		t.Fatalf("convertPostman error: %v", err)
	}
	if title != "Shop API" {
		t.Errorf("title = %q, want Shop API", title)
	}

	var spec apiSpecFile
	if err := yaml.Unmarshal(out, &spec); err != nil {
		t.Fatalf("output is not valid YAML: %v", err)
	}
	if len(spec.API.Endpoints) != 2 {
		t.Fatalf("endpoints = %d, want 2", len(spec.API.Endpoints))
	}

	list := spec.API.Endpoints[0]
	if list.Method != "GET" || list.Path != "/products" {
		t.Errorf("list endpoint wrong: %+v", list)
	}
	if list.Description != "all products" {
		t.Errorf("description = %q", list.Description)
	}

	create := spec.API.Endpoints[1]
	if create.Method != "POST" || create.Path != "/orders" {
		t.Errorf("create endpoint wrong: %+v", create)
	}
	if len(create.Tags) != 1 || create.Tags[0] != "Orders" {
		t.Errorf("folder should become a tag: %+v", create.Tags)
	}
	if create.Description != "make order" {
		t.Errorf("object description not parsed: %q", create.Description)
	}
}

func TestConvertPostmanInvalid(t *testing.T) {
	if _, _, err := convertPostman([]byte("not json")); err == nil {
		t.Error("expected error on invalid JSON")
	}
	if _, _, err := convertPostman([]byte("{}")); err == nil {
		t.Error("expected error on non-collection JSON")
	}
}

func TestStripHost(t *testing.T) {
	tests := map[string]string{
		"https://api.example.com/v1/users?x=1": "/v1/users",
		"http://host/a/b":                      "/a/b",
		"{{baseUrl}}/orders":                   "/orders",
		"/already/a/path":                      "/already/a/path",
	}
	for in, want := range tests {
		if got := stripHost(in); got != want {
			t.Errorf("stripHost(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"Shop API":    "shop-api",
		"  Weird!!  ": "weird",
		"":            "imported",
	}
	for in, want := range tests {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildFrameworkConfig(t *testing.T) {
	cfg, err := buildFrameworkConfig("gin")
	if err != nil {
		t.Fatalf("buildFrameworkConfig error: %v", err)
	}
	if len(cfg.Protocols) != 1 || cfg.Protocols[0].Plugin != "openapi" {
		t.Fatalf("expected one openapi protocol: %+v", cfg.Protocols)
	}
	if cfg.Protocols[0].Spec != "./docs/swagger.yaml" {
		t.Errorf("gin spec path = %q", cfg.Protocols[0].Spec)
	}
	if _, err := buildFrameworkConfig("django"); err == nil {
		t.Error("unknown framework should error")
	}
}

func TestUpsertProtocol(t *testing.T) {
	cfg := config.DefaultConfig()
	upsertProtocol(cfg, config.ProtocolConfig{Plugin: "openapi", Spec: "a.yaml", Enabled: true})
	upsertProtocol(cfg, config.ProtocolConfig{Plugin: "grpc", Spec: "b.proto", Enabled: true})
	if len(cfg.Protocols) != 2 {
		t.Fatalf("expected 2 protocols, got %d", len(cfg.Protocols))
	}
	// Same plugin+spec replaces rather than appends.
	upsertProtocol(cfg, config.ProtocolConfig{Plugin: "openapi", Spec: "a.yaml", Enabled: false})
	if len(cfg.Protocols) != 2 {
		t.Fatalf("upsert should not append duplicate, got %d", len(cfg.Protocols))
	}
	if cfg.Protocols[0].Enabled {
		t.Error("existing entry should have been updated to disabled")
	}
}

func TestWriteCIWorkflowAndHooks(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	// CI workflow
	path, err := writeCIWorkflow("github")
	if err != nil {
		t.Fatalf("writeCIWorkflow error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("workflow file missing: %v", err)
	}
	if _, err := writeCIWorkflow("gitlab"); err == nil {
		t.Error("unsupported provider should error")
	}

	// Hooks require a .git directory.
	if _, err := installGitHooks(); err == nil {
		t.Error("installGitHooks should fail without .git")
	}
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatal(err)
	}
	written, err := installGitHooks()
	if err != nil {
		t.Fatalf("installGitHooks error: %v", err)
	}
	if len(written) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(written))
	}
	for _, h := range written {
		info, err := os.Stat(h)
		if err != nil {
			t.Errorf("hook missing: %v", err)
			continue
		}
		// Windows filesystems do not carry the Unix executable bit (and git on
		// Windows does not require it), so only assert it elsewhere.
		if runtime.GOOS != "windows" && info.Mode().Perm()&0100 == 0 {
			t.Errorf("hook %s is not executable", filepath.Base(h))
		}
	}
}

// chdir changes to dir and returns a function that restores the previous cwd.
func chdir(t *testing.T, dir string) func() {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return func() { _ = os.Chdir(prev) }
}
