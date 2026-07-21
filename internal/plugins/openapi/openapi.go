package openapi

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sirrobot01/unifydoc/internal/ir"
	"gopkg.in/yaml.v3"
)

// Plugin implements the OpenAPI plugin
type Plugin struct{}

// NewPlugin creates a new OpenAPI plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return "openapi"
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Validate validates an OpenAPI specification
func (p *Plugin) Validate(spec []byte) error {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Try to load as JSON first
	doc, err := loader.LoadFromData(spec)
	if err != nil {
		return fmt.Errorf("invalid OpenAPI specification: %w", err)
	}

	// Validate the document
	if err := doc.Validate(loader.Context); err != nil {
		return fmt.Errorf("OpenAPI validation failed: %w", err)
	}

	return nil
}

// Parse parses an OpenAPI specification and converts it to IR
func (p *Plugin) Parse(spec []byte) (*ir.IR, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromData(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	result := ir.NewIR("openapi")
	result.Title = doc.Info.Title
	result.Description = doc.Info.Description
	result.Version = doc.Info.Version

	// Parse servers
	for _, server := range doc.Servers {
		s := ir.Server{
			URL:         server.URL,
			Description: server.Description,
			Variables:   make(map[string]string),
		}
		for name, variable := range server.Variables {
			s.Variables[name] = variable.Default
		}
		result.Servers = append(result.Servers, s)
	}

	// Parse paths (resources)
	for path, pathItem := range doc.Paths.Map() {
		if pathItem == nil {
			continue
		}

		operations := map[string]*openapi3.Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"DELETE":  pathItem.Delete,
			"PATCH":   pathItem.Patch,
			"HEAD":    pathItem.Head,
			"OPTIONS": pathItem.Options,
			"TRACE":   pathItem.Trace,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			resource := ir.Resource{
				Name:        operation.Summary,
				Path:        path,
				Method:      method,
				Description: operation.Description,
				Parameters:  make([]ir.Parameter, 0),
				Tags:        operation.Tags,
				Deprecated:  operation.Deprecated,
				Responses:   make(map[string]ir.Response),
				Examples:    make([]ir.Example, 0),
				Metadata:    make(map[string]interface{}),
			}

			if resource.Name == "" {
				resource.Name = fmt.Sprintf("%s %s", method, path)
			}

			// Parse parameters
			for _, param := range operation.Parameters {
				if param.Value == nil {
					continue
				}

				p := ir.Parameter{
					Name:        param.Value.Name,
					In:          param.Value.In,
					Description: param.Value.Description,
					Required:    param.Value.Required,
				}

				if param.Value.Schema != nil && param.Value.Schema.Value != nil {
					p.Type = param.Value.Schema.Value.Type.Slice()[0]
					p.Schema = convertSchema(param.Value.Schema.Value)
					if param.Value.Schema.Value.Example != nil {
						p.Example = param.Value.Schema.Value.Example
					}
					if param.Value.Schema.Value.Default != nil {
						p.Default = param.Value.Schema.Value.Default
					}
				}

				resource.Parameters = append(resource.Parameters, p)
			}

			// Parse request body
			if operation.RequestBody != nil && operation.RequestBody.Value != nil {
				for contentType, mediaType := range operation.RequestBody.Value.Content {
					if mediaType.Schema != nil && mediaType.Schema.Value != nil {
						resource.Request = convertSchema(mediaType.Schema.Value)
						resource.Metadata["requestContentType"] = contentType

						// Add examples
						if mediaType.Example != nil {
							resource.Examples = append(resource.Examples, ir.Example{
								Name:  "Request Example",
								Value: mediaType.Example,
							})
						}
						if len(mediaType.Examples) > 0 {
							for exName, ex := range mediaType.Examples {
								if ex.Value != nil && ex.Value.Value != nil {
									resource.Examples = append(resource.Examples, ir.Example{
										Name:    exName,
										Summary: ex.Value.Summary,
										Value:   ex.Value.Value,
									})
								}
							}
						}
						break
					}
				}
			}

			// Parse responses
			for status, response := range operation.Responses.Map() {
				if response.Value == nil {
					continue
				}

				description := ""
				if response.Value.Description != nil {
					description = *response.Value.Description
				}
				resp := ir.Response{
					Description: description,
					Headers:     make(map[string]ir.Header),
					Content:     make(map[string]*ir.Schema),
				}

				// Parse headers
				for headerName, header := range response.Value.Headers {
					if header.Value == nil {
						continue
					}
					h := ir.Header{
						Description: header.Value.Description,
						Required:    header.Value.Required,
					}
					if header.Value.Schema != nil && header.Value.Schema.Value != nil {
						h.Type = header.Value.Schema.Value.Type.Slice()[0]
					}
					resp.Headers[headerName] = h
				}

				// Parse content - use the first available content type's schema
				for contentType, mediaType := range response.Value.Content {
					if mediaType.Schema != nil && mediaType.Schema.Value != nil {
						resp.Schema = convertSchema(mediaType.Schema.Value)
						resp.Content[contentType] = resp.Schema
						break // Use first content type only
					}
				}

				resource.Responses[status] = resp
			}

			// Parse security
			if operation.Security != nil {
				for _, secReq := range *operation.Security {
					for name, scopes := range secReq {
						resource.Security = append(resource.Security, ir.SecurityRequirement{
							Name:   name,
							Scopes: scopes,
						})
					}
				}
			}

			result.Resources = append(result.Resources, resource)
		}
	}

	// Parse components/schemas as types
	if doc.Components != nil && doc.Components.Schemas != nil {
		for name, schemaRef := range doc.Components.Schemas {
			if schemaRef.Value == nil {
				continue
			}

			typeDef := ir.TypeDef{
				Name:        name,
				Description: schemaRef.Value.Description,
				Schema:      convertSchema(schemaRef.Value),
			}
			result.Types = append(result.Types, typeDef)
		}
	}

	// Parse security schemes
	if doc.Components != nil && doc.Components.SecuritySchemes != nil {
		for name, secSchemeRef := range doc.Components.SecuritySchemes {
			if secSchemeRef.Value == nil {
				continue
			}

			scheme := ir.SecurityScheme{
				Type:         secSchemeRef.Value.Type,
				Description:  secSchemeRef.Value.Description,
				Name:         name,
				Scheme:       secSchemeRef.Value.Scheme,
				BearerFormat: secSchemeRef.Value.BearerFormat,
			}

			if secSchemeRef.Value.In != "" {
				scheme.In = secSchemeRef.Value.In
			}

			if secSchemeRef.Value.Flows != nil {
				scheme.Flows = convertOAuthFlows(secSchemeRef.Value.Flows)
			}

			result.Security = append(result.Security, scheme)
		}
	}

	return result, nil
}

