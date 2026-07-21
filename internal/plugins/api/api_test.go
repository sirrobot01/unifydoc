package api

import "testing"

const spec = `
api:
  title: Simple API
  version: 3.1.0
  baseUrl: https://api.example.com
  endpoints:
    - name: List widgets
      path: /widgets
      method: GET
      description: List all widgets
      tags: [Widgets]
      parameters:
        - name: limit
          in: query
          type: integer
          required: false
    - name: Create widget
      path: /widgets
      method: POST
      request:
        type: object
        properties:
          name:
            type: string
      response:
        type: object
        properties:
          id:
            type: string
`

func TestMetadata(t *testing.T) {
	p := NewPlugin()
	if p.Name() != "api" {
		t.Errorf("Name() = %q", p.Name())
	}
	if p.Version() == "" {
		t.Error("Version() empty")
	}
}

func TestValidate(t *testing.T) {
	p := NewPlugin()
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{"valid", spec, false},
		{"missing title", "api:\n  endpoints: []\n", true},
		{"missing path", "api:\n  title: X\n  endpoints:\n    - method: GET\n", true},
		{"missing method", "api:\n  title: X\n  endpoints:\n    - path: /x\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.Validate([]byte(tt.spec)); (err != nil) != tt.wantErr {
				t.Errorf("Validate() err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestParse(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte(spec))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.Title != "Simple API" || result.Version != "3.1.0" {
		t.Errorf("header wrong: %q %q", result.Title, result.Version)
	}
	if len(result.Servers) != 1 || result.Servers[0].URL != "https://api.example.com" {
		t.Errorf("baseUrl not mapped to server: %+v", result.Servers)
	}
	if len(result.Resources) != 2 {
		t.Fatalf("resources = %d, want 2", len(result.Resources))
	}

	list := result.Resources[0]
	if list.Method != "GET" || len(list.Parameters) != 1 || list.Parameters[0].In != "query" {
		t.Errorf("GET endpoint wrong: %+v", list)
	}

	create := result.Resources[1]
	if create.Request == nil || create.Request.Properties["name"] == nil {
		t.Errorf("request schema not converted: %+v", create.Request)
	}
	if create.Response == nil || create.Response.Properties["id"] == nil {
		t.Errorf("response schema not converted: %+v", create.Response)
	}
}
