package main

import (
	"errors"
	"github.com/adrg/xdg"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"log"
	"os"
	"path/filepath"
)

type AppConfig struct {
	GitlabUrl    string `koanf:"MERGEMATE_GITLAB_URL"`
	ProjectName  string `koanf:"MERGEMATE_PROJECT_NAME"`
	UserName     string `koanf:"MERGEMATE_USER_NAME"`
	BranchPrefix string `koanf:"MERGEMATE_BRANCH_PREFIX"`
	ApiToken     string `koanf:"MERGEMATE_API_TOKEN"`
}

const configFile = "/mergemate/mergemate_config.env"

var k = koanf.New(".")

func main() {
	config, err := parseConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	err = validateConfig(config)
	if err != nil {
		log.Fatalf("Invalid config: %v.", err)
	}
}

func validateConfig(config *AppConfig) error {
	if len(config.GitlabUrl) == 0 {
		return errors.New("please provide MERGEMATE_GITLAB_URL config entry")
	}
	if len(config.ProjectName) == 0 {
		return errors.New("please provide MERGEMATE_PROJECT_NAME config entry")
	}
	if len(config.UserName) == 0 {
		return errors.New("please provide MERGEMATE_USER_NAME config entry")
	}
	if len(config.BranchPrefix) == 0 {
		return errors.New("please provide MERGEMATE_BRANCH_PREFIX config entry")
	}
	if len(config.ApiToken) == 0 {
		return errors.New("please provide MERGEMATE_API_TOKEN config entry")
	}
	return nil
}

func parseConfig() (*AppConfig, error) {
	configFilePath := filepath.Join(xdg.ConfigHome, configFile)
	fileInfo, _ := os.Stat(configFilePath)
	if fileInfo != nil {
		// file exists, we can load config
		err := k.Load(file.Provider(configFilePath), dotenv.Parser())
		if err != nil {
			return nil, err
		}
	}
	err := k.Load(env.Provider("", ".", func(s string) string { return s }), nil)
	if err != nil {
		return nil, err
	}
	var out AppConfig
	err = k.Unmarshal("", &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
