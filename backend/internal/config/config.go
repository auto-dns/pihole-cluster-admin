package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Database      DatabaseConfig      `mapstructure:"database"`
	EncryptionKey string              `mapstructure:"encryption_key"`
	HealthService HealthServiceConfig `mapstructure:"health_service"`
	Log           LoggingConfig       `mapstructure:"log"`
	Server        ServerConfig        `mapstructure:"server"`
}

type DatabaseConfig struct {
	Path           string `mapstructure:"path"`
	MigrationsPath string `mapstructure:"migrations_path"`
}

type HealthServiceConfig struct {
	GracePeriodSeconds     int `mapstructure:"grace_period_seconds"`
	PollingIntervalSeconds int `mapstructure:"polling_interval_seconds"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

type ServerConfig struct {
	Port                     int                    `mapstructure:"port"`
	TLSEnabled               bool                   `mapstructure:"tls_enabled"`
	TLSCertFile              string                 `mapstructure:"tls_cert_file"`
	TLSKeyFile               string                 `mapstructure:"tls_key_file"`
	ReadHeaderTimeoutSeconds int                    `mapstructure:"read_header_timeout_seconds"`
	Session                  SessionConfig          `mapstructure:"session"`
	ServerSideEvents         ServerSideEventsConfig `mapstructure:"server_side_events"`
}

type SessionConfig struct {
	TTLHours            int    `mapstructure:"ttl_hours"`
	CookieName          string `mapstructure:"cookie_name"`
	CookiePath          string `mapstructure:"cookie_path"`
	SameSite            string `mapstructure:"same_site"`
	Secure              bool   `mapstructure:"secure"`
	AllowInsecureCookie bool   `mapstructure:"allow_insecure_cookie"`
}

type ServerSideEventsConfig struct {
	HeartbeatSeconds int `mapstructure:"heartbeat_seconds"`
}

func Load() (*Config, error) {
	if err := initConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

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
	viper.SetDefault("database.path", "/var/lib/pihole-cluster-admin/data.db")
	viper.SetDefault("database.migrations_path", "/migrations/server")
	viper.SetDefault("encryption_key", "")
	viper.SetDefault("health_service.grace_period_seconds", 10)
	viper.SetDefault("health_service.polling_interval_seconds", 5)
	viper.SetDefault("log.level", "INFO")
	viper.SetDefault("server.port", 8081)
	viper.SetDefault("server.tls_enabled", false)
	viper.SetDefault("server.tls_cert_file", "")
	viper.SetDefault("server.tls_key_file", "")
	viper.SetDefault("server.read_header_timeout_seconds", 10)
	viper.SetDefault("server.session.ttl_hours", 24)
	viper.SetDefault("server.session.cookie_name", "session_id")
	viper.SetDefault("server.session.cookie_path", "/")
	viper.SetDefault("server.session.same_site", "Strict")
	viper.SetDefault("server.session.secure", false)
	viper.SetDefault("server.session.allow_insecure_cookie", false)
	viper.SetDefault("server.server_side_events.heartbeat_seconds", 20)
	viper.SetDefault("encryption_key", "")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	return nil
}

// validate checks for config consistency.
func (c *Config) validate() error {
	// Database
	if strings.TrimSpace(c.Database.Path) == "" {
		return fmt.Errorf("database.path cannot be empty")
	}
	if strings.HasSuffix(c.Database.Path, "/") {
		return fmt.Errorf("database.path must be a file path, not a directory")
	}
	dir := filepath.Dir(c.Database.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("unable to create database directory %s: %w", dir, err)
	}
	if len(strings.TrimSpace(c.EncryptionKey)) < 32 {
		return fmt.Errorf("encryption_key must be at least 32 characters for AES-256")
	}

	if strings.TrimSpace(c.EncryptionKey) == "" {
		return fmt.Errorf("encryption_key is required for encrypting sensitive data")
	}

	validLevels := map[string]struct{}{
		"TRACE": {}, "DEBUG": {}, "INFO": {}, "WARN": {}, "ERROR": {}, "FATAL": {},
	}

	// Health Service
	if c.HealthService.GracePeriodSeconds < 0 {
		return fmt.Errorf("health_service.grace_period_seconds must be greater than 0")
	}
	if c.HealthService.PollingIntervalSeconds < 1 {
		return fmt.Errorf("health_service.polling_interval_seconds must be greater than 1")
	}

	// Logs
	if _, ok := validLevels[strings.ToUpper(c.Log.Level)]; !ok {
		return fmt.Errorf("log.level must be a valid log level, got: %s", c.Log.Level)
	}

	// Server
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be a valid TCP port")
	}
	if c.Server.TLSEnabled {
		if strings.TrimSpace(c.Server.TLSCertFile) == "" || strings.TrimSpace(c.Server.TLSKeyFile) == "" {
			return fmt.Errorf("TLS enabled but cert or key file not provided")
		}
	}
	if c.Server.ReadHeaderTimeoutSeconds < 10 {
		return fmt.Errorf("server.read_header_timeout_seconds may not be lower than 10")
	}

	// Server - Session
	if c.Server.Session.TTLHours <= 0 {
		return fmt.Errorf("server.session.ttl_hours must be > 0 (got %d)", c.Server.Session.TTLHours)
	}
	if strings.TrimSpace(c.Server.Session.CookieName) == "" {
		return fmt.Errorf("server.session.cookie_name cannot be empty")
	}
	if strings.TrimSpace(c.Server.Session.CookiePath) == "" {
		return fmt.Errorf("server.session.cookie_path cannot be empty")
	}
	switch strings.ToLower(c.Server.Session.SameSite) {
	case "strict", "lax", "none":
		// ok
	default:
		return fmt.Errorf("server.session.same_site must be one of Strict, Lax, or None (got %s)", c.Server.Session.SameSite)
	}
	if !c.Server.TLSEnabled && c.Server.Session.Secure && !c.Server.Session.AllowInsecureCookie {
		return fmt.Errorf("server.session.secure=true requires TLS or allow_insecure_cookie=true")
	}

	// If TLS is disabled and secure cookies are required, warn or fail
	if !c.Server.TLSEnabled && c.Server.Session.Secure && !c.Server.Session.AllowInsecureCookie {
		return fmt.Errorf("server.session.secure=true requires TLS or allow_insecure_cookie=true")
	}

	// Server - Server Side Events
	if c.Server.ServerSideEvents.HeartbeatSeconds < 5 {
		return fmt.Errorf("server.server_side_events.heartbeat_seconds must be greater than 5")
	}

	return nil
}
