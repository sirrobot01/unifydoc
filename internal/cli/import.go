package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirrobot01/unifydoc/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	importFrom   string
	importSpec   string
	importOutput string
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import documentation from another tool",
	Long: `Import an existing specification from another tool into a Unifidoc
project. Supported sources:

  postman   Convert a Postman v2.x collection into a Unifidoc API spec
  openapi   Register an existing OpenAPI 3.x spec (no conversion)`,
	RunE: runImport,
}

func init() {
	importCmd.Flags().StringVar(&importFrom, "from", "", "source tool: postman | openapi")
	importCmd.Flags().StringVar(&importSpec, "spec", "", "path to the source specification file")
	importCmd.Flags().StringVar(&importOutput, "output", "", "output spec file (postman only; default <name>.api.yaml)")
	_ = importCmd.MarkFlagRequired("from")
	_ = importCmd.MarkFlagRequired("spec")
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(importSpec)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", importSpec, err)
	}

	configPath := resolveConfigPath()
	cfg, err := loadOrDefaultConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch strings.ToLower(importFrom) {
	case "postman":
		specYAML, title, err := convertPostman(data)
		if err != nil {
			return fmt.Errorf("failed to convert Postman collection: %w", err)
		}
		out := importOutput
		if out == "" {
			out = slugify(title) + ".api.yaml"
		}
		if err := os.WriteFile(out, specYAML, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", out, err)
		}
		upsertProtocol(cfg, config.ProtocolConfig{Plugin: "api", Spec: out, Enabled: true})
		if !quiet {
			fmt.Printf("✓ Converted Postman collection %q -> %s\n", title, out)
		}
	case "openapi", "swagger":
		upsertProtocol(cfg, config.ProtocolConfig{Plugin: "openapi", Spec: importSpec, Enabled: true})
		if !quiet {
			fmt.Printf("✓ Registered OpenAPI spec %s\n", importSpec)
		}
	default:
		return fmt.Errorf("unsupported source %q (supported: postman, openapi)", importFrom)
	}

	if err := saveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	if !quiet {
		fmt.Printf("✓ Updated %s\n", configPath)
	}
	return nil
}

// ---------- Postman conversion ----------

type postmanCollection struct {
	Info postmanInfo   `json:"info"`
	Item []postmanItem `json:"item"`
}

type postmanInfo struct {
	Name string `json:"name"`
}

type postmanItem struct {
	Name    string          `json:"name"`
	Request *postmanRequest `json:"request"`
	Item    []postmanItem   `json:"item"` // sub-folder
}

type postmanRequest struct {
	Method      string          `json:"method"`
	URL         json.RawMessage `json:"url"`
	Description json.RawMessage `json:"description"`
}

// apiSpecFile mirrors the `api` plugin's YAML shape.
type apiSpecFile struct {
	API apiConfig `yaml:"api"`
}

type apiConfig struct {
	Title     string        `yaml:"title"`
	Version   string        `yaml:"version"`
	Endpoints []apiEndpoint `yaml:"endpoints"`
}

type apiEndpoint struct {
	Name        string   `yaml:"name"`
	Method      string   `yaml:"method"`
	Path        string   `yaml:"path"`
	Description string   `yaml:"description,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

// convertPostman converts a Postman v2.x collection into an `api` spec YAML,
// returning the YAML bytes and the collection title.
func convertPostman(data []byte) ([]byte, string, error) {
	var coll postmanCollection
	if err := json.Unmarshal(data, &coll); err != nil {
		return nil, "", fmt.Errorf("invalid Postman JSON: %w", err)
	}
	if coll.Info.Name == "" && len(coll.Item) == 0 {
		return nil, "", fmt.Errorf("not a Postman collection")
	}

	title := coll.Info.Name
	if title == "" {
		title = "Imported API"
	}

	spec := apiSpecFile{API: apiConfig{Title: title, Version: "1.0.0"}}
	flattenPostman(coll.Item, nil, &spec.API.Endpoints)

	out, err := yaml.Marshal(spec)
	if err != nil {
		return nil, "", err
	}
	return out, title, nil
}

// flattenPostman walks the (possibly nested) Postman item tree, collecting
// request items as endpoints and using folder names as tags.
func flattenPostman(items []postmanItem, folder []string, out *[]apiEndpoint) {
	for _, item := range items {
		if len(item.Item) > 0 { // folder
			flattenPostman(item.Item, append(folder, item.Name), out)
			continue
		}
		if item.Request == nil {
			continue
		}
		ep := apiEndpoint{
			Name:        item.Name,
			Method:      strings.ToUpper(item.Request.Method),
			Path:        postmanPath(item.Request.URL),
			Description: postmanDescription(item.Request.Description),
		}
		if ep.Method == "" {
			ep.Method = "GET"
		}
		if len(folder) > 0 {
			ep.Tags = []string{folder[len(folder)-1]}
		}
		*out = append(*out, ep)
	}
}

// postmanPath extracts a request path from Postman's url field, which may be a
// string or an object with a path array.
func postmanPath(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "/"
	}
	// Try object form first.
	var obj struct {
		Raw  string   `json:"raw"`
		Path []string `json:"path"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		if len(obj.Path) > 0 {
			return "/" + strings.Join(obj.Path, "/")
		}
		if obj.Raw != "" {
			return stripHost(obj.Raw)
		}
	}
	// Fall back to string form.
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && s != "" {
		return stripHost(s)
	}
	return "/"
}

// postmanDescription accepts either a string or a {content} object.
func postmanDescription(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var obj struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj.Content
	}
	return ""
}

// stripHost reduces a full URL to its path (and drops any query string).
func stripHost(u string) string {
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	if i := strings.Index(u, "/"); i != -1 {
		u = u[i:]
	} else {
		u = "/"
	}
	if i := strings.IndexAny(u, "?#"); i != -1 {
		u = u[:i]
	}
	// Drop Postman template variables' host artifacts like {{baseUrl}}.
	u = strings.ReplaceAll(u, "{{", "")
	u = strings.ReplaceAll(u, "}}", "")
	if u == "" {
		return "/"
	}
	return u
}

// slugify turns a title into a filename-safe slug.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteRune('-')
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "imported"
	}
	return slug
}
