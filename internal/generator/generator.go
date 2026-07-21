package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirrobot01/unifydoc/internal/config"
	"github.com/sirrobot01/unifydoc/internal/ir"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

// Generator handles documentation generation
type Generator struct {
	config        *config.Config
	merger        *ir.Merger
	exampleGen    *ExampleGenerator
	searchIndexer *SearchIndexer
	minifier      *minify.M
	templates     *template.Template
}

// NewGenerator creates a new documentation generator
func NewGenerator(cfg *config.Config) *Generator {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("application/javascript", js.Minify)

	return &Generator{
		config:        cfg,
		merger:        ir.NewMerger(),
		exampleGen:    NewExampleGenerator(cfg.Features.CodeSnippets.Languages),
		searchIndexer: NewSearchIndexer(),
		minifier:      m,
	}
}

// AddIR adds an IR to the generator
func (g *Generator) AddIR(ir *ir.IR) {
	g.merger.Add(ir)
}

// Generate generates the documentation
func (g *Generator) Generate() error {
	// Merge all IRs
	merged, err := g.merger.Merge()
	if err != nil {
		return fmt.Errorf("failed to merge IRs: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(g.config.Output.Dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Load templates
	if err := g.loadTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Generate code examples for all resources
	g.generateCodeExamples(merged)

	// Build search index
	g.searchIndexer.Build(merged)

	// Generate HTML pages
	if err := g.generateHTML(merged); err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}

	// Copy assets
	if err := g.copyAssets(); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}

	// Write search index
	if g.config.Features.Search {
		if err := g.writeSearchIndex(); err != nil {
			return fmt.Errorf("failed to write search index: %w", err)
		}
	}

	return nil
}

// loadTemplates loads HTML templates
func (g *Generator) loadTemplates() error {
	g.templates = template.New("app").Funcs(g.templateFuncs())

	_, err := g.templates.Parse(appTemplate)
	return err
}

// templateFuncs returns template helper functions
func (g *Generator) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.MarshalIndent(v, "", "  ")
			return string(b)
		},
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"title": strings.Title,
		"join":  strings.Join,
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"statusClass": func(status string) string {
			if len(status) == 0 {
				return "2xx"
			}
			switch status[0] {
			case '1':
				return "1xx"
			case '2':
				return "2xx"
			case '3':
				return "3xx"
			case '4':
				return "4xx"
			case '5':
				return "5xx"
			default:
				return "2xx"
			}
		},
	}
}

// generateCodeExamples generates code examples for all resources
func (g *Generator) generateCodeExamples(merged *ir.MergedIR) {
	for _, protocol := range merged.Protocols {
		baseURL := ""
		if len(protocol.Servers) > 0 {
			baseURL = protocol.Servers[0].URL
		}

		for i := range protocol.Resources {
			resource := &protocol.Resources[i]
			examples := g.exampleGen.Generate(resource, protocol.Protocol, baseURL)

			if resource.Metadata == nil {
				resource.Metadata = make(map[string]interface{})
			}
			resource.Metadata["code_examples"] = examples
		}
	}
}

// generateHTML generates the single-page documentation app.
func (g *Generator) generateHTML(merged *ir.MergedIR) error {
	model := g.buildDocModel(merged)
	modelJSON, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("failed to marshal doc model: %w", err)
	}

	var buf bytes.Buffer
	data := map[string]interface{}{
		"Title": g.config.Project.Name,
		// template.JS injects the JSON verbatim into the script context.
		// json.Marshal already escapes <, > and & so there is no </script> risk.
		"Data": template.JS(modelJSON),
	}

	if err := g.templates.ExecuteTemplate(&buf, "app", data); err != nil {
		return err
	}

	// Minify HTML
	minified, err := g.minifier.String("text/html", buf.String())
	if err != nil {
		minified = buf.String() // Use unminified on error
	}

	outputPath := filepath.Join(g.config.Output.Dir, "index.html")
	return os.WriteFile(outputPath, []byte(minified), 0644)
}

// copyAssets copies static assets
func (g *Generator) copyAssets() error {
	assetsDir := filepath.Join(g.config.Output.Dir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return err
	}

	// Write CSS
	cssPath := filepath.Join(assetsDir, "style.css")
	minifiedCSS, err := g.minifier.String("text/css", appCSS)
	if err != nil {
		minifiedCSS = appCSS
	}
	if err := os.WriteFile(cssPath, []byte(minifiedCSS), 0644); err != nil {
		return err
	}

	// Write JavaScript
	jsPath := filepath.Join(assetsDir, "app.js")
	minifiedJS, err := g.minifier.String("application/javascript", appJS)
	if err != nil {
		minifiedJS = appJS
	}
	if err := os.WriteFile(jsPath, []byte(minifiedJS), 0644); err != nil {
		return err
	}

	return nil
}

// writeSearchIndex writes the search index JSON file
func (g *Generator) writeSearchIndex() error {
	indexPath := filepath.Join(g.config.Output.Dir, "assets", "search-index.json")
	indexData := g.searchIndexer.GetIndex()

	jsonData, err := json.Marshal(indexData)
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, jsonData, 0644)
}
