package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
    Environment   string             `yaml:"environment"`
    MySQL        MySQLConfig        `yaml:"mysql"`
    PostgreSQL   PostgreSQLConfig   `yaml:"postgres"`
    Redis        RedisConfig        `yaml:"redis"`
    Kafka        KafkaConfig        `yaml:"kafka"`
    Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
}

type MySQLConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Database string
}

type PostgreSQLConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Database string
}

type RedisConfig struct {
    Host string
    Port string
}

type KafkaConfig struct {
    Brokers []string
    Topic   string
}

type ElasticsearchConfig struct {
    URL string
}

func Load(configPath string) (*Config, error) {
    data, err := os.ReadFile(configPath)
    if (err != nil) {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}


