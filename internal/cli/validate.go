package cli

import (
	"fmt"
	"os"

	"github.com/sirrobot01/unifydoc/internal/config"
	"github.com/sirrobot01/unifydoc/internal/plugin"
	"github.com/spf13/cobra"
)

var strict bool

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate specification files",
	Long:  `Validates all enabled protocol specifications for correctness`,
	RunE:  runValidate,
}

func init() {
	validateCmd.Flags().BoolVar(&strict, "strict", false, "strict validation mode")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
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

	// Initialize plugin manager
	manager := plugin.NewManager()
	if err := registerPlugins(manager); err != nil {
		return fmt.Errorf("failed to register plugins: %w", err)
	}

	hasErrors := false

	// Validate all enabled protocols
	for _, protocolCfg := range cfg.Protocols {
		if !protocolCfg.Enabled {
			continue
		}

		fmt.Printf("Validating %s: %s\n", protocolCfg.Plugin, protocolCfg.Spec)

		// Get plugin
		p, err := manager.Get(protocolCfg.Plugin)
		if err != nil {
			fmt.Printf("  ✗ Plugin '%s' not found\n", protocolCfg.Plugin)
			hasErrors = true
			continue
		}

		// Read spec file
		specData, err := os.ReadFile(protocolCfg.Spec)
		if err != nil {
			fmt.Printf("  ✗ Failed to read file: %v\n", err)
			hasErrors = true
			continue
		}

		// Validate spec
		if err := p.Validate(specData); err != nil {
			fmt.Printf("  ✗ Validation failed: %v\n", err)
			hasErrors = true
			continue
		}

		fmt.Println("  ✓ Valid")
	}

	if hasErrors {
		return fmt.Errorf("validation failed for one or more specifications")
	}

	fmt.Println("\n✓ All specifications are valid!")
	return nil
}
