package mcp

import (
	"testing"

	"github.com/sirrobot01/unifydoc/internal/ir"
)

const spec = `
mcp:
  title: Shop MCP
  version: 1.2.0
  tools:
    - name: create_order
      description: Create an order
      inputSchema:
        type: object
        required: [sku]
        properties:
          sku:
            type: string
      example:
        sku: tee-black
  resources:
    - name: recent orders
      uri: orders://recent
      description: Recent orders
      mimeType: application/json
  prompts:
    - name: summarize
      description: Summarize an order
      template: "Summarize {{order}}"
      parameters:
        - name: order
          description: order id
          required: true
          type: string
`

func TestMetadata(t *testing.T) {
	p := NewPlugin()
	if p.Name() != "mcp" {
		t.Errorf("Name() = %q", p.Name())
	}
	if p.Version() == "" {
		t.Error("Version() empty")
	}
}

func TestValidate(t *testing.T) {
	p := NewPlugin()
	if err := p.Validate([]byte(spec)); err != nil {
		t.Errorf("valid spec rejected: %v", err)
	}
	if err := p.Validate([]byte("\tbad: [")); err == nil {
		t.Error("malformed YAML should fail")
	}
}

func TestParseToolsResourcesPrompts(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte(spec))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.Title != "Shop MCP" || result.Version != "1.2.0" {
		t.Errorf("header wrong: %q %q", result.Title, result.Version)
	}
	// tool + resource + prompt
	if len(result.Resources) != 3 {
		t.Fatalf("resources = %d, want 3", len(result.Resources))
	}

	tool := byType(t, result, "mcp-tool")
	if tool.Method != "TOOL" {
		t.Errorf("tool method = %q", tool.Method)
	}
	if tool.Request == nil || tool.Request.Properties["sku"] == nil {
		t.Errorf("inputSchema not converted: %+v", tool.Request)
	}

	res := byType(t, result, "mcp-resource")
	if res.Method != "RESOURCE" || res.Path != "orders://recent" {
		t.Errorf("resource wrong: %q %q", res.Method, res.Path)
	}
	if res.Metadata["mime_type"] != "application/json" {
		t.Errorf("mime_type = %v", res.Metadata["mime_type"])
	}

	prompt := byType(t, result, "mcp-prompt")
	if prompt.Method != "PROMPT" {
		t.Errorf("prompt method = %q", prompt.Method)
	}
	if len(prompt.Parameters) != 1 || !prompt.Parameters[0].Required {
		t.Errorf("prompt params wrong: %+v", prompt.Parameters)
	}
	if prompt.Metadata["template"] == "" {
		t.Error("prompt template metadata missing")
	}
}

func byType(t *testing.T, result *ir.IR, typ string) *ir.Resource {
	t.Helper()
	for i := range result.Resources {
		if result.Resources[i].Metadata["type"] == typ {
			return &result.Resources[i]
		}
	}
	t.Fatalf("no resource with type %q", typ)
	return nil
}
