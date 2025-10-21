package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIToken      string `json:"api_token"`
	VPNDirectory  string `json:"vpn_directory"`
	LastServerID  int    `json:"last_server_id"`
	LastProtocol  string `json:"last_protocol"`
	WindowWidth   int    `json:"window_width"`
	WindowHeight  int    `json:"window_height"`
}

var configPath string

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "htb-tool")
	os.MkdirAll(configDir, 0755)
	configPath = filepath.Join(configDir, "config.json")
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config
			return &Config{
				VPNDirectory: filepath.Join(os.Getenv("HOME"), "Downloads", "htb-vpn"),
				LastProtocol: "tcp",
				WindowWidth:  1200,
				WindowHeight: 800,
			}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return configPath
}
