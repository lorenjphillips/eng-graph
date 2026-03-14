package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DirName    = ".eng-graph"
	configFile = "config.yaml"
)

type Config struct {
	LLM           LLMConfig      `yaml:"llm,omitempty"`
	Sources       []SourceConfig `yaml:"sources"`
	ActiveProfile string         `yaml:"active_profile,omitempty"`
}

type LLMConfig struct {
	BaseURL   string `yaml:"base_url,omitempty"`
	APIKeyEnv string `yaml:"api_key_env,omitempty"`
	Model     string `yaml:"model,omitempty"`
}

type SourceConfig struct {
	Name   string         `yaml:"name"`
	Kind   string         `yaml:"kind"`
	Config map[string]any `yaml:"config"`
}

func DefaultConfig() *Config {
	return &Config{}
}

func Init(dir string) error {
	engDir := filepath.Join(dir, DirName)
	if err := os.MkdirAll(engDir, 0755); err != nil {
		return err
	}
	return Save(DefaultConfig(), dir)
}

func Load(dir string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, DirName, configFile))
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(cfg *Config, dir string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, DirName, configFile), data, 0644)
}

func (c *Config) FindSource(name string) *SourceConfig {
	for i := range c.Sources {
		if c.Sources[i].Name == name {
			return &c.Sources[i]
		}
	}
	return nil
}

func (c *Config) SourceByKind(kind string) []SourceConfig {
	var out []SourceConfig
	for _, s := range c.Sources {
		if s.Kind == kind {
			out = append(out, s)
		}
	}
	return out
}
