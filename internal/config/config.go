package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	configName = "config"
	configType = "yaml"
	appName    = "starsage"
)

// Config holds the application's configuration
type Config struct {
	GitHubToken string `mapstructure:"github_token"`
}

// InitConfig initializes viper to read from the config file.
func InitConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %w", err)
	}

	configPath := filepath.Join(home, ".config", appName)
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	// Create config file if it doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create it
			if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
				return fmt.Errorf("could not create config directory: %w", err)
			}
			if err := viper.SafeWriteConfig(); err != nil {
				return fmt.Errorf("could not create config file: %w", err)
			}
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("error reading config file: %w", err)
		}
	}
	return nil
}

// SaveToken saves the GitHub token to the config file.
func SaveToken(token string) error {
	viper.Set("github_token", token)
	return viper.WriteConfig()
}

// GetToken retrieves the GitHub token from the config file.
func GetToken() string {
	return viper.GetString("github_token")
}
