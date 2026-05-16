package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AppConfig struct {
	VpsIP      string `json:"vps_ip"`
	VpsUser    string `json:"vps_user"`
	VpsKeyPath string `json:"vps_key_path"`
}

func LoadConfig() (AppConfig, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".gopipe.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}
	var config AppConfig
	err = json.Unmarshal(data, &config)
	return config, err
}

func SaveConfig(config AppConfig) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".gopipe.json")
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(path, data, 0600)
}
