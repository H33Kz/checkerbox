package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type AppSettings struct {
	Sites    int    `yaml:"sites"`
	Stages   int    `yaml:"stages"`
	Uiengine string `yaml:"uiengine"`
}

func NewAppSettings() *AppSettings {
	appSettingsFile, err := os.ReadFile("app.yml")
	if err != nil {
		log.Fatal(err.Error())
	}
	var appSettingsInstance AppSettings
	err = yaml.Unmarshal(appSettingsFile, &appSettingsInstance)
	if err != nil {
		log.Fatal(err.Error())
	}
	return &appSettingsInstance
}
