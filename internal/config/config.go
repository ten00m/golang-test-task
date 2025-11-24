package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPServer HTTPServerConfig `yaml:"http_server"`
	PostgreSQL PostgreSQLConfig `yaml:"psql_info"`
}

type HTTPServerConfig struct {
	Address     string        `yaml:"address" env:"HTTP_ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"6s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type PostgreSQLConfig struct {
	Host     string `yaml:"host" env:"PSQL_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"PSQL_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"PSQL_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"PSQL_PASSWORD"`
	Database string `yaml:"database" env:"PSQL_DATABASE"`
}

// LoadConfig loads configuration from a YAML file specified by flag or environment variable
func LoadConfig() *Config {
	var configPath string

	// Check if config path is provided via command line flag
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	// If not provided via flag, check environment variable
	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	// If still not provided, use default
	if configPath == "" {
		configPath = "config.yaml"
	}

	var cfg Config

	// Load from YAML file
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config from file %s: %v", configPath, err)
	}

	// Override with environment variables
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read config from environment: %v", err)
	}

	return &cfg
}
