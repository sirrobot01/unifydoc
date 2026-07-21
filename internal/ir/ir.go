package ir

// IR represents the Intermediate Representation that normalizes all protocols
type IR struct {
	Protocol    string                 `json:"protocol"`
	Version     string                 `json:"version"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Servers     []Server               `json:"servers"`
	Resources   []Resource             `json:"resources"`
	Types       []TypeDef              `json:"types"`
	Security    []SecurityScheme       `json:"security"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Server represents a server endpoint
type Server struct {
	URL         string            `json:"url"`
	Description string            `json:"description"`
	Variables   map[string]string `json:"variables,omitempty"`
}

// Resource represents an endpoint, event, message, or any protocol resource
type Resource struct {
	Name        string                 `json:"name"`
	Path        string                 `json:"path,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Description string                 `json:"description"`
	Parameters  []Parameter            `json:"parameters,omitempty"`
	Request     *Schema                `json:"request,omitempty"`
	Response    *Schema                `json:"response,omitempty"`
	Responses   map[string]Response    `json:"responses,omitempty"`
	Examples    []Example              `json:"examples,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Security    []SecurityRequirement  `json:"security,omitempty"`
	Deprecated  bool                   `json:"deprecated,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Parameter represents a request parameter
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // query, header, path, cookie, body
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Type        string      `json:"type"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
	Default     interface{} `json:"default,omitempty"`
}

// Schema represents a data schema
type Schema struct {
	Type        string                `json:"type"`
	Format      string                `json:"format,omitempty"`
	Description string                `json:"description,omitempty"`
	Properties  map[string]*Schema    `json:"properties,omitempty"`
	Items       *Schema               `json:"items,omitempty"`
	Required    []string              `json:"required,omitempty"`
	Enum        []interface{}         `json:"enum,omitempty"`
	Example     interface{}           `json:"example,omitempty"`
	Default     interface{}           `json:"default,omitempty"`
	Ref         string                `json:"$ref,omitempty"`
	AllOf       []*Schema             `json:"allOf,omitempty"`
	AnyOf       []*Schema             `json:"anyOf,omitempty"`
	OneOf       []*Schema             `json:"oneOf,omitempty"`
	Not         *Schema               `json:"not,omitempty"`
	Nullable    bool                  `json:"nullable,omitempty"`
	ReadOnly    bool                  `json:"readOnly,omitempty"`
	WriteOnly   bool                  `json:"writeOnly,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Response represents an HTTP response or protocol response
type Response struct {
	Description string            `json:"description"`
	Headers     map[string]Header `json:"headers,omitempty"`
	Content     map[string]*Schema `json:"content,omitempty"`
	Schema      *Schema           `json:"schema,omitempty"`
}

// Header represents an HTTP header
type Header struct {
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Type        string      `json:"type"`
	Example     interface{} `json:"example,omitempty"`
}

// Example represents an example value
type Example struct {
	Name        string                 `json:"name"`
	Summary     string                 `json:"summary,omitempty"`
	Description string                 `json:"description,omitempty"`
	Value       interface{}            `json:"value"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TypeDef represents a type definition
type TypeDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      *Schema                `json:"schema"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityScheme represents an authentication/authorization scheme
type SecurityScheme struct {
	Type             string            `json:"type"` // http, apiKey, oauth2, openIdConnect
	Scheme           string            `json:"scheme,omitempty"`
	BearerFormat     string            `json:"bearerFormat,omitempty"`
	Description      string            `json:"description"`
	Name             string            `json:"name,omitempty"`
	In               string            `json:"in,omitempty"` // query, header, cookie
	Flows            *OAuthFlows       `json:"flows,omitempty"`
	OpenIDConnectURL string            `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows represents OAuth 2.0 flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow represents a single OAuth flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// SecurityRequirement represents a security requirement
type SecurityRequirement struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes,omitempty"`
}

// NewIR creates a new IR instance
func NewIR(protocol string) *IR {
	return &IR{
		Protocol:  protocol,
		Servers:   make([]Server, 0),
		Resources: make([]Resource, 0),
		Types:     make([]TypeDef, 0),
		Security:  make([]SecurityScheme, 0),
		Metadata:  make(map[string]interface{}),
	}
}
