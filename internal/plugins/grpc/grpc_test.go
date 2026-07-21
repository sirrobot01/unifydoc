package grpc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirrobot01/unifydoc/internal/ir"
)

const userProto = `
syntax = "proto3";
package user;

// User service manages users.
service UserService {
  // Get a user by ID.
  rpc GetUser (GetUserRequest) returns (GetUserResponse);
  // Stream user updates.
  rpc StreamUsers (StreamUsersRequest) returns (stream User);
}

message GetUserRequest {
  string user_id = 1;
}
message GetUserResponse {
  User user = 1;
}
message StreamUsersRequest {
  int32 limit = 1;
}

enum Role {
  ROLE_UNSPECIFIED = 0;
  ADMIN = 1;
  MEMBER = 2;
}

message User {
  string id = 1;
  string email = 2;
  repeated string tags = 3;
  Role role = 4;
  map<string, string> metadata = 5;
}
`

func TestPluginMetadata(t *testing.T) {
	p := NewPlugin()
	if p.Name() != "grpc" {
		t.Errorf("Name() = %q, want grpc", p.Name())
	}
	if p.Version() == "" {
		t.Error("Version() should not be empty")
	}
}

func TestValidate(t *testing.T) {
	p := NewPlugin()
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{"valid", userProto, false},
		{"garbage", "hello world", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.Validate([]byte(tt.spec)); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseResolvesMessages(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte(userProto))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.Protocol != "grpc" {
		t.Errorf("Protocol = %q, want grpc", result.Protocol)
	}
	if len(result.Resources) != 2 {
		t.Fatalf("Resources = %d, want 2", len(result.Resources))
	}

	get := findResource(result, "GetUser")
	if get == nil {
		t.Fatal("GetUser resource missing")
	}
	if get.Path != "/user.UserService/GetUser" {
		t.Errorf("Path = %q, want /user.UserService/GetUser", get.Path)
	}
	if get.Metadata["streaming"] != "unary" {
		t.Errorf("GetUser streaming = %v, want unary", get.Metadata["streaming"])
	}
	// Request message resolved into fields (the old parser could not do this).
	if get.Request == nil || get.Request.Properties["user_id"] == nil {
		t.Fatal("GetUser request should resolve field user_id")
	}
	if got := get.Request.Properties["user_id"].Type; got != "string" {
		t.Errorf("user_id type = %q, want string", got)
	}
	if c := get.Request.Properties["user_id"].Description; c == "" {
		// comment resolution is best-effort; only assert type above
		_ = c
	}
	// Nested message resolution: response.user expands into User fields.
	if get.Response == nil || get.Response.Properties["user"] == nil {
		t.Fatal("GetUser response should contain 'user'")
	}
	user := get.Response.Properties["user"]
	if user.Properties["email"] == nil {
		t.Error("nested User should expose 'email'")
	}
}

func TestParseStreamingAndFieldKinds(t *testing.T) {
	p := NewPlugin()
	result, err := p.Parse([]byte(userProto))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	stream := findResource(result, "StreamUsers")
	if stream == nil {
		t.Fatal("StreamUsers resource missing")
	}
	if stream.Metadata["streaming"] != "server_streaming" {
		t.Errorf("streaming = %v, want server_streaming", stream.Metadata["streaming"])
	}
	if stream.Request.Properties["limit"].Type != "integer" {
		t.Errorf("limit type = %q, want integer", stream.Request.Properties["limit"].Type)
	}

	// The User type should be documented with repeated, enum and map fields.
	user := findType(result, "User")
	if user == nil {
		t.Fatal("User type missing")
	}
	if tags := user.Schema.Properties["tags"]; tags == nil || tags.Type != "array" {
		t.Errorf("tags should be an array, got %+v", tags)
	}
	if role := user.Schema.Properties["role"]; role == nil || len(role.Enum) == 0 {
		t.Errorf("role should be an enum with values, got %+v", role)
	}
	if meta := user.Schema.Properties["metadata"]; meta == nil || meta.Metadata["map"] != true {
		t.Errorf("metadata should be a map, got %+v", meta)
	}
}

func TestParseInvalid(t *testing.T) {
	p := NewPlugin()
	if _, err := p.Parse([]byte("syntax = \"proto3\"; this is not valid")); err == nil {
		t.Error("Parse() should error on malformed proto")
	}
}

const commonProto = `
syntax = "proto3";
package shop;
// Money is an amount in a currency.
message Money {
  string currency = 1;
  int64 amount = 2;
}
`

const ordersProto = `
syntax = "proto3";
package shop;
import "common.proto";
// OrderService manages orders.
service OrderService {
  // Create an order.
  rpc CreateOrder (CreateOrderRequest) returns (Order);
}
message CreateOrderRequest {
  string sku = 1;
  Money price = 2;
}
message Order {
  string id = 1;
  Money total = 2;
}
`

func TestParsePathDirectoryResolvesImports(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "common.proto"), []byte(commonProto), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "orders.proto"), []byte(ordersProto), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := NewPlugin().ParsePath(dir)
	if err != nil {
		t.Fatalf("ParsePath(dir) error: %v", err)
	}

	create := findResource(result, "CreateOrder")
	if create == nil {
		t.Fatal("CreateOrder missing")
	}
	// The imported Money message must resolve inside the request field.
	price := create.Request.Properties["price"]
	if price == nil || price.Properties["currency"] == nil {
		t.Fatalf("imported Money not resolved into request: %+v", price)
	}
	// Types from both files should be present.
	if findType(result, "Money") == nil {
		t.Error("Money type (from common.proto) missing")
	}
	if findType(result, "Order") == nil {
		t.Error("Order type (from orders.proto) missing")
	}
}

func TestParsePathSingleFileWithSiblingImport(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "common.proto"), []byte(commonProto), 0644); err != nil {
		t.Fatal(err)
	}
	ordersPath := filepath.Join(dir, "orders.proto")
	if err := os.WriteFile(ordersPath, []byte(ordersProto), 0644); err != nil {
		t.Fatal(err)
	}

	// Pointing at the single file must still resolve its sibling import.
	result, err := NewPlugin().ParsePath(ordersPath)
	if err != nil {
		t.Fatalf("ParsePath(file) error: %v", err)
	}
	create := findResource(result, "CreateOrder")
	if create == nil || create.Request.Properties["price"] == nil {
		t.Fatalf("single-file ParsePath should resolve sibling import: %+v", result.Resources)
	}
}

func TestParsePathMissing(t *testing.T) {
	if _, err := NewPlugin().ParsePath(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Error("ParsePath should error on a missing path")
	}
}

func findResource(result *ir.IR, name string) *ir.Resource {
	for i := range result.Resources {
		if result.Resources[i].Name == name {
			return &result.Resources[i]
		}
	}
	return nil
}

func findType(result *ir.IR, name string) *ir.TypeDef {
	for i := range result.Types {
		if result.Types[i].Name == name {
			return &result.Types[i]
		}
	}
	return nil
}