// convertSchema converts OpenAPI schema to IR schema
func convertSchema(schema *openapi3.Schema) *ir.Schema {
	if schema == nil {
		return nil
	}

	result := &ir.Schema{
		Type:        "",
		Format:      schema.Format,
		Description: schema.Description,
		Properties:  make(map[string]*ir.Schema),
		Required:    schema.Required,
		Nullable:    schema.Nullable,
		ReadOnly:    schema.ReadOnly,
		WriteOnly:   schema.WriteOnly,
		Deprecated:  schema.Deprecated,
		Example:     schema.Example,
		Default:     schema.Default,
		Metadata:    make(map[string]interface{}),
	}

	// Handle type (can be multiple in OpenAPI 3.1)
	if len(schema.Type.Slice()) > 0 {
		result.Type = schema.Type.Slice()[0]
	}

	// Handle enum
	if len(schema.Enum) > 0 {
		result.Enum = schema.Enum
	}

	// Handle properties
	for name, propRef := range schema.Properties {
		if propRef.Value != nil {
			result.Properties[name] = convertSchema(propRef.Value)
		}
	}

	// Handle items (for arrays)
	if schema.Items != nil && schema.Items.Value != nil {
		result.Items = convertSchema(schema.Items.Value)
	}

	// Handle composition
	if len(schema.AllOf) > 0 {
		result.AllOf = make([]*ir.Schema, 0)
		for _, s := range schema.AllOf {
			if s.Value != nil {
				result.AllOf = append(result.AllOf, convertSchema(s.Value))
			}
		}
	}

	if len(schema.AnyOf) > 0 {
		result.AnyOf = make([]*ir.Schema, 0)
		for _, s := range schema.AnyOf {
			if s.Value != nil {
				result.AnyOf = append(result.AnyOf, convertSchema(s.Value))
			}
		}
	}

	if len(schema.OneOf) > 0 {
		result.OneOf = make([]*ir.Schema, 0)
		for _, s := range schema.OneOf {
			if s.Value != nil {
				result.OneOf = append(result.OneOf, convertSchema(s.Value))
			}
		}
	}

	if schema.Not != nil && schema.Not.Value != nil {
		result.Not = convertSchema(schema.Not.Value)
	}

	// Store additional metadata
	if schema.Min != nil {
		result.Metadata["minimum"] = *schema.Min
	}
	if schema.Max != nil {
		result.Metadata["maximum"] = *schema.Max
	}
	if schema.MinLength > 0 {
		result.Metadata["minLength"] = schema.MinLength
	}
	if schema.MaxLength != nil {
		result.Metadata["maxLength"] = *schema.MaxLength
	}
	if schema.Pattern != "" {
		result.Metadata["pattern"] = schema.Pattern
	}

	return result
}

