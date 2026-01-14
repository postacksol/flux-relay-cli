package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fluxrelay/flux-relay-cli/internal/api"
)

type Config struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	DeveloperID  string    `json:"developer_id"`
	Email        string    `json:"email"`
	APIURL       string    `json:"api_url,omitempty"`
}

type ConfigManager struct {
	configPath string
}

func New() *ConfigManager {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".flux-relay")
	configPath := filepath.Join(configDir, "config.json")

	return &ConfigManager{
		configPath: configPath,
	}
}

func (cm *ConfigManager) ConfigPath() string {
	return cm.configPath
}

func (cm *ConfigManager) GetToken() (*Config, error) {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No config file is not an error
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Check if token is expired
	if time.Now().After(config.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	return &config, nil
}

func (cm *ConfigManager) SaveToken(token *api.TokenResponse) error {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	config := Config{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    expiresAt,
		DeveloperID:  token.Developer.ID,
		Email:        token.Developer.Email,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write with secure permissions (read/write for owner only)
	return os.WriteFile(cm.configPath, data, 0600)
}

func (cm *ConfigManager) RemoveToken() error {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to remove
	}
	return os.Remove(cm.configPath)
}

func (cm *ConfigManager) GetAccessToken() string {
	config, err := cm.GetToken()
	if err != nil || config == nil {
		return ""
	}
	return config.AccessToken
}
