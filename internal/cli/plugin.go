package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sirrobot01/unifydoc/internal/config"
	"github.com/sirrobot01/unifydoc/internal/plugin"
	"github.com/spf13/cobra"
)

var (
	builtinOnly   bool
	installedOnly bool
	allPlugins    bool

	pluginAddSpec     string
	pluginAddTemplate string
	pluginAddDisabled bool
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

var pluginAddCmd = &cobra.Command{
	Use:   "add <plugin>",
	Short: "Add a protocol plugin to the project config",
	Long: `Adds a built-in protocol plugin to unifidoc.yaml, pointing it at a
spec file. This is a convenience over editing the config by hand.`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginAdd,
}

func init() {
	pluginListCmd.Flags().BoolVar(&builtinOnly, "builtin", false, "show only built-in plugins")
	pluginListCmd.Flags().BoolVar(&installedOnly, "installed", false, "show only installed plugins")
	pluginListCmd.Flags().BoolVar(&allPlugins, "all", true, "show all plugins")

	pluginAddCmd.Flags().StringVar(&pluginAddSpec, "spec", "", "path to the spec file for this protocol")
	pluginAddCmd.Flags().StringVar(&pluginAddTemplate, "template", "", "optional custom template file")
	pluginAddCmd.Flags().BoolVar(&pluginAddDisabled, "disabled", false, "add the protocol but leave it disabled")
	_ = pluginAddCmd.MarkFlagRequired("spec")

	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginAddCmd)
	rootCmd.AddCommand(pluginCmd)
}

func runPluginAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	known := knownPlugins()
	if !known[name] {
		return fmt.Errorf("unknown plugin %q. Available built-in plugins: %s\n"+
			"(external plugins are not yet supported)", name, strings.Join(sortedSet(known), ", "))
	}

	if _, err := os.Stat(pluginAddSpec); err != nil && !quiet {
		fmt.Fprintf(os.Stderr, "Warning: spec %s does not exist yet\n", pluginAddSpec)
	}

	configPath := resolveConfigPath()
	cfg, err := loadOrDefaultConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	upsertProtocol(cfg, config.ProtocolConfig{
		Plugin:   name,
		Spec:     pluginAddSpec,
		Template: pluginAddTemplate,
		Enabled:  !pluginAddDisabled,
	})

	if err := saveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	if !quiet {
		fmt.Printf("✓ Added %s plugin (%s) to %s\n", name, pluginAddSpec, configPath)
	}
	return nil
}

// sortedSet returns the keys of a set in sorted order.
func sortedSet(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
