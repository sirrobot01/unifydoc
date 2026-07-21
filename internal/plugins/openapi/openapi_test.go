package openapi

import (
	"testing"
)

const validSpec = `
openapi: 3.0.0
info:
  title: Pet Store API
  description: A sample API
  version: 2.1.0
servers:
  - url: https://api.petstore.com/v1
    description: Production
paths:
  /pets:
    get:
      summary: List all pets
      description: Returns all pets
      tags: [Pets]
      parameters:
        - name: limit
          in: query
          description: Max pets
          required: false
          schema:
            type: integer
            default: 20
      responses:
        '200':
          description: A list of pets
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
    post:
      summary: Create a pet
      security:
        - apiKey: []
      responses:
        '201':
          description: Created
components:
  schemas:
    Pet:
      type: object
      required: [id, name]
      properties:
        id:
          type: integer
        name:
          type: string
  securitySchemes:
    apiKey:
      type: apiKey
      in: header
      name: X-API-Key
`

func TestPluginMetadata(t *testing.T) {
	p := NewPlugin()
	if p.Name() != "openapi" {
		t.Errorf("Name() = %q, want %q", p.Name(), "openapi")
	}
	if p.Version() == "" {
		t.Error("Version() should not be empty")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{"valid spec", validSpec, false},
		{"not openapi", "just some text", true},
		{"empty", "", true},
	}

	p := NewPlugin()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Validate([]byte(tt.spec))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParse(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte(validSpec))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	if result.Protocol != "openapi" {
		t.Errorf("Protocol = %q, want %q", result.Protocol, "openapi")
	}
	if result.Title != "Pet Store API" {
		t.Errorf("Title = %q, want %q", result.Title, "Pet Store API")
	}
	if result.Version != "2.1.0" {
		t.Errorf("Version = %q, want %q", result.Version, "2.1.0")
	}
	if len(result.Servers) != 1 || result.Servers[0].URL != "https://api.petstore.com/v1" {
		t.Errorf("Servers = %+v, want one production server", result.Servers)
	}

	// Two operations across one path (GET, POST) => two resources.
	if len(result.Resources) != 2 {
		t.Fatalf("Resources = %d, want 2", len(result.Resources))
	}

	// Find the GET /pets resource and check its parameter.
	foundGet, hasLimit := false, false
	for _, r := range result.Resources {
		if r.Method == "GET" && r.Path == "/pets" {
			foundGet = true
			for _, param := range r.Parameters {
				if param.Name == "limit" && param.In == "query" {
					hasLimit = true
				}
			}
		}
	}
	if !foundGet {
		t.Fatal("did not find GET /pets resource")
	}
	if !hasLimit {
		t.Error("GET /pets should have a 'limit' query parameter")
	}

	// Components schema => one type.
	if len(result.Types) != 1 || result.Types[0].Name != "Pet" {
		t.Errorf("Types = %+v, want one 'Pet' type", result.Types)
	}
	if len(result.Types) == 1 {
		pet := result.Types[0].Schema
		if pet == nil || pet.Type != "object" {
			t.Errorf("Pet schema type = %v, want object", pet)
		}
		if len(pet.Required) != 2 {
			t.Errorf("Pet required = %v, want [id name]", pet.Required)
		}
	}

	// Security scheme parsed.
	if len(result.Security) != 1 || result.Security[0].Type != "apiKey" {
		t.Errorf("Security = %+v, want one apiKey scheme", result.Security)
	}
}

func TestParseInvalid(t *testing.T) {
	p := NewPlugin()
	if _, err := p.Parse([]byte("::: not valid :::")); err == nil {
		t.Error("Parse() should error on malformed input")
	}
}
