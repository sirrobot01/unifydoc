package api

import (
	"fmt"

	"github.com/sirrobot01/unifydoc/internal/ir"
	"gopkg.in/yaml.v3"
)

// Plugin implements the API plugin for generic REST APIs
type Plugin struct{}

// NewPlugin creates a new API plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "api"
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Spec represents a simplified REST API specification
type Spec struct {
	API Config `yaml:"api" json:"api"`
}

// Config represents API configuration
type Config struct {
	Title       string     `yaml:"title" json:"title"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string     `yaml:"version,omitempty" json:"version,omitempty"`
	BaseURL     string     `yaml:"baseUrl,omitempty" json:"baseUrl,omitempty"`
	Endpoints   []Endpoint `yaml:"endpoints" json:"endpoints"`
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Name        string                 `yaml:"name" json:"name"`
	Path        string                 `yaml:"path" json:"path"`
	Method      string                 `yaml:"method" json:"method"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Parameters  []Parameter            `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Request     map[string]interface{} `yaml:"request,omitempty" json:"request,omitempty"`
	Response    map[string]interface{} `yaml:"response,omitempty" json:"response,omitempty"`
	Examples    []Example              `yaml:"examples,omitempty" json:"examples,omitempty"`
	Tags        []string               `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// Parameter represents a parameter
type Parameter struct {
	Name        string      `yaml:"name" json:"name"`
	In          string      `yaml:"in" json:"in"` // query, header, path
	Type        string      `yaml:"type" json:"type"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool        `yaml:"required,omitempty" json:"required,omitempty"`
	Example     interface{} `yaml:"example,omitempty" json:"example,omitempty"`
}

// Example represents an example
type Example struct {
	Name  string      `yaml:"name" json:"name"`
	Value interface{} `yaml:"value" json:"value"`
}

// Validate validates an API specification
func (p *Plugin) Validate(spec []byte) error {
	var apiSpec Spec
	if err := yaml.Unmarshal(spec, &apiSpec); err != nil {
		return fmt.Errorf("invalid API YAML: %w", err)
	}

	if apiSpec.API.Title == "" {
		return fmt.Errorf("title is required")
	}

	for i, endpoint := range apiSpec.API.Endpoints {
		if endpoint.Path == "" {
			return fmt.Errorf("endpoint %d: path is required", i)
		}
		if endpoint.Method == "" {
			return fmt.Errorf("endpoint %s: method is required", endpoint.Path)
		}
	}

	return nil
}

// Parse parses an API specification and converts it to IR
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	var apiSpec Spec
	if err := yaml.Unmarshal(spec, &apiSpec); err != nil {
		return nil, fmt.Errorf("failed to parse API spec: %w", err)
	}

	result := ir.NewIR("api")
	result.Title = apiSpec.API.Title
	result.Description = apiSpec.API.Description
	result.Version = apiSpec.API.Version

	if result.Version == "" {
		result.Version = "1.0.0"
	}

	// Add server
	if apiSpec.API.BaseURL != "" {
		result.Servers = append(result.Servers, ir.Server{
			URL:         apiSpec.API.BaseURL,
			Description: "API Base URL",
		})
	}

	// Parse endpoints
	for _, endpoint := range apiSpec.API.Endpoints {
		resource := ir.Resource{
			Name:        endpoint.Name,
			Path:        endpoint.Path,
			Method:      endpoint.Method,
			Description: endpoint.Description,
			Tags:        endpoint.Tags,
			Parameters:  make([]ir.Parameter, 0),
			Examples:    make([]ir.Example, 0),
		}

		if resource.Name == "" {
			resource.Name = fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path)
		}

		// Parse parameters
		for _, param := range endpoint.Parameters {
			resource.Parameters = append(resource.Parameters, ir.Parameter{
				Name:        param.Name,
				In:          param.In,
				Type:        param.Type,
				Description: param.Description,
				Required:    param.Required,
				Example:     param.Example,
			})
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
				Name:  example.Name,
				Value: example.Value,
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

	if items, ok := data["items"].(map[string]interface{}); ok {
		schema.Items = convertMapToSchema(items)
	}

	return schema
}

// GetTemplate returns the custom template
func (p *Plugin) GetTemplate() string {
	return ""
}
