package cli

import (
	"fmt"
	"os"

	"github.com/sirrobot01/unifydoc/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	autoDetect    bool
	initFramework string
	initCI        string
	initWithHooks bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Unifidoc project",
	Long:  `Creates a new unifidoc.yaml configuration file in the current directory`,
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&autoDetect, "auto-detect", false, "auto-detect specification files")
	initCmd.Flags().StringVar(&initFramework, "framework", "", "scaffold for a framework: express | fastapi | gin | spring")
	initCmd.Flags().StringVar(&initCI, "ci", "", "also generate a CI workflow: github")
	initCmd.Flags().BoolVar(&initWithHooks, "with-hooks", false, "install git pre-commit and pre-push hooks")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := "unifidoc.yaml"

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("unifidoc.yaml already exists")
	}

	var cfg *config.Config

	switch {
	case initFramework != "":
		var err error
		cfg, err = buildFrameworkConfig(initFramework)
		if err != nil {
			return err
		}
		if !quiet {
			fmt.Printf("Scaffolded for %s\n", initFramework)
		}
	case autoDetect:
		// Auto-detect specifications
		var err error
		cfg, err = config.DetectAndCreateConfig()
		if err != nil {
			return fmt.Errorf("auto-detection failed: %w", err)
		}
		if !quiet {
			fmt.Printf("Auto-detected %d protocol(s)\n", len(cfg.Protocols))
		}
	default:
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

	// Optional CI workflow.
	if initCI != "" {
		path, err := writeCIWorkflow(initCI)
		if err != nil {
			return fmt.Errorf("failed to write CI workflow: %w", err)
		}
		if !quiet {
			fmt.Printf("✓ Created %s\n", path)
		}
	}

	// Optional git hooks.
	if initWithHooks {
		written, err := installGitHooks()
		if err != nil {
			return fmt.Errorf("failed to install git hooks: %w", err)
		}
		if !quiet {
			for _, path := range written {
				fmt.Printf("✓ Installed %s\n", path)
			}
		}
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
