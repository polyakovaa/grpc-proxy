package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type EventServiceConfig struct {
	Server   ServerConfig  `yaml:"server"`
	Database DBConfig      `yaml:"database"`
	Logging  LoggingConfig `yaml:"logging"`
}

type ServerConfig struct {
	Port    string        `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DBName   string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`

	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

func LoadConfig(path string) (*EventServiceConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file for EventService: %w", err)
	}
	defer file.Close()

	var cfg EventServiceConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	return &cfg, nil
}
