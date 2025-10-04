package config

import "time"

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type ServerConfig struct {
	Port    string        `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type DatabaseConfig struct {
}
