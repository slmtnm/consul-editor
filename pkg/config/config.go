package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ConfigFilename string

// Context is struct that stores information about consul connection
type Context struct {
	Name    string `yaml:"name"`
	Token   string `yaml:"token"`
	Address string `yaml:"address"`
}

type Config struct {
	Contexts []Context `yaml:"contexts"`

	CurrentContextName string `yaml:"current-context"`
}

var DefaultConfig = &Config{
	Contexts: []Context{
		{
			Name:    "default",
			Token:   os.Getenv("CONSUL_HTTP_TOKEN"),
			Address: os.Getenv("CONSUL_HTTP_ADDR"),
		},
	},
	CurrentContextName: "default",
}

func readConfig(filename string) (*Config, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	yaml.Unmarshal(bytes, &config)

	return &config, nil
}

func findConfig() (string, error) {
	possiblePaths := []string{
		filepath.Join(os.Getenv("HOME"), ".consul-editor.yml"),
		filepath.Join(os.Getenv("HOME"), ".config", "consul-editor", "consul-editor.yml"),
		filepath.Join("/etc", "consul-editor", "consul-editor.yml"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New("could not find configuration file")
}

// New returns configuration that either read from config file or is default
func New() *Config {
	configPath, err := findConfig()
	var config *Config

	if err == nil {
		ConfigFilename = configPath
		config, err = readConfig(ConfigFilename)
	}

	// if config file not found or cannot be parsed
	if err != nil {
		// fmt.Printf("%v\nusing default configuration\n", err)
		config = DefaultConfig
	}

	return config
}

// CurrentContext returns pointer to current context
func (c *Config) CurrentContext() *Context {
	if c.CurrentContextName == "" {
		if len(c.Contexts) > 0 {
			return &c.Contexts[0]
		}
		panic("internal error: current context name isn't set and there are no contexts")
	}

	for _, context := range c.Contexts {
		if context.Name == c.CurrentContextName {
			return &context
		}
	}

	panic("internal error: current context name doesn't found in contexts list")
}
