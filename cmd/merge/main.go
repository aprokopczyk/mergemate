package main

import (
	"errors"
	"github.com/adrg/xdg"
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui"
	"github.com/aprokopczyk/mergemate/ui/context"
	"github.com/aprokopczyk/mergemate/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type AppConfig struct {
	GitlabUrl               string `koanf:"MERGEMATE_GITLAB_URL"`
	ProjectName             string `koanf:"MERGEMATE_PROJECT_NAME"`
	UserName                string `koanf:"MERGEMATE_USER_NAME"`
	SlbBranchPrefix         string `koanf:"MERGEMATE_SLB_BRANCH_PREFIX"`
	TargetBranchPrefixes    string `koanf:"MERGEMATE_TARGET_BRANCH_PREFIXES"`
	ApiToken                string `koanf:"MERGEMATE_API_TOKEN"`
	MergeJobIntervalSeconds int    `koanf:"MERGEMATE_MERGE_JOB_INTERVAL_SECONDS"`
	FavouriteBranches       string `koanf:"MERGEMATE_FAVORITE_BRANCHES"`
}

const configFile = "/mergemate/mergemate_config.env"
const mergeMateDir = "/mergemate"
const logFile = "/debug.log"

var k = koanf.New(".")

func main() {
	loggerFile, err := configureLogFile()
	if err != nil {
		log.Fatalf("Error when configuring logfile: %v", err)
	}

	log.Println("Started application")

	config, err := parseConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	err = validateConfig(config)
	if err != nil {
		log.Fatalf("Invalid config: %v.", err)
	}
	client := gitlab.New(config.GitlabUrl, config.ProjectName, config.UserName, config.ApiToken)
	var appContext = context.AppContext{
		Styles:               styles.NewStyles(),
		GitlabClient:         client,
		MergeJobInterval:     config.MergeJobIntervalSeconds,
		UserBranchPrefix:     config.SlbBranchPrefix,
		TargetBranchPrefixes: strings.Split(config.TargetBranchPrefixes, ","),
		FavouriteBranches:    strings.Split(config.FavouriteBranches, ","),
		TablePageSize:        styles.MinTablePageSize,
	}
	p := tea.NewProgram(ui.New(&appContext), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	err = loggerFile.Close()
	if err != nil {
		log.Fatalf("Error closing logfile: %v", err)
	}
}

func configureLogFile() (*os.File, error) {
	logDir := filepath.Join(xdg.StateHome, mergeMateDir)
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	logFilePath := filepath.Join(logDir, logFile)
	return tea.LogToFile(logFilePath, "debug")
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
	if len(config.SlbBranchPrefix) == 0 {
		return errors.New("please provide MERGEMATE_SLB_BRANCH_PREFIX config entry")
	}
	if len(config.ApiToken) == 0 {
		return errors.New("please provide MERGEMATE_API_TOKEN config entry")
	}
	if config.MergeJobIntervalSeconds <= 0 {
		return errors.New("MERGEMATE_MERGE_JOB_INTERVAL_SECONDS has to be bigger than 0")
	}
	return nil
}

func parseConfig() (*AppConfig, error) {
	configFilePath := filepath.Join(xdg.ConfigHome, configFile)

	// init default values
	err := k.Load(confmap.Provider(map[string]interface{}{
		"MERGEMATE_MERGE_JOB_INTERVAL_SECONDS": 60,
	}, ""), nil)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if err != nil {
		return nil, err
	}

	// file exists, we can load config
	err = k.Load(file.Provider(configFilePath), dotenv.Parser())
	if err != nil {
		return nil, err
	}

	err = k.Load(env.Provider("", ".", func(s string) string { return s }), nil)
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
