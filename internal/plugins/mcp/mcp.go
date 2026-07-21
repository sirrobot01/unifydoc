package mcp

import (
	"fmt"

	"github.com/sirrobot01/unifydoc/internal/ir"
	"gopkg.in/yaml.v3"
)

// Plugin implements the MCP (Model Context Protocol) plugin
type Plugin struct{}

// NewPlugin creates a new MCP plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "mcp"
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Spec represents the MCP specification
type Spec struct {
	MCP Config `yaml:"mcp" json:"mcp"`
}

// Config represents MCP configuration
type Config struct {
	Title       string     `yaml:"title,omitempty" json:"title,omitempty"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string     `yaml:"version,omitempty" json:"version,omitempty"`
	Tools       []Tool     `yaml:"tools,omitempty" json:"tools,omitempty"`
	Resources   []Resource `yaml:"resources,omitempty" json:"resources,omitempty"`
	Prompts     []Prompt   `yaml:"prompts,omitempty" json:"prompts,omitempty"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	InputSchema map[string]interface{} `yaml:"inputSchema,omitempty" json:"inputSchema,omitempty"`
	Example     interface{}            `yaml:"example,omitempty" json:"example,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	Name        string                 `yaml:"name" json:"name"`
	URI         string                 `yaml:"uri" json:"uri"`
	Description string                 `yaml:"description" json:"description"`
	MimeType    string                 `yaml:"mimeType,omitempty" json:"mimeType,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// Prompt represents an MCP prompt template
type Prompt struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Template    string            `yaml:"template" json:"template"`
	Parameters  []PromptParameter `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// PromptParameter represents a prompt parameter
type PromptParameter struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Required    bool   `yaml:"required" json:"required"`
	Type        string `yaml:"type" json:"type"`
}

// Validate validates an MCP specification
func (p *Plugin) Validate(spec []byte) error {
	var mcpSpec Spec
	if err := yaml.Unmarshal(spec, &mcpSpec); err != nil {
		return fmt.Errorf("invalid MCP YAML: %w", err)
	}

	return nil
}

// Parse parses an MCP specification and converts it to IR
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	var mcpSpec Spec
	if err := yaml.Unmarshal(spec, &mcpSpec); err != nil {
		return nil, fmt.Errorf("failed to parse MCP spec: %w", err)
	}

	result := ir.NewIR("mcp")
	result.Title = mcpSpec.MCP.Title
	if result.Title == "" {
		result.Title = "Model Context Protocol"
	}
	result.Description = mcpSpec.MCP.Description
	result.Version = mcpSpec.MCP.Version
	if result.Version == "" {
		result.Version = "1.0.0"
	}

	// Parse tools as resources
	for _, tool := range mcpSpec.MCP.Tools {
		resource := ir.Resource{
			Name:        tool.Name,
			Path:        fmt.Sprintf("/tools/%s", tool.Name),
			Method:      "TOOL",
			Description: tool.Description,
			Examples:    make([]ir.Example, 0),
			Metadata:    make(map[string]interface{}),
			Tags:        []string{"Tools"},
		}

		resource.Metadata["type"] = "mcp-tool"

		if tool.InputSchema != nil {
			resource.Request = convertMapToSchema(tool.InputSchema)
		}

		if tool.Example != nil {
			resource.Examples = append(resource.Examples, ir.Example{
				Name:  "Tool Input Example",
				Value: tool.Example,
			})
		}

		result.Resources = append(result.Resources, resource)
	}

	// Parse resources
	for _, res := range mcpSpec.MCP.Resources {
		resource := ir.Resource{
			Name:        res.Name,
			Path:        res.URI,
			Method:      "RESOURCE",
			Description: res.Description,
			Metadata:    res.Metadata,
			Tags:        []string{"Resources"},
		}

		if resource.Metadata == nil {
			resource.Metadata = make(map[string]interface{})
		}
		resource.Metadata["type"] = "mcp-resource"
		resource.Metadata["mime_type"] = res.MimeType

		result.Resources = append(result.Resources, resource)
	}

	// Parse prompts
	for _, prompt := range mcpSpec.MCP.Prompts {
		resource := ir.Resource{
			Name:        prompt.Name,
			Path:        fmt.Sprintf("/prompts/%s", prompt.Name),
			Method:      "PROMPT",
			Description: prompt.Description,
			Parameters:  make([]ir.Parameter, 0),
			Metadata:    make(map[string]interface{}),
			Tags:        []string{"Prompts"},
		}

		resource.Metadata["type"] = "mcp-prompt"
		resource.Metadata["template"] = prompt.Template

		for _, param := range prompt.Parameters {
			resource.Parameters = append(resource.Parameters, ir.Parameter{
				Name:        param.Name,
				Description: param.Description,
				Required:    param.Required,
				Type:        param.Type,
			})
		}

		result.Resources = append(result.Resources, resource)
	}

	return result, nil
}

// convertMapToSchema converts a map to IR Schema
func convertMapToSchema(data map[string]interface{}) *ir.Schema {
	if data == nil {
		return nil
	}

	schema := &ir.Schema{
		Type:       "object",
		Properties: make(map[string]*ir.Schema),
		Metadata:   make(map[string]interface{}),
	}

	if typ, ok := data["type"].(string); ok {
		schema.Type = typ
	}

	if desc, ok := data["description"].(string); ok {
		schema.Description = desc
	}

	if props, ok := data["properties"].(map[string]interface{}); ok {
		for name, propData := range props {
			if propMap, ok := propData.(map[string]interface{}); ok {
				schema.Properties[name] = convertMapToSchema(propMap)
			}
		}
	}

	if required, ok := data["required"].([]interface{}); ok {
		schema.Required = make([]string, 0)
		for _, r := range required {
			if rStr, ok := r.(string); ok {
				schema.Required = append(schema.Required, rStr)
			}
		}
	}

	return schema
}

// GetTemplate returns the custom template
func (p *Plugin) GetTemplate() string {
	return ""
}
