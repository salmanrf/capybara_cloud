package utils

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	POSTGRES_URI    string
	API_PORT        string
	AUTH_JWT_SECRET string
	MAX_DEPLOY_FORM_SIZE int
	MAX_DEPLOY_BUNDLE_SIZE int
	DOCKER_REGISTRY string
	BASE_TEMP_PATH string
}

var app_cfg = Config{}

func LoadConfig(env_path string) (Config, error) {
	if err := godotenv.Load(env_path); err != nil {
		return Config{}, fmt.Errorf("unable to load env vars: %w", err)
	}

	max_form_size, err := strconv.Atoi(os.Getenv("MAX_DEPLOY_FORM_SIZE"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid MAX_DEPLOY_FORM_SIZE: %w", err)
	}
	max_bundle_size, err := strconv.Atoi(os.Getenv("MAX_DEPLOY_BUNDLE_SIZE"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid MAX_DEPLOY_BUNDLE_SIZE: %w", err)
	}
	cfg := Config{
		POSTGRES_URI:    os.Getenv("POSTGRES_URI"),
		API_PORT:        os.Getenv("API_PORT"),
		AUTH_JWT_SECRET: os.Getenv("AUTH_JWT_SECRET"),
		DOCKER_REGISTRY: os.Getenv("DOCKER_REGISTRY"),
		BASE_TEMP_PATH: os.Getenv("TEMP_ARTIFACT_PATH"),
		MAX_DEPLOY_FORM_SIZE: max_form_size,
		MAX_DEPLOY_BUNDLE_SIZE: max_bundle_size,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid configuration: %w", err)
	}

	app_cfg = cfg

	return app_cfg, nil
}

func (c *Config) Validate() error {
	var errs []error

	if c.POSTGRES_URI == "" {
		errs = append(errs, errors.New("POSTGRES_URI is required"))
	}
	if c.API_PORT == "" {
		errs = append(errs, errors.New("API_PORT is required"))
	}
	if c.AUTH_JWT_SECRET == "" {
		errs = append(errs, errors.New("AUTH_JWT_SECRET is required"))
	}
	if c.DOCKER_REGISTRY == "" {
		errs = append(errs, errors.New("DOCKER_REGISTRY is required"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func GetConfig() Config {
	return app_cfg
}

func SetConfig(newconf Config) {
	app_cfg = Config{
		POSTGRES_URI:    			  newconf.POSTGRES_URI,
		API_PORT:        				newconf.API_PORT,
		AUTH_JWT_SECRET: 				newconf.AUTH_JWT_SECRET,
		MAX_DEPLOY_FORM_SIZE: 	newconf.MAX_DEPLOY_FORM_SIZE,
		MAX_DEPLOY_BUNDLE_SIZE: newconf.MAX_DEPLOY_BUNDLE_SIZE,
	}
}