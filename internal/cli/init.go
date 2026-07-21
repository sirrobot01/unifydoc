package cli

import (
	"fmt"
	"os"

	"github.com/sirrobot01/unifydoc/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	autoDetect bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Unifidoc project",
	Long:  `Creates a new unifidoc.yaml configuration file in the current directory`,
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&autoDetect, "auto-detect", false, "auto-detect specification files")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := "unifidoc.yaml"

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("unifidoc.yaml already exists")
	}

	var cfg *config.Config

	if autoDetect {
		// Auto-detect specifications
		var err error
		cfg, err = config.DetectAndCreateConfig()
		if err != nil {
			return fmt.Errorf("auto-detection failed: %w", err)
		}
		if !quiet {
			fmt.Printf("Auto-detected %d protocol(s)\n", len(cfg.Protocols))
		}
	} else {
		// Create default config
		cfg = config.DefaultConfig()
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	if !quiet {
		fmt.Println("✓ Created unifidoc.yaml")
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Edit unifidoc.yaml to add your specifications")
		fmt.Println("  2. Run 'unifidoc generate' to build documentation")
		fmt.Println("  3. Run 'unifidoc serve' to preview locally")
	}

	return nil
}
