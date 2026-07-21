package cli

import (
	"os"

	"github.com/sirrobot01/unifydoc/internal/config"
	"github.com/sirrobot01/unifydoc/internal/plugin"
	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "unifidoc.yaml"

// resolveConfigPath returns the config file to operate on: the --config flag if
// set, otherwise the conventional unifidoc.yaml in the working directory.
func resolveConfigPath() string {
	if configFile != "" {
		return configFile
	}
	return defaultConfigPath
}

// loadOrDefaultConfig loads an existing config file, or returns defaults if the
// file does not exist yet.
func loadOrDefaultConfig(path string) (*config.Config, error) {
	if _, err := os.Stat(path); err != nil {
		return config.DefaultConfig(), nil
	}
	return config.LoadFromFile(path)
}

// saveConfig writes a config to disk as YAML.
func saveConfig(path string, cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// upsertProtocol adds a protocol entry, replacing any existing entry with the
// same plugin and spec so the operation is idempotent.
func upsertProtocol(cfg *config.Config, pc config.ProtocolConfig) {
	for i := range cfg.Protocols {
		if cfg.Protocols[i].Plugin == pc.Plugin && cfg.Protocols[i].Spec == pc.Spec {
			cfg.Protocols[i] = pc
			return
		}
	}
	cfg.Protocols = append(cfg.Protocols, pc)
}

// knownPlugins returns the set of registered built-in plugin names.
func knownPlugins() map[string]bool {
	manager := plugin.NewManager()
	_ = registerPlugins(manager)
	set := make(map[string]bool)
	for _, meta := range manager.List() {
		set[meta.Name] = true
	}
	return set
}
