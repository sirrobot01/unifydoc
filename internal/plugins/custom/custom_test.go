package custom

import "testing"

const spec = `
protocol: graphql
title: My GraphQL API
version: 1.0.0
description: A custom protocol
servers:
  - url: https://api.example.com/graphql
    description: Production
endpoints:
  - name: getUser
    method: QUERY
    description: Fetch a user
    request:
      type: object
      properties:
        id:
          type: string
    tags: [Users]
types:
  - name: User
    description: A user
    schema:
      type: object
      properties:
        id:
          type: string
`

func TestMetadata(t *testing.T) {
	p := NewPlugin()
	if p.Name() != "custom" {
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
		{"missing protocol", "title: X\n", true},
		{"missing title", "protocol: x\n", true},
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
	// The custom protocol name drives the IR protocol.
	if result.Protocol != "graphql" {
		t.Errorf("Protocol = %q, want graphql", result.Protocol)
	}
	if result.Title != "My GraphQL API" {
		t.Errorf("Title = %q", result.Title)
	}
	if len(result.Servers) != 1 || result.Servers[0].URL != "https://api.example.com/graphql" {
		t.Errorf("server wrong: %+v", result.Servers)
	}
	if len(result.Resources) != 1 {
		t.Fatalf("resources = %d, want 1", len(result.Resources))
	}
	if result.Resources[0].Method != "QUERY" {
		t.Errorf("method = %q, want QUERY", result.Resources[0].Method)
	}
	if len(result.Types) != 1 || result.Types[0].Name != "User" {
		t.Errorf("types wrong: %+v", result.Types)
	}
}
