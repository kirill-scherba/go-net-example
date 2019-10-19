package main

import (
	"os"

	"github.com/kirill-scherba/teonet-go/services/teoconf"
)

// Parameters is teocli-termloop teonet connection config structure
type Parameters struct {
	Cookies string
}

// Config contain methods and stucture with config data
type Config struct {
	Parameters
	configName string
}

// NewConfig initialize config
func newConfig(name string) (*teoconf.Teoconf, *Parameters) {
	conf := &Config{configName: name}
	return teoconf.New(conf), &conf.Parameters
}

// Default set config default values
func (c *Config) Default() []byte {
	return nil
}

// Value return config value interface
func (c *Config) Value() interface{} {
	return &c.Parameters
}

// Name return config name
func (c *Config) Name() string {
	return "termloop-" + c.configName
}

// Dir return config directory
func (c *Config) Dir() string {
	return os.Getenv("HOME") + "/.config/teonet/teocli-termloop/"
}
