package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ChainID    string `json:"chainId"`
	DataDir    string `json:"dataDir,omitempty"`
	APIPort    int    `json:"apiPort,omitempty"`
	P2PPort    int    `json:"p2pPort,omitempty"`
	Seeds      []string `json:"seeds,omitempty"`
	EnableMDNS bool   `json:"enableMdns,omitempty"`
	DevMode    bool   `json:"devMode,omitempty"`
}

var DefaultConfig = Config{
	ChainID:    "wblue-mainnet-1",
	APIPort:    8080,
	P2PPort:    30303,
	EnableMDNS: true,
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := DefaultConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func LoadOrDefault(dataDir string) *Config {
	path := filepath.Join(dataDir, "config.json")
	cfg, err := Load(path)
	if err != nil {
		c := DefaultConfig
		c.DataDir = dataDir
		return &c
	}
	if cfg.DataDir == "" {
		cfg.DataDir = dataDir
	}
	return cfg
}

func (c *Config) Save(dataDir string) error {
	path := filepath.Join(dataDir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
