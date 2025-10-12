package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	// App Settings
	AppEnv    string `mapstructure:"APP_ENV"`
	LogLevel  string `mapstructure:"LOG_LEVEL"`
	LogFormat string `mapstructure:"LOG_LEVEL"`

	// Server settings
	ServerPort int    `mapstructure:"SERVER_PORT"`
	ServerHost string `mapstructure:"SERVER_HOST"`
	GRPCPort   int    `mapstructure:"GRPC_PORT"`

	// Database settings
	Database DatabaseConfig `mapstructure:",squash"`

	// JWT settings
	JWTSecret      string `mapstructure:"JWT_SECRET"`
	JWTExpiryHours int    `mapstructure:"JWT_EXPIRY_HOURS"`

	// Redis settings
	Redis RedisConfig `mapstructure:",squash"`

	// Rate limiting
	RateLimitRequestsPerMinute int `mapstructure:"RATE_LIMIT_REQUESTS_PER_MINUTE"`

	// CQRS settings
	CQRSAllowedOrigins []string `mapstructure:"CQRS_ALLOWED_ORIGINS"`
	CQRSAllowedMethods []string `mapstructure:"CQRS_ALLOWED_METHODS"`
	CQRSAllowedHeaders []string `mapstructure:"CQRS_ALLOWED_HEADERS"`

	// File upload settings
	MaxFileSizeMB int    `mapstructure:"MAX_FILE_SIZE_MB"`
	UploadPath    string `mapstructure:"UPLOAD_PATH"`

	// Monitoring settings
	MetricsEnabled bool `mapstructure:"METRICS_ENABLED"`
	MetricsPort    int  `mapstructure:"METRICS_PORT"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     int    `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSL_MODE"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     int    `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func (d DatabaseConfig) MigrationDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

func (r RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func Load() (*Config, error) {
	setDefaults()

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "-"))

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/hr-mangement")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Application defaults
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")

	// Server defaults
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("GRPC_PORT", 9090)

	// Database defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "hruser")
	viper.SetDefault("DB_PASSWORD", "hrpassword")
	viper.SetDefault("DB_NAME", "hrmanagement")
	viper.SetDefault("DB_SSL_MODE", "disable")

	// JWT defaults
	viper.SetDefault("JWT_SECRET", "39w0jcnsu9dns8end8e30dxk20snjw9enn9fnci39dn73839djd93")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)

	// Redis defaults
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_PASSWORD", "")

	// Rate limiting defaults
	viper.SetDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 100)

	// CORS defaults
	viper.SetDefault("CORS_ALLOWED_ORIGINS", []string{"*"})
	viper.SetDefault("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"})

	// File upload defaults
	viper.SetDefault("MAX_FILE_SIZE_MB", 10)
	viper.SetDefault("UPLOAD_PATH", "./uploads")

	// Monitoring defaults
	viper.SetDefault("METRICS_ENABLED", true)
	viper.SetDefault("METRICS_PORT", 8081)
}

func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}
	if c.GRPCPort <= 0 || c.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", c.GRPCPort)
	}
	if c.ServerPort <= 0 || c.ServerPort > 65535 {
		return fmt.Errorf("invalid server port: %d", c.ServerPort)
	}
	return nil
}

func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.AppEnv) == "development"
}

func (c *Config) IsProduction() bool {
	return strings.ToLower(c.AppEnv) == "Production"
}

func (c *Config) IsTest() bool {
	return strings.ToLower(c.AppEnv) == "test"
}
