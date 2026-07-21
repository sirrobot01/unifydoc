package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirrobot01/unifydoc/internal/config"
)

// frameworkSpecs maps a framework name to the spec path it conventionally
// exposes an OpenAPI document at.
var frameworkSpecs = map[string]string{
	"express": "./openapi.json",
	"fastapi": "./openapi.json",
	"gin":     "./docs/swagger.yaml",
	"spring":  "./openapi.yaml",
}

// supportedFrameworks lists the frameworks buildFrameworkConfig accepts.
func supportedFrameworks() []string {
	return []string{"express", "fastapi", "gin", "spring"}
}

// buildFrameworkConfig returns a config pre-wired for a known framework, or an
// error if the framework is not recognized.
func buildFrameworkConfig(framework string) (*config.Config, error) {
	spec, ok := frameworkSpecs[framework]
	if !ok {
		return nil, fmt.Errorf("unknown framework %q (supported: express, fastapi, gin, spring)", framework)
	}
	cfg := config.DefaultConfig()
	cfg.Project.Name = fmt.Sprintf("%s API Documentation", framework)
	cfg.Protocols = []config.ProtocolConfig{
		{Plugin: "openapi", Spec: spec, Enabled: true},
	}
	return cfg, nil
}

// ciWorkflow is the GitHub Actions workflow written by `init --ci github`.
const ciWorkflow = `name: Documentation

on:
  push:
    branches: [main]

permissions:
  contents: write

jobs:
  docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 'stable'
      - name: Install unifidoc
        run: go install github.com/sirrobot01/unifydoc/cmd/unifidoc@latest
      - name: Generate documentation
        run: unifidoc generate
      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v4
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs
`

// writeCIWorkflow writes the CI workflow for the given provider.
func writeCIWorkflow(provider string) (string, error) {
	if provider != "github" {
		return "", fmt.Errorf("unsupported CI provider %q (supported: github)", provider)
	}
	dir := filepath.Join(".github", "workflows")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "docs.yml")
	if err := os.WriteFile(path, []byte(ciWorkflow), 0644); err != nil {
		return "", err
	}
	return path, nil
}

const preCommitHook = `#!/bin/sh
# unifidoc: validate specs before committing
unifidoc validate || {
  echo "unifidoc: spec validation failed" >&2
  exit 1
}
`

const prePushHook = `#!/bin/sh
# unifidoc: regenerate documentation before pushing
unifidoc generate
`

// installGitHooks writes pre-commit and pre-push hooks into the repo's
// .git/hooks directory. It errors if the working directory is not a git repo.
func installGitHooks() ([]string, error) {
	hooksDir := filepath.Join(".git", "hooks")
	if _, err := os.Stat(".git"); err != nil {
		return nil, fmt.Errorf("not a git repository (no .git directory)")
	}
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return nil, err
	}

	hooks := map[string]string{
		"pre-commit": preCommitHook,
		"pre-push":   prePushHook,
	}
	written := make([]string, 0, len(hooks))
	for name, content := range hooks {
		path := filepath.Join(hooksDir, name)
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			return written, err
		}
		written = append(written, path)
	}
	return written, nil
}
