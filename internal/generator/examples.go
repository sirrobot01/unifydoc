package generator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirrobot01/unifydoc/internal/ir"
)

// ExampleGenerator generates code examples in multiple languages
type ExampleGenerator struct {
	languages []string
}

// NewExampleGenerator creates a new example generator
func NewExampleGenerator(languages []string) *ExampleGenerator {
	if len(languages) == 0 {
		languages = []string{"curl", "javascript", "python", "go"}
	}
	return &ExampleGenerator{
		languages: languages,
	}
}

// Generate generates code examples for a resource
func (g *ExampleGenerator) Generate(resource *ir.Resource, baseURL string) map[string]string {
	examples := make(map[string]string)

	for _, lang := range g.languages {
		switch lang {
		case "curl":
			examples["curl"] = g.generateCurl(resource, baseURL)
		case "javascript":
			examples["javascript"] = g.generateJavaScript(resource, baseURL)
		case "python":
			examples["python"] = g.generatePython(resource, baseURL)
		case "go":
			examples["go"] = g.generateGo(resource, baseURL)
		}
	}

	return examples
}

// generateCurl generates a curl example
func (g *ExampleGenerator) generateCurl(resource *ir.Resource, baseURL string) string {
	var sb strings.Builder

	method := strings.ToUpper(resource.Method)
	if method == "" || method == "RPC" || method == "EVENT" || method == "TOOL" || method == "RESOURCE" || method == "PROMPT" {
		method = "GET"
	}

	url := baseURL + resource.Path
	sb.WriteString(fmt.Sprintf("curl -X %s \\\n", method))
	sb.WriteString(fmt.Sprintf("  '%s'", url))

	// Add headers
	hasContentType := false
	for _, param := range resource.Parameters {
		if param.In == "header" {
			sb.WriteString(fmt.Sprintf(" \\\n  -H '%s: %v'", param.Name, param.Example))
			if param.Name == "Content-Type" {
				hasContentType = true
			}
		}
	}

	// Add request body
	if resource.Request != nil && (method == "POST" || method == "PUT" || method == "PATCH") {
		if !hasContentType {
			sb.WriteString(" \\\n  -H 'Content-Type: application/json'")
		}

		var exampleData interface{}
		if len(resource.Examples) > 0 {
			exampleData = resource.Examples[0].Value
		} else {
			exampleData = g.generateExampleData(resource.Request)
		}

		jsonData, _ := json.MarshalIndent(exampleData, "  ", "  ")
		sb.WriteString(fmt.Sprintf(" \\\n  -d '%s'", string(jsonData)))
	}

	return sb.String()
}

// generateJavaScript generates a JavaScript fetch example
func (g *ExampleGenerator) generateJavaScript(resource *ir.Resource, baseURL string) string {
	var sb strings.Builder

	method := strings.ToUpper(resource.Method)
	if method == "" || method == "RPC" || method == "EVENT" || method == "TOOL" || method == "RESOURCE" || method == "PROMPT" {
		method = "GET"
	}

	url := baseURL + resource.Path

	sb.WriteString(fmt.Sprintf("fetch('%s', {\n", url))
	sb.WriteString(fmt.Sprintf("  method: '%s',\n", method))

	// Add headers
	headers := make(map[string]string)
	for _, param := range resource.Parameters {
		if param.In == "header" {
			headers[param.Name] = fmt.Sprintf("%v", param.Example)
		}
	}

	if resource.Request != nil && (method == "POST" || method == "PUT" || method == "PATCH") {
		headers["Content-Type"] = "application/json"

		var exampleData interface{}
		if len(resource.Examples) > 0 {
			exampleData = resource.Examples[0].Value
		} else {
			exampleData = g.generateExampleData(resource.Request)
		}

		if len(headers) > 0 {
			sb.WriteString("  headers: {\n")
			for k, v := range headers {
				sb.WriteString(fmt.Sprintf("    '%s': '%s',\n", k, v))
			}
			sb.WriteString("  },\n")
		}

		jsonData, _ := json.MarshalIndent(exampleData, "  ", "  ")
		sb.WriteString(fmt.Sprintf("  body: JSON.stringify(%s)\n", string(jsonData)))
	} else if len(headers) > 0 {
		sb.WriteString("  headers: {\n")
		for k, v := range headers {
			sb.WriteString(fmt.Sprintf("    '%s': '%s',\n", k, v))
		}
		sb.WriteString("  }\n")
	}

	sb.WriteString("})\n")
	sb.WriteString("  .then(response => response.json())\n")
	sb.WriteString("  .then(data => console.log(data))\n")
	sb.WriteString("  .catch(error => console.error('Error:', error));")

	return sb.String()
}

