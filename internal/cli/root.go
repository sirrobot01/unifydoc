package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
	verbose    bool
	quiet      bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "unifidoc",
	Short: "Unified documentation generator for all protocols",
	Long: `Unifidoc is a CLI tool that generates unified, professional documentation
from multiple protocol specifications through a plugin-based architecture.

Supports: OpenAPI, gRPC, WebSocket, AsyncAPI, Webhooks, Events, MCP, and custom protocols.`,
	Version: version,
}

// Execute executes the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ./unifidoc.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-error output")
}
