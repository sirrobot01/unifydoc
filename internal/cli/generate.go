package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirrobot01/unifydoc/internal/config"
	"github.com/sirrobot01/unifydoc/internal/generator"
	"github.com/sirrobot01/unifydoc/internal/ir"
	"github.com/sirrobot01/unifydoc/internal/plugin"
	"github.com/sirrobot01/unifydoc/internal/plugins/api"
	"github.com/sirrobot01/unifydoc/internal/plugins/asyncapi"
	"github.com/sirrobot01/unifydoc/internal/plugins/custom"
	"github.com/sirrobot01/unifydoc/internal/plugins/events"
	"github.com/sirrobot01/unifydoc/internal/plugins/grpc"
	"github.com/sirrobot01/unifydoc/internal/plugins/mcp"
	"github.com/sirrobot01/unifydoc/internal/plugins/openapi"
	"github.com/sirrobot01/unifydoc/internal/plugins/webhook"
	"github.com/sirrobot01/unifydoc/internal/plugins/websocket"
	"github.com/spf13/cobra"
)

var (
	outputDir string
	watch     bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate documentation from specifications",
	Long:  `Parses all enabled protocol specifications and generates unified HTML documentation`,
	RunE:  runGenerate,
}

func init() {
	generateCmd.Flags().StringVar(&outputDir, "output", "", "output directory (overrides config)")
	generateCmd.Flags().BoolVar(&watch, "watch", false, "watch for changes and regenerate")
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Load configuration
	var cfg *config.Config
	var err error

	if configFile != "" {
		cfg, err = config.LoadFromFile(configFile)
	} else {
		cfg, err = config.LoadDefault()
	}

	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override output directory if specified
	if outputDir != "" {
		cfg.Output.Dir = outputDir
	}

	if !quiet {
		fmt.Println("Generating documentation...")
		fmt.Printf("Output directory: %s\n", cfg.Output.Dir)
	}

	if err := buildDocs(cfg, quiet); err != nil {
		return err
	}

	if !quiet {
		fmt.Println("\n✓ Documentation generated successfully!")
		fmt.Printf("  Open %s/index.html in your browser\n", cfg.Output.Dir)
	}

	return nil
}

// buildDocs runs the full parse-and-generate pipeline for a config. It is
// shared by the generate and serve commands so both stay in sync.
func buildDocs(cfg *config.Config, quiet bool) error {
	// Initialize plugin manager
	manager := plugin.NewManager()

	// Register all built-in plugins
	if err := registerPlugins(manager); err != nil {
		return fmt.Errorf("failed to register plugins: %w", err)
	}

	// Create generator
	gen := generator.NewGenerator(cfg)

	// Parse all enabled protocols
	for _, protocolCfg := range cfg.Protocols {
		if !protocolCfg.Enabled {
			continue
		}

		if !quiet {
			fmt.Printf("Processing %s: %s\n", protocolCfg.Plugin, protocolCfg.Spec)
		}

		// Get plugin
		p, err := manager.Get(protocolCfg.Plugin)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: plugin '%s' not found, skipping\n", protocolCfg.Plugin)
			continue
		}

		result, err := parseProtocolSpec(p, protocolCfg.Spec)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			continue
		}

		// Add to generator
		gen.AddIR(result)

		if !quiet {
			fmt.Printf("  ✓ Parsed %d resources\n", len(result.Resources))
		}
	}

	// Generate documentation
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	return nil
}

// specExtensions are the file types treated as specs when a protocol points at
// a directory (gRPC handles its own .proto discovery via PathParser).
var specExtensions = map[string]bool{".yaml": true, ".yml": true, ".json": true}

// parseProtocolSpec parses a protocol's configured spec into a single IR. It
// supports three cases: plugins that implement PathParser (which handle files
// and directories themselves), a directory of specs (parsed and merged), and a
// single spec file.
func parseProtocolSpec(p plugin.Plugin, specPath string) (*ir.IR, error) {
	// Plugins that need filesystem access handle the path directly.
	if pp, ok := p.(plugin.PathParser); ok {
		result, err := pp.ParsePath(specPath)
		if err != nil {
			return nil, fmt.Errorf("parsing failed for %s: %w", specPath, err)
		}
		return result, nil
	}

	info, err := os.Stat(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", specPath, err)
	}
	if info.IsDir() {
		return parseSpecDir(p, specPath)
	}
	return parseSpecFile(p, specPath)
}

// parseSpecFile reads, validates and parses a single spec file.
func parseSpecFile(p plugin.Plugin, path string) (*ir.IR, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	if err := p.Validate(data); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: validation failed for %s: %v\n", path, err)
	}
	result, err := p.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing failed for %s: %w", path, err)
	}
	return result, nil
}

// parseSpecDir parses every spec file in a directory and merges the results
// into a single IR so the protocol renders as one section.
func parseSpecDir(p plugin.Plugin, dir string) (*ir.IR, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if specExtensions[strings.ToLower(filepath.Ext(e.Name()))] {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files) // deterministic order

	if len(files) == 0 {
		return nil, fmt.Errorf("no spec files found in %s", dir)
	}

	var merged *ir.IR
	for _, f := range files {
		got, err := parseSpecFile(p, f)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			continue
		}
		merged = mergeIR(merged, got)
	}
	if merged == nil {
		return nil, fmt.Errorf("no spec files in %s could be parsed", dir)
	}
	return merged, nil
}

// mergeIR combines src into dst, aggregating resources, types, servers and
// security. The first non-empty title/description/version wins.
func mergeIR(dst, src *ir.IR) *ir.IR {
	if dst == nil {
		return src
	}
	dst.Resources = append(dst.Resources, src.Resources...)
	dst.Types = append(dst.Types, src.Types...)
	dst.Servers = append(dst.Servers, src.Servers...)
	dst.Security = append(dst.Security, src.Security...)
	if dst.Title == "" {
		dst.Title = src.Title
	}
	if dst.Description == "" {
		dst.Description = src.Description
	}
	if dst.Version == "" {
		dst.Version = src.Version
	}
	return dst
}

func registerPlugins(manager *plugin.Manager) error {
	plugins := []plugin.Plugin{
		openapi.NewPlugin(),
		grpc.NewPlugin(),
		websocket.NewPlugin(),
		webhook.NewPlugin(),
		events.NewPlugin(),
		mcp.NewPlugin(),
		api.NewPlugin(),
		asyncapi.NewPlugin(),
		custom.NewPlugin(),
	}

	for _, p := range plugins {
		if err := manager.Register(p); err != nil {
			return err
		}
	}

	return nil
}
