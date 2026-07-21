package ir

import (
	"testing"
)

func TestNewIR(t *testing.T) {
	ir := NewIR("test-protocol")

	if ir.Protocol != "test-protocol" {
		t.Errorf("Expected protocol 'test-protocol', got '%s'", ir.Protocol)
	}

	if ir.Servers == nil {
		t.Error("Servers should not be nil")
	}

	if ir.Resources == nil {
		t.Error("Resources should not be nil")
	}

	if ir.Types == nil {
		t.Error("Types should not be nil")
	}

	if ir.Security == nil {
		t.Error("Security should not be nil")
	}

	if ir.Metadata == nil {
		t.Error("Metadata should not be nil")
	}
}

func TestIRFields(t *testing.T) {
	ir := NewIR("openapi")
	ir.Title = "Test API"
	ir.Version = "1.0.0"
	ir.Description = "Test description"

	if ir.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", ir.Title)
	}

	if ir.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", ir.Version)
	}

	if ir.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", ir.Description)
	}
}
