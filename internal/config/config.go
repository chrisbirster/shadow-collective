package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	HealthListen string        `json:"health_listen"`
	HTTP         []HTTPService `json:"http"`
	TCP          []TCPService  `json:"tcp"`
	DNS          DNSConfig     `json:"dns"`
}

type HTTPService struct {
	Name     string `json:"name"`
	Listen   string `json:"listen"`
	Upstream string `json:"upstream"`
}

type TCPService struct {
	Name     string `json:"name"`
	Listen   string `json:"listen"`
	Upstream string `json:"upstream"`
}

type DNSConfig struct {
	Enabled  bool   `json:"enabled"`
	Listen   string `json:"listen"`
	Upstream string `json:"upstream"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if cfg.HealthListen == "" {
		cfg.HealthListen = "0.0.0.0:9000"
	}
	if cfg.DNS.Listen == "" {
		cfg.DNS.Listen = "0.0.0.0:53"
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	seen := map[string]string{}
	register := func(name, listen string) error {
		name = strings.TrimSpace(name)
		listen = strings.TrimSpace(listen)
		if name == "" {
			return fmt.Errorf("service name cannot be empty")
		}
		if listen == "" {
			return fmt.Errorf("service %q listen address cannot be empty", name)
		}
		if existing, ok := seen[listen]; ok {
			return fmt.Errorf("service %q and %q both use listen address %q", existing, name, listen)
		}
		seen[listen] = name
		return nil
	}

	for _, svc := range c.HTTP {
		if err := register(svc.Name, svc.Listen); err != nil {
			return err
		}
		if !strings.HasPrefix(svc.Upstream, "http://") && !strings.HasPrefix(svc.Upstream, "https://") {
			return fmt.Errorf("http service %q upstream must start with http:// or https://", svc.Name)
		}
	}
	for _, svc := range c.TCP {
		if err := register(svc.Name, svc.Listen); err != nil {
			return err
		}
		if strings.TrimSpace(svc.Upstream) == "" {
			return fmt.Errorf("tcp service %q upstream cannot be empty", svc.Name)
		}
	}
	if c.DNS.Enabled && strings.TrimSpace(c.DNS.Upstream) == "" {
		return fmt.Errorf("dns.upstream is required when dns.enabled is true")
	}
	return nil
}
