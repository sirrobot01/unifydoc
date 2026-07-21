package custom

import (
	"fmt"

	"github.com/sirrobot01/unifydoc/internal/ir"
	"gopkg.in/yaml.v3"
)

// Plugin implements the Custom plugin for user-defined protocols
type Plugin struct{}

// NewPlugin creates a new Custom plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "custom"
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Spec represents a flexible custom specification format
type Spec struct {
	Protocol    string                 `yaml:"protocol" json:"protocol"`
	Title       string                 `yaml:"title" json:"title"`
	Version     string                 `yaml:"version" json:"version"`
	Description string                 `yaml:"description" json:"description"`
	Servers     []ServerSpec           `yaml:"servers,omitempty" json:"servers,omitempty"`
	Endpoints   []EndpointSpec         `yaml:"endpoints,omitempty" json:"endpoints,omitempty"`
	Resources   []ResourceSpec         `yaml:"resources,omitempty" json:"resources,omitempty"`
	Types       []TypeSpec             `yaml:"types,omitempty" json:"types,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// ServerSpec represents a server endpoint
type ServerSpec struct {
	URL         string            `yaml:"url" json:"url"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Variables   map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// EndpointSpec represents an endpoint or resource
type EndpointSpec struct {
	Name        string                 `yaml:"name" json:"name"`
	Path        string                 `yaml:"path,omitempty" json:"path,omitempty"`
	Method      string                 `yaml:"method,omitempty" json:"method,omitempty"`
	Description string                 `yaml:"description" json:"description"`
	Parameters  []ParameterSpec        `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Request     map[string]interface{} `yaml:"request,omitempty" json:"request,omitempty"`
	Response    map[string]interface{} `yaml:"response,omitempty" json:"response,omitempty"`
	Examples    []ExampleSpec          `yaml:"examples,omitempty" json:"examples,omitempty"`
	Tags        []string               `yaml:"tags,omitempty" json:"tags,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// ResourceSpec is an alias for EndpointSpec
type ResourceSpec EndpointSpec

// ParameterSpec represents a parameter
type ParameterSpec struct {
	Name        string                 `yaml:"name" json:"name"`
	In          string                 `yaml:"in,omitempty" json:"in,omitempty"`
	Type        string                 `yaml:"type" json:"type"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool                   `yaml:"required,omitempty" json:"required,omitempty"`
	Example     interface{}            `yaml:"example,omitempty" json:"example,omitempty"`
	Schema      map[string]interface{} `yaml:"schema,omitempty" json:"schema,omitempty"`
}

// ExampleSpec represents an example
type ExampleSpec struct {
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Value       interface{} `yaml:"value" json:"value"`
}

// TypeSpec represents a type definition
type TypeSpec struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Schema      map[string]interface{} `yaml:"schema" json:"schema"`
}

// Validate validates a custom specification
func (p *Plugin) Validate(spec []byte) error {
	var customSpec Spec
	if err := yaml.Unmarshal(spec, &customSpec); err != nil {
		return fmt.Errorf("invalid custom spec YAML: %w", err)
	}

	if customSpec.Protocol == "" {
		return fmt.Errorf("protocol field is required")
	}

	if customSpec.Title == "" {
		return fmt.Errorf("title field is required")
	}

	return nil
}

// Parse parses a custom specification and converts it to IR
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	var customSpec Spec
	if err := yaml.Unmarshal(spec, &customSpec); err != nil {
		return nil, fmt.Errorf("failed to parse custom spec: %w", err)
	}

	result := ir.NewIR(customSpec.Protocol)
	result.Title = customSpec.Title
	result.Version = customSpec.Version
	result.Description = customSpec.Description
	result.Metadata = customSpec.Metadata

	// Parse servers
	for _, server := range customSpec.Servers {
		result.Servers = append(result.Servers, ir.Server{
			URL:         server.URL,
			Description: server.Description,
			Variables:   server.Variables,
		})
	}

	// Parse endpoints
	for _, endpoint := range customSpec.Endpoints {
		resource := p.convertEndpointToResource(endpoint)
		result.Resources = append(result.Resources, resource)
	}

	// Parse resources (same as endpoints)
	for _, res := range customSpec.Resources {
		resource := p.convertEndpointToResource(EndpointSpec(res))
		result.Resources = append(result.Resources, resource)
	}

	// Parse types
	for _, typeSpec := range customSpec.Types {
		result.Types = append(result.Types, ir.TypeDef{
			Name:        typeSpec.Name,
			Description: typeSpec.Description,
			Schema:      convertMapToSchema(typeSpec.Schema),
		})
	}

	return result, nil
}

// convertEndpointToResource converts an EndpointSpec to IR Resource
func (p *Plugin) convertEndpointToResource(endpoint EndpointSpec) ir.Resource {
	resource := ir.Resource{
		Name:        endpoint.Name,
		Path:        endpoint.Path,
		Method:      endpoint.Method,
		Description: endpoint.Description,
		Tags:        endpoint.Tags,
		Parameters:  make([]ir.Parameter, 0),
		Examples:    make([]ir.Example, 0),
		Metadata:    endpoint.Metadata,
	}

	// Parse parameters
	for _, param := range endpoint.Parameters {
		p := ir.Parameter{
			Name:        param.Name,
			In:          param.In,
			Type:        param.Type,
			Description: param.Description,
			Required:    param.Required,
			Example:     param.Example,
		}
		if param.Schema != nil {
			p.Schema = convertMapToSchema(param.Schema)
		}
		resource.Parameters = append(resource.Parameters, p)
	}

	// Parse request
	if endpoint.Request != nil {
		resource.Request = convertMapToSchema(endpoint.Request)
	}

	// Parse response
	if endpoint.Response != nil {
		resource.Response = convertMapToSchema(endpoint.Response)
	}

	// Parse examples
	for _, example := range endpoint.Examples {
		resource.Examples = append(resource.Examples, ir.Example{
			Name:        example.Name,
			Description: example.Description,
			Value:       example.Value,
		})
	}

	return resource
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

	if format, ok := data["format"].(string); ok {
		schema.Format = format
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

	if items, ok := data["items"].(map[string]interface{}); ok {
		schema.Items = convertMapToSchema(items)
	}

	if example, ok := data["example"]; ok {
		schema.Example = example
	}

	if defaultVal, ok := data["default"]; ok {
		schema.Default = defaultVal
	}

	if enum, ok := data["enum"].([]interface{}); ok {
		schema.Enum = enum
	}

	// Store additional fields in metadata
	for key, value := range data {
		if key != "type" && key != "description" && key != "format" &&
			key != "properties" && key != "required" && key != "items" &&
			key != "example" && key != "default" && key != "enum" {
			schema.Metadata[key] = value
		}
	}

	return schema
}

// GetTemplate returns the custom template
func (p *Plugin) GetTemplate() string {
	return ""
}
