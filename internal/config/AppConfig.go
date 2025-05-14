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

type DeviceSettings struct {
	Site       int            `yaml:"site"`
	DeviceName string         `yaml:"device_name"`
	Settings   map[string]any `yaml:"settings"`
}

type SequenceStepSettings struct {
	StepLabel    string         `yaml:"step_label"`
	Retry        int            `yaml:"retry"`
	Device       string         `yaml:"device"`
	Timeout      int            `yaml:"timeout"`
	StepSettings map[string]any `yaml:"stepsettings"`
}

type Config struct {
	Hardware []DeviceSettings       `yaml:"hardware"`
	Sequence []SequenceStepSettings `yaml:"sequence"`
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

func NewConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var unmarshalledData Config
	err = yaml.Unmarshal(file, &unmarshalledData)
	if err != nil {
		return nil, err
	}
	return &unmarshalledData, nil
}

func (c *Config) GetSequenceConfig() []SequenceStepSettings {
	return c.Sequence
}

func (c *Config) GetHardwareConfig() []DeviceSettings {
	return c.Hardware
}
