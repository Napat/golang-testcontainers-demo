package config

import "time"

type Config struct {
    MySQL struct {
        Host     string
        Port     string
        User     string
        Password string
        Database string
    }
    PostgreSQL struct {
        Host     string
        Port     string
        User     string
        Password string
        Database string
    }
    Redis struct {
        Host     string
        Port     string
        Password string
        DB       int
        TTL      time.Duration
    }
    Kafka struct {
        Brokers []string
        Topic   string
    }
}
