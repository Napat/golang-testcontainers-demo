package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Environment string `yaml:"environment"`

	Server Server `yaml:"server"`

	MySQL struct {
		Host         string `yaml:"host"`
		Port         string `yaml:"port"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		Database     string `yaml:"database"`
		MaxOpenConns int    `yaml:"max_open_conns"`
		MaxIdleConns int    `yaml:"max_idle_conns"`
		MaxLifetime  int    `yaml:"max_lifetime"`  // in minutes
		MaxIdleTime  int    `yaml:"max_idle_time"` // in minutes
	} `yaml:"mysql"`

	PostgreSQL struct {
		Host         string `yaml:"host"`
		Port         string `yaml:"port"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		Database     string `yaml:"database"`
		MaxOpenConns int    `yaml:"max_open_conns"`
		MaxIdleConns int    `yaml:"max_idle_conns"`
		MaxLifetime  int    `yaml:"max_lifetime"`  // in minutes
		MaxIdleTime  int    `yaml:"max_idle_time"` // in minutes
	} `yaml:"postgresql"`

	Redis struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		PoolSize int    `yaml:"pool_size"`
	} `yaml:"redis"`

	Kafka struct {
		Brokers []string `yaml:"brokers"`
		Topic   string   `yaml:"topic"`
	} `yaml:"kafka"`

	Elasticsearch struct {
		URL string `yaml:"url"`
	} `yaml:"elasticsearch"`

	Tracing TracingConfig `yaml:"tracing"`
}

type Server struct {
	Port         string `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	IdleTimeout  int    `yaml:"idle_timeout"`
}

type TracingConfig struct {
	Enabled       bool    `yaml:"enabled"`
	ServiceName   string  `yaml:"serviceName"`
	CollectorURL  string  `yaml:"collectorUrl"`
	SamplingRatio float64 `yaml:"samplingRatio"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
