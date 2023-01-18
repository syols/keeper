package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Option function of a certain type
type Option func(s *Config)

// Config struct
type Config struct {
	Server ServerConfig `yaml:"server" json:"server"`
	Client ClientConfig `yaml:"client,omitempty" json:"client,omitempty"`
}

// ServerConfig Server config struct
type ServerConfig struct {
	Address                  Address `yaml:"address" json:"address"`
	DatabaseConnectionString *string `yaml:"database,omitempty" json:"database,omitempty"`
	Sign                     string  `yaml:"sign"`
	PrivateKey               *string `yaml:"pkey,omitempty" json:"pkey,omitempty"`
	Certificate              *string `yaml:"certificate,omitempty" json:"certificate,omitempty"`
}

// ClientConfig Client config struct
type ClientConfig struct {
	Address                  Address `yaml:"address" json:"address"`
	DatabaseConnectionString *string `yaml:"database,omitempty" json:"database,omitempty"`
	Certificate              *string `yaml:"certificate,omitempty" json:"certificate,omitempty"`
}

// Address struct
type Address struct {
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
}

// NewConfig creates config struct
func NewConfig(path string) (settings Config, err error) {
	err = settings.setDefault(path)
	if err != nil {
		return Config{}, err
	}
	return settings.set(NewEnvironmentVariables().Options()...), nil
}

func (s *Config) set(options ...Option) Config {
	for _, fn := range options {
		fn(s)
	}
	return *s
}

// Address create HTTP address
func (a *Address) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

func (s *Config) setDefault(configPath string) error {
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(file, s); err != nil {
		return err
	}
	return nil
}
