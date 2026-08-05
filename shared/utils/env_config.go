package utils

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	store map[string]string
}

type Config interface {
	Load(env_path string, keys []string) error
	Get(key string) (string, bool)
	Set(key, val string)
}

func NewConfig() Config {
	store := make(map[string]string)
	
	return &config{store}
}

func (c *config) Get(key string) (string, bool) {
	val, ok := c.store[key]
	return val, ok
}

func (c *config) Set(key, val string) {
	c.store[key] = val
}

func (c *config) Load(env_path string, keys []string) error {
	if err := godotenv.Load(env_path); err != nil {
		return fmt.Errorf("unable to load env vars: %v", err)
	}

	for _, k := range keys {
		val := os.Getenv(k)
		if val == "" {
			return fmt.Errorf("env file doesn't contain key '%s'", k)
		}
		c.store[k] = val
	}
	
	return nil
}