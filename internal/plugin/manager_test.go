package plugin

import (
	"testing"

	"github.com/sirrobot01/unifydoc/internal/ir"
)

// MockPlugin is a mock implementation of Plugin for testing
type MockPlugin struct {
	name string
}

func (m *MockPlugin) Name() string {
	return m.name
}

func (m *MockPlugin) Version() string {
	return "1.0.0"
}

func (m *MockPlugin) Parse(spec []byte) (*ir.IR, error) {
	return ir.NewIR(m.name), nil
}

func (m *MockPlugin) Validate(spec []byte) error {
	return nil
}

func (m *MockPlugin) GetTemplate() string {
	return ""
}

func TestManagerRegister(t *testing.T) {
	manager := NewManager()
	plugin := &MockPlugin{name: "test"}

	err := manager.Register(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	if manager.Count() != 1 {
		t.Errorf("Expected 1 plugin, got %d", manager.Count())
	}
}

func TestManagerGet(t *testing.T) {
	manager := NewManager()
	plugin := &MockPlugin{name: "test"}

	manager.Register(plugin)

	p, err := manager.Get("test")
	if err != nil {
		t.Fatalf("Failed to get plugin: %v", err)
	}

	if p.Name() != "test" {
		t.Errorf("Expected plugin name 'test', got '%s'", p.Name())
	}
}

func TestManagerGetNotFound(t *testing.T) {
	manager := NewManager()

	_, err := manager.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent plugin")
	}
}

func TestManagerList(t *testing.T) {
	manager := NewManager()
	plugin1 := &MockPlugin{name: "test1"}
	plugin2 := &MockPlugin{name: "test2"}

	manager.Register(plugin1)
	manager.Register(plugin2)

	list := manager.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 plugins in list, got %d", len(list))
	}
}

func TestManagerDuplicate(t *testing.T) {
	manager := NewManager()
	plugin := &MockPlugin{name: "test"}

	manager.Register(plugin)
	err := manager.Register(plugin)

	if err == nil {
		t.Error("Expected error when registering duplicate plugin")
	}
}
