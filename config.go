package main

import "github.com/presnalex/go-micro/v3/service"

func newConfig(name, version string) *Config {
	return &Config{
		App:    &AppConfig{},
		Core:   &service.CoreConfig{},
		Server: &service.ServerConfig{Name: name, Version: version},
		Consul: &service.ConsulConfig{
			NamespacePath: "/go-micro-layouts",
			AppPath:       "pub-sub-layout",
		},
		Vault:           &service.VaultConfig{},
		Metric:          &service.MetricConfig{},
		Broker:          &service.BrokerConfig{},
		PostgresPrimary: &service.PostgresConfig{},
	}
}

type Config struct {
	App             *AppConfig              `json:"app"`
	Core            *service.CoreConfig     `json:"core"`
	Server          *service.ServerConfig   `json:"server"`
	Consul          *service.ConsulConfig   `json:"consul"`
	Broker          *service.BrokerConfig   `json:"broker"`
	Vault           *service.VaultConfig    `json:"vault"`
	Metric          *service.MetricConfig   `json:"metric"`
	PostgresPrimary *service.PostgresConfig `json:"postgres_primary"`
}

type AppConfig struct {
	Topics struct {
		AnimalAdd   string `json:"animalAdd"`
		AnimalAddRs string `json:"animalAddRs"`
	} `json:"topics"`
}
