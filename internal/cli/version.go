package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// version is the current unifidoc version. Kept in sync with the root command.
const version = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("unifidoc %s\n", version)
		fmt.Printf("  go:   %s\n", runtime.Version())
		fmt.Printf("  os:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
