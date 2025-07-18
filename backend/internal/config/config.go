package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Log     LoggingConfig  `mapstructure:"log"`
	Piholes []PiholeConfig `mapstructure:"piholes"`
	Server  ServerConfig   `mapstructure:"server"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

type PiholeConfig struct {
	ID       string `mapstructure:"id"`
	Scheme   string `mapstructure:"scheme"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

type ServerConfig struct {
	Port  int         `mapstructure:"port"`
	Proxy ProxyConfig `mapstructure:"proxy"`
}

type ProxyConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Hostname string `mapstructure:"hostname"`
	Port     int    `mapstructure:"port"`
}

func Load() (*Config, error) {
	if err := initConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	cfg.Piholes = buildPiholeConfigsFromEnv()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func initConfig() error {
	// Respect the --config CLI flag if set
	if cfgFile := viper.GetString("config"); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Default config file name
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// Add common config paths
		if configDir, err := os.UserConfigDir(); err == nil {
			viper.AddConfigPath(filepath.Join(configDir, "pihole-cluster-admin"))
		}
		viper.AddConfigPath("/etc/pihole-cluster-admin")
		viper.AddConfigPath("/config")
		viper.AddConfigPath(".")
	}

	// Environment variable support
	viper.SetEnvPrefix("PIHOLE_CLUSTER_ADMIN")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set Viper defaults
	viper.SetDefault("log.level", "INFO")
	viper.SetDefault("piholes", []map[string]interface{}{})
	viper.SetDefault("server.port", 8081)
	viper.SetDefault("server.proxy.enable", false)
	viper.SetDefault("server.proxy.hostname", "localhost")
	viper.SetDefault("server.proxy.port", 5173)

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	return nil
}

func buildPiholeConfigsFromEnv() []PiholeConfig {
	var result []PiholeConfig
	for i := 0; ; i++ {
		prefix := fmt.Sprintf("PIHOLE_CLUSTER_ADMIN_PIHOLES_%d_", i)

		id := os.Getenv(prefix + "ID")
		host := os.Getenv(prefix + "HOST")
		pass := os.Getenv(prefix + "PASSWORD")

		if id == "" && host == "" && pass == "" {
			break // no more entries
		}

		port := 80
		if p := os.Getenv(prefix + "PORT"); p != "" {
			fmt.Sscanf(p, "%d", &port)
		}

		scheme := os.Getenv(prefix + "SCHEME")
		if scheme == "" {
			scheme = "http"
		}

		result = append(result, PiholeConfig{
			ID:       id,
			Host:     host,
			Port:     port,
			Scheme:   scheme,
			Password: pass,
		})
	}
	return result
}

// validate checks for config consistency.
func (c *Config) validate() error {
	validLevels := map[string]struct{}{
		"TRACE": {}, "DEBUG": {}, "INFO": {}, "WARN": {}, "ERROR": {}, "FATAL": {},
	}
	if _, ok := validLevels[strings.ToUpper(c.Log.Level)]; !ok {
		return fmt.Errorf("log.level must be a valid log level, got: %s", c.Log.Level)
	}
	if c.Server.Proxy.Hostname == "" {
		return fmt.Errorf("proxy.hostname cannot be empty")
	}
	if c.Server.Proxy.Port <= 0 || c.Server.Proxy.Port > 65535 {
		return fmt.Errorf("proxy.port must be a valid TCP port")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be a valid TCP port")
	}

	// Validate pihole configurations
	if len(c.Piholes) == 0 {
		return fmt.Errorf("at least one pihole instance must be configured")
	}

	seenIDs := make(map[string]struct{})
	seenHostsPorts := make(map[string]struct{})
	for i, node := range c.Piholes {
		if strings.TrimSpace(node.ID) == "" {
			return fmt.Errorf("pihole[%d]: id cannot be empty", i)
		}

		// Dedupe by ID
		if _, exists := seenIDs[node.ID]; exists {
			return fmt.Errorf("pihole[%d]: duplicate id '%s'", i, node.ID)
		}
		seenIDs[node.ID] = struct{}{}

		// Dedupe by host:port
		hostPort := fmt.Sprintf("%s:%d", node.Host, node.Port)
		if _, exists := seenHostsPorts[hostPort]; exists {
			return fmt.Errorf("pihole[%d]: duplicate host:port '%s'", i, hostPort)
		}
		seenHostsPorts[hostPort] = struct{}{}

		if strings.TrimSpace(node.Host) == "" {
			return fmt.Errorf("pihole[%d]: host cannot be empty", i)
		}
		if node.Port <= 0 || node.Port > 65535 {
			return fmt.Errorf("pihole[%d]: port must be a valid TCP port", i)
		}
		if strings.TrimSpace(node.Password) == "" {
			return fmt.Errorf("pihole[%d]: password cannot be empty", i)
		}
		if node.Scheme != "http" && node.Scheme != "https" {
			return fmt.Errorf("pihole[%d]: scheme must be either 'http' or 'https'", i)
		}
	}

	return nil
}
