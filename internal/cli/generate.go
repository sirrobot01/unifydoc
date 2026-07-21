package cli

import (
	"fmt"
	"os"

	"github.com/sirrobot01/unifydoc/internal/config"
	"github.com/sirrobot01/unifydoc/internal/generator"
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

		// Read spec file
		specData, err := os.ReadFile(protocolCfg.Spec)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to read %s: %v\n", protocolCfg.Spec, err)
			continue
		}

		// Validate spec
		if err := p.Validate(specData); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: validation failed for %s: %v\n", protocolCfg.Spec, err)
			//continue
		}

		// Parse spec
		ir, err := p.Parse(specData)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: parsing failed for %s: %v\n", protocolCfg.Spec, err)
			continue
		}

		// Add to generator
		gen.AddIR(ir)

		if !quiet {
			fmt.Printf("  ✓ Parsed %d resources\n", len(ir.Resources))
		}
	}

	// Generate documentation
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	return nil
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