// convertOAuthFlows converts OpenAPI OAuth flows to IR
func convertOAuthFlows(flows *openapi3.OAuthFlows) *ir.OAuthFlows {
	if flows == nil {
		return nil
	}

	result := &ir.OAuthFlows{}

	if flows.Implicit != nil {
		result.Implicit = &ir.OAuthFlow{
			AuthorizationURL: flows.Implicit.AuthorizationURL,
			RefreshURL:       flows.Implicit.RefreshURL,
			Scopes:           flows.Implicit.Scopes,
		}
	}

	if flows.Password != nil {
		result.Password = &ir.OAuthFlow{
			TokenURL:   flows.Password.TokenURL,
			RefreshURL: flows.Password.RefreshURL,
			Scopes:     flows.Password.Scopes,
		}
	}

	if flows.ClientCredentials != nil {
		result.ClientCredentials = &ir.OAuthFlow{
			TokenURL:   flows.ClientCredentials.TokenURL,
			RefreshURL: flows.ClientCredentials.RefreshURL,
			Scopes:     flows.ClientCredentials.Scopes,
		}
	}

	if flows.AuthorizationCode != nil {
		result.AuthorizationCode = &ir.OAuthFlow{
			AuthorizationURL: flows.AuthorizationCode.AuthorizationURL,
			TokenURL:         flows.AuthorizationCode.TokenURL,
			RefreshURL:       flows.AuthorizationCode.RefreshURL,
			Scopes:           flows.AuthorizationCode.Scopes,
		}
	}

	return result
}

// GetTemplate returns the custom template (empty for default)
func (p *Plugin) GetTemplate() string {
	return ""
}

// ParseYAMLOrJSON attempts to parse spec as YAML or JSON
func ParseYAMLOrJSON(data []byte, v interface{}) error {
	// Try JSON first
	if err := json.Unmarshal(data, v); err == nil {
		return nil
	}

	// Try YAML
	if err := yaml.Unmarshal(data, v); err == nil {
		return nil
	}

	return fmt.Errorf("failed to parse as JSON or YAML")
}
