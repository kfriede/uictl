package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for uictl.
type Config struct {
	Host     string `mapstructure:"host"`
	Site     string `mapstructure:"site"`
	Profile  string `mapstructure:"profile"`
	APIKey   string `mapstructure:"api_key"`
	Insecure bool   `mapstructure:"insecure"`
	Verbose  bool   `mapstructure:"verbose"`
	Debug    bool   `mapstructure:"debug"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Site: "default",
	}
}

// Load reads configuration from XDG config dir, env vars, and the named profile.
func Load(profile string) (*Config, error) {
	v := viper.GetViper()

	configDir := configDir()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)
	v.AddConfigPath(".")

	// Profile-specific config overrides the base config
	if profile != "" {
		v.SetConfigName(profile)
	}

	v.SetEnvPrefix("UICTL")
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("site", "default")
	v.SetDefault("insecure", false)

	cfg := Default()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return cfg, fmt.Errorf("reading config: %w", err)
		}
		// No config file is fine — use defaults + env vars
	}

	if err := v.Unmarshal(cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	cfg.Profile = profile
	return cfg, nil
}

// Dir returns the XDG-compliant config directory for uictl.
func Dir() string {
	return configDir()
}

func configDir() string {
	if d := os.Getenv("UICTL_CONFIG"); d != "" {
		return d
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "uictl")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".uictl"
	}
	return filepath.Join(home, ".config", "uictl")
}

// EnsureDir creates the config directory if it doesn't exist.
func EnsureDir() error {
	dir := configDir()
	return os.MkdirAll(dir, 0700)
}

// Save writes the config to the config directory.
func Save(cfg *Config) error {
	if err := EnsureDir(); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	v := viper.New()
	v.Set("host", cfg.Host)
	v.Set("site", cfg.Site)
	v.Set("insecure", cfg.Insecure)

	name := "config"
	if cfg.Profile != "" {
		name = cfg.Profile
	}

	configPath := filepath.Join(configDir(), name+".yaml")
	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	// Restrictive permissions — config may contain non-secret settings,
	// but we protect it anyway since API keys fall back here.
	return os.Chmod(configPath, 0600)
}
