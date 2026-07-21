package ir

import (
	"testing"
)

func TestMergerAdd(t *testing.T) {
	merger := NewMerger()

	ir1 := NewIR("openapi")
	ir1.Title = "API 1"

	merger.Add(ir1)

	if len(merger.irs) != 1 {
		t.Errorf("Expected 1 IR, got %d", len(merger.irs))
	}
}

func TestMergerMerge(t *testing.T) {
	merger := NewMerger()

	ir1 := NewIR("openapi")
	ir1.Title = "OpenAPI"
	ir1.Resources = append(ir1.Resources, Resource{Name: "GET /users", Method: "GET"})

	ir2 := NewIR("grpc")
	ir2.Title = "gRPC"
	ir2.Resources = append(ir2.Resources, Resource{Name: "GetUser", Method: "RPC"})

	merger.Add(ir1)
	merger.Add(ir2)

	merged, err := merger.Merge()
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if len(merged.Protocols) != 2 {
		t.Errorf("Expected 2 protocols, got %d", len(merged.Protocols))
	}

	if len(merged.AllResources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(merged.AllResources))
	}

	if len(merged.Navigation) != 2 {
		t.Errorf("Expected 2 navigation items, got %d", len(merged.Navigation))
	}
}

func TestMergerEmpty(t *testing.T) {
	merger := NewMerger()

	_, err := merger.Merge()
	if err == nil {
		t.Error("Expected error when merging empty IRs")
	}
}
