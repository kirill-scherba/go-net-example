package main

import "os"

// Config is teonet config
type Config struct {
	Cookies string
}

// Default set config default values
func (c *Config) Default() []byte {
	return nil
}

// Dir return config directory
func (c *Config) Dir() string {
	return os.Getenv("HOME") + "/.config/teonet/teocli-termloop/"
}

// FileName return config directory
func (c *Config) FileName() string {
	return "termloop"
}

// Struct return config directory
func (c *Config) Struct() interface{} {
	return c
}