// generatePython generates a Python requests example
func (g *ExampleGenerator) generatePython(resource *ir.Resource, baseURL string) string {
	var sb strings.Builder

	method := strings.ToLower(resource.Method)
	if method == "" || method == "rpc" || method == "event" || method == "tool" || method == "resource" || method == "prompt" {
		method = "get"
	}

	url := baseURL + resource.Path

	sb.WriteString("import requests\n\n")

	// Prepare headers
	hasHeaders := false
	for _, param := range resource.Parameters {
		if param.In == "header" {
			hasHeaders = true
			break
		}
	}

	if hasHeaders || resource.Request != nil {
		sb.WriteString("headers = {\n")
		for _, param := range resource.Parameters {
			if param.In == "header" {
				sb.WriteString(fmt.Sprintf("    '%s': '%v',\n", param.Name, param.Example))
			}
		}
		if resource.Request != nil && (method == "post" || method == "put" || method == "patch") {
			sb.WriteString("    'Content-Type': 'application/json',\n")
		}
		sb.WriteString("}\n\n")
	}

	// Prepare data
	if resource.Request != nil && (method == "post" || method == "put" || method == "patch") {
		var exampleData interface{}
		if len(resource.Examples) > 0 {
			exampleData = resource.Examples[0].Value
		} else {
			exampleData = g.generateExampleData(resource.Request)
		}

		jsonData, _ := json.MarshalIndent(exampleData, "", "    ")
		sb.WriteString(fmt.Sprintf("data = %s\n\n", string(jsonData)))
	}

	// Make request
	sb.WriteString(fmt.Sprintf("response = requests.%s(\n", method))
	sb.WriteString(fmt.Sprintf("    '%s'", url))

	if hasHeaders || resource.Request != nil {
		sb.WriteString(",\n    headers=headers")
	}

	if resource.Request != nil && (method == "post" || method == "put" || method == "patch") {
		sb.WriteString(",\n    json=data")
	}

	sb.WriteString("\n)\n\n")
	sb.WriteString("print(response.json())")

	return sb.String()
}

// generateGo generates a Go example
func (g *ExampleGenerator) generateGo(resource *ir.Resource, baseURL string) string {
	var sb strings.Builder

	method := strings.ToUpper(resource.Method)
	if method == "" || method == "RPC" || method == "EVENT" || method == "TOOL" || method == "RESOURCE" || method == "PROMPT" {
		method = "GET"
	}

	url := baseURL + resource.Path

	sb.WriteString("package main\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("    \"bytes\"\n")
	sb.WriteString("    \"encoding/json\"\n")
	sb.WriteString("    \"fmt\"\n")
	sb.WriteString("    \"io\"\n")
	sb.WriteString("    \"net/http\"\n")
	sb.WriteString(")\n\n")

	sb.WriteString("func main() {\n")

	// Prepare request body
	if resource.Request != nil && (method == "POST" || method == "PUT" || method == "PATCH") {
		var exampleData interface{}
		if len(resource.Examples) > 0 {
			exampleData = resource.Examples[0].Value
		} else {
			exampleData = g.generateExampleData(resource.Request)
		}

		jsonData, _ := json.MarshalIndent(exampleData, "    ", "    ")
		sb.WriteString(fmt.Sprintf("    data := []byte(`%s`)\n", string(jsonData)))
		sb.WriteString(fmt.Sprintf("    req, err := http.NewRequest(\"%s\", \"%s\", bytes.NewBuffer(data))\n", method, url))
	} else {
		sb.WriteString(fmt.Sprintf("    req, err := http.NewRequest(\"%s\", \"%s\", nil)\n", method, url))
	}

	sb.WriteString("    if err != nil {\n")
	sb.WriteString("        panic(err)\n")
	sb.WriteString("    }\n\n")

	// Add headers
	for _, param := range resource.Parameters {
		if param.In == "header" {
			sb.WriteString(fmt.Sprintf("    req.Header.Set(\"%s\", \"%v\")\n", param.Name, param.Example))
		}
	}

	if resource.Request != nil && (method == "POST" || method == "PUT" || method == "PATCH") {
		sb.WriteString("    req.Header.Set(\"Content-Type\", \"application/json\")\n\n")
	}

	sb.WriteString("    client := &http.Client{}\n")
	sb.WriteString("    resp, err := client.Do(req)\n")
	sb.WriteString("    if err != nil {\n")
	sb.WriteString("        panic(err)\n")
	sb.WriteString("    }\n")
	sb.WriteString("    defer resp.Body.Close()\n\n")

	sb.WriteString("    body, _ := io.ReadAll(resp.Body)\n")
	sb.WriteString("    fmt.Println(string(body))\n")
	sb.WriteString("}")

	return sb.String()
}

// generateExampleData generates example data from a schema
func (g *ExampleGenerator) generateExampleData(schema *ir.Schema) interface{} {
	if schema == nil {
		return nil
	}

	// If there's an example, use it
	if schema.Example != nil {
		return schema.Example
	}

	// Generate based on type
	switch schema.Type {
	case "object":
		obj := make(map[string]interface{})
		for name, prop := range schema.Properties {
			obj[name] = g.generateExampleData(prop)
		}
		return obj

	case "array":
		if schema.Items != nil {
			return []interface{}{g.generateExampleData(schema.Items)}
		}
		return []interface{}{}

	case "string":
		if schema.Format == "email" {
			return "user@example.com"
		}
		if schema.Format == "date-time" {
			return "2025-01-01T00:00:00Z"
		}
		if len(schema.Enum) > 0 {
			return schema.Enum[0]
		}
		return "string"

	case "integer":
		return 0

	case "number":
		return 0.0

	case "boolean":
		return true

	default:
		return nil
	}
}
