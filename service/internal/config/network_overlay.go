package config

import (
	"strings"

	"gopkg.in/yaml.v2"
)

func applyNetworkOverlay(file []byte, cfg *Config) {
	if cfg == nil || cfg.Network == nil || len(file) == 0 {
		return
	}

	var overlay struct {
		Network *NetworkConfig `yaml:"network"`
	}
	if err := yaml.Unmarshal(file, &overlay); err != nil || overlay.Network == nil {
		return
	}

	if v := strings.TrimSpace(overlay.Network.BindGrpc); v != "" {
		cfg.Network.BindGrpc = v
	}
	if v := strings.TrimSpace(overlay.Network.BindRest); v != "" {
		cfg.Network.BindRest = v
	}
	if v := strings.TrimSpace(overlay.Network.BindProxy); v != "" {
		cfg.Network.BindProxy = v
	}
}
