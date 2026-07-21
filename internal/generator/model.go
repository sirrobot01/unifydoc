package generator

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/sirrobot01/unifydoc/internal/ir"
)

// docModel is the JSON payload embedded into the generated page and rendered
// client-side. It mirrors the interactive design's data shape.
type docModel struct {
	Project   docProject    `json:"project"`
	Protocols []docProtocol `json:"protocols"`
}

type docProject struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

type docProtocol struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Servers     []ir.Server   `json:"servers"`
	Resources   []docResource `json:"resources"`
}

type docResource struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Method          string       `json:"method"`
	Path            string       `json:"path"`
	Description     string       `json:"description"`
	Tags            []string     `json:"tags"`
	Deprecated      bool         `json:"deprecated"`
	Params          []docParam   `json:"params"`
	Body            []docParam   `json:"body"`
	Responses       []docResponse `json:"responses"`
	Security        []docAuth    `json:"security"`
	Code            map[string]string `json:"code"`
	ResponseStatus  string       `json:"responseStatus"`
	ResponseExample string       `json:"responseExample"`
}

type docParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	In          string `json:"in,omitempty"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type docResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

type docAuth struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes,omitempty"`
}

// buildDocModel converts the merged IR into the client render model.
func (g *Generator) buildDocModel(merged *ir.MergedIR) *docModel {
	model := &docModel{
		Project: docProject{
			Name:        g.config.Project.Name,
			Version:     g.config.Project.Version,
			Description: g.config.Project.Description,
		},
		Protocols: make([]docProtocol, 0, len(merged.Protocols)),
	}

	for _, proto := range merged.Protocols {
		dp := docProtocol{
			ID:          proto.Protocol,
			Title:       proto.Title,
			Description: proto.Description,
			Servers:     proto.Servers,
			Resources:   make([]docResource, 0, len(proto.Resources)),
		}
		if dp.Title == "" {
			dp.Title = proto.Protocol
		}

		for i := range proto.Resources {
			res := &proto.Resources[i]
			dr := docResource{
				ID:          resourceID(proto.Protocol, i),
				Name:        res.Name,
				Method:      res.Method,
				Path:        res.Path,
				Description: res.Description,
				Tags:        res.Tags,
				Deprecated:  res.Deprecated,
				Params:      buildParams(res.Parameters),
				Body:        buildBody(res.Request),
				Responses:   buildResponses(res.Responses),
				Security:    buildAuth(res.Security),
				Code:        codeExamples(res),
			}
			dr.ResponseStatus, dr.ResponseExample = g.pickResponseExample(res)
			dp.Resources = append(dp.Resources, dr)
		}

		model.Protocols = append(model.Protocols, dp)
	}

	return model
}

// resourceID builds a stable, URL-fragment-safe id for a resource.
func resourceID(protocol string, index int) string {
	return protocol + "-" + itoa(index)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func buildParams(params []ir.Parameter) []docParam {
	out := make([]docParam, 0, len(params))
	for _, p := range params {
		out = append(out, docParam{
			Name:        p.Name,
			Type:        paramType(p),
			In:          p.In,
			Required:    p.Required,
			Description: p.Description,
		})
	}
	return out
}

func paramType(p ir.Parameter) string {
	if p.Type != "" {
		return p.Type
	}
	if p.Schema != nil {
		return schemaTypeLabel(p.Schema)
	}
	return "string"
}

// buildBody flattens a request body object schema into parameter-style rows.
func buildBody(schema *ir.Schema) []docParam {
	if schema == nil || len(schema.Properties) == 0 {
		return nil
	}
	required := make(map[string]bool, len(schema.Required))
	for _, r := range schema.Required {
		required[r] = true
	}
	out := make([]docParam, 0, len(schema.Properties))
	for name, prop := range schema.Properties {
		desc := ""
		if prop != nil {
			desc = prop.Description
		}
		out = append(out, docParam{
			Name:        name,
			Type:        schemaTypeLabel(prop),
			Required:    required[name],
			Description: desc,
		})
	}
	// Deterministic order: required first, then alphabetical.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Required != out[j].Required {
			return out[i].Required
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func buildResponses(responses map[string]ir.Response) []docResponse {
	if len(responses) == 0 {
		return nil
	}
	statuses := make([]string, 0, len(responses))
	for s := range responses {
		statuses = append(statuses, s)
	}
	sort.Strings(statuses)
	out := make([]docResponse, 0, len(statuses))
	for _, s := range statuses {
		out = append(out, docResponse{Status: s, Description: responses[s].Description})
	}
	return out
}

func buildAuth(reqs []ir.SecurityRequirement) []docAuth {
	if len(reqs) == 0 {
		return nil
	}
	out := make([]docAuth, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, docAuth{Name: r.Name, Scopes: r.Scopes})
	}
	return out
}

func codeExamples(res *ir.Resource) map[string]string {
	if res.Metadata == nil {
		return nil
	}
	if ex, ok := res.Metadata["code_examples"].(map[string]string); ok {
		return ex
	}
	return nil
}

// schemaTypeLabel renders a human-readable type label for a schema.
func schemaTypeLabel(s *ir.Schema) string {
	if s == nil {
		return "object"
	}
	switch s.Type {
	case "array":
		if s.Items != nil && s.Items.Type != "" {
			return "array<" + s.Items.Type + ">"
		}
		return "array"
	case "":
		if s.Ref != "" {
			return refName(s.Ref)
		}
		return "object"
	default:
		if len(s.Enum) > 0 {
			return "enum<" + s.Type + ">"
		}
		return s.Type
	}
}

func refName(ref string) string {
	if i := strings.LastIndex(ref, "/"); i != -1 {
		return ref[i+1:]
	}
	return ref
}

// pickResponseExample returns the status and pretty-printed example body for the
// code panel's response block, preferring the lowest 2xx response.
func (g *Generator) pickResponseExample(res *ir.Resource) (string, string) {
	// Prefer a 2xx response schema.
	statuses := make([]string, 0, len(res.Responses))
	for s := range res.Responses {
		statuses = append(statuses, s)
	}
	sort.Strings(statuses)

	for _, s := range statuses {
		if !strings.HasPrefix(s, "2") {
			continue
		}
		resp := res.Responses[s]
		schema := resp.Schema
		if schema == nil {
			for _, c := range resp.Content {
				schema = c
				break
			}
		}
		if schema != nil {
			if data := g.exampleGen.generateExampleData(schema); data != nil {
				return s, marshalPretty(data)
			}
		}
		return s, ""
	}

	// Fall back to an explicit example on the resource (events, websocket, etc.).
	if len(res.Examples) > 0 {
		return "", marshalPretty(res.Examples[0].Value)
	}
	if res.Response != nil {
		if data := g.exampleGen.generateExampleData(res.Response); data != nil {
			return "", marshalPretty(data)
		}
	}
	return "", ""
}

func marshalPretty(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}
