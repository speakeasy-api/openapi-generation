package licensebootstrap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultServerURL = "https://app.speakeasy.com"

var ErrNoAPIKey = errors.New("no Speakeasy API key: set SPEAKEASY_API_KEY or run `speakeasy auth login`")

func ResolveAPIKey(getenv func(string) string, homeDir string) (string, error) {
	if key := strings.TrimSpace(getenv("SPEAKEASY_API_KEY")); key != "" {
		return key, nil
	}

	configPath := filepath.Join(homeDir, ".speakeasy", "config.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNoAPIKey
		}
		return "", fmt.Errorf("read %s: %w", configPath, err)
	}

	var config struct {
		APIKey string `yaml:"speakeasy_api_key"`
	}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return "", fmt.Errorf("parse %s: %w", configPath, err)
	}
	if key := strings.TrimSpace(config.APIKey); key != "" {
		return key, nil
	}
	return "", ErrNoAPIKey
}

func ResolveServerURL(flagValue string, getenv func(string) string) string {
	for _, candidate := range []string{flagValue, getenv("SPEAKEASY_SERVER_URL"), DefaultServerURL} {
		if candidate = strings.TrimSpace(candidate); candidate != "" {
			return strings.TrimRight(candidate, "/")
		}
	}
	return DefaultServerURL
}
