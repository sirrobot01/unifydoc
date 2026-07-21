package cli

import (
	"fmt"

	"github.com/sirrobot01/unifydoc/internal/plugin"
	"github.com/spf13/cobra"
)

var (
	builtinOnly   bool
	installedOnly bool
	allPlugins    bool
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
	Long:  `List, add, and manage Unifidoc plugins`,
}

var pluginListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List available plugins",
	Long:    `Lists all available plugins (built-in and installed)`,
	RunE:    runPluginList,
}

func init() {
	pluginListCmd.Flags().BoolVar(&builtinOnly, "builtin", false, "show only built-in plugins")
	pluginListCmd.Flags().BoolVar(&installedOnly, "installed", false, "show only installed plugins")
	pluginListCmd.Flags().BoolVar(&allPlugins, "all", true, "show all plugins")

	pluginCmd.AddCommand(pluginListCmd)
	rootCmd.AddCommand(pluginCmd)
}

func runPluginList(cmd *cobra.Command, args []string) error {
	// Initialize plugin manager and register plugins
	manager := plugin.NewManager()
	if err := registerPlugins(manager); err != nil {
		return fmt.Errorf("failed to register plugins: %w", err)
	}

	// Get plugin list
	plugins := manager.List()

	if len(plugins) == 0 {
		fmt.Println("No plugins found")
		return nil
	}

	fmt.Println("Available Plugins:")
	fmt.Println()

	for _, p := range plugins {
		fmt.Printf("  • %s (v%s)\n", p.Name, p.Version)
		if p.BuiltIn {
			fmt.Println("    Type: Built-in")
		}
		if len(p.Protocols) > 0 {
			fmt.Printf("    Protocols: %v\n", p.Protocols)
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d plugin(s)\n", len(plugins))
	return nil
}
