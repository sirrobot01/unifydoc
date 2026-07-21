package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/sirrobot01/unifydoc/internal/config"
	"github.com/sirrobot01/unifydoc/internal/server"
	"github.com/spf13/cobra"
)

var (
	servePort int
	noOpen    bool
	noWatch   bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a local development server with live reload",
	Long: `Serve generated documentation over HTTP. When live reload is enabled,
spec files are watched and the browser refreshes automatically on changes.`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 0, "server port (overrides config)")
	serveCmd.Flags().BoolVar(&noOpen, "no-open", false, "do not open the browser automatically")
	serveCmd.Flags().BoolVar(&noWatch, "no-watch", false, "disable file watching and live reload")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
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

	if servePort != 0 {
		cfg.Server.Port = servePort
	}

	liveReload := cfg.Server.LiveReload && !noWatch

	// Build once up front so there is something to serve.
	if err := buildDocs(cfg, quiet); err != nil {
		return fmt.Errorf("initial build failed: %w", err)
	}

	srv := &server.Server{
		Dir:        cfg.Output.Dir,
		Port:       cfg.Server.Port,
		LiveReload: liveReload,
		Watch:      watchTargets(cfg),
		Rebuild:    func() error { return buildDocs(cfg, true) },
		Quiet:      quiet,
	}

	if !noOpen {
		openBrowser(fmt.Sprintf("http://localhost:%d", cfg.Server.Port))
	}

	return srv.ListenAndServe()
}

// watchTargets returns the spec files (and config file) the dev server should
// watch for changes.
func watchTargets(cfg *config.Config) []string {
	targets := make([]string, 0, len(cfg.Protocols)+1)
	for _, p := range cfg.Protocols {
		if p.Enabled && p.Spec != "" {
			targets = append(targets, p.Spec)
		}
	}
	if configFile != "" {
		targets = append(targets, configFile)
	}
	return targets
}

// openBrowser attempts to open url in the default browser. Failures are silent
// since the server is usable without it.
func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	default:
		cmd = "xdg-open"
	}

	args = append(args, url)
	_ = exec.Command(cmd, args...).Start()
}
