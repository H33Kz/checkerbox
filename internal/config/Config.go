package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

const baseYAMLConf = `---
hardware:
- site: '1'
  device: genericuart
  address: ttyUSB0
  baudrate: '9600'
- site: '2'
  device: modbus
  address: COM6
  baudrate: '4800'
- site: '2'
  device: genericuart
  address: COM7
sequence:
- step_label: STAGE1
- step_label: Do something
  device: name
  timeout: '1000'
  function: name
  threshold: name
- step_label: STAGE2
- step_label: Do other thing
  device: name
  timeout: '3000'
  function: name
  threshold: name
misc_settings:
  sites: 2
  stages: 2`

type Config struct {
	path     string
	hardware []interface{}
	sequence []interface{}
	misc     map[string]interface{}
}

func NewConfig(path string) (*Config, error) {
	hardware, sequence, misc, error := readConfigFile(path)
	if error != nil {
		return nil, error
	}
	return &Config{
		path:     path,
		hardware: hardware,
		sequence: sequence,
		misc:     misc,
	}, nil
}

func readConfigFile(path string) ([]interface{}, []interface{}, map[string]interface{}, error) {
	jsonMatch, _ := regexp.MatchString("\\.json$", path)
	yamlMatch, _ := regexp.MatchString("\\.(yaml|yml)", path)
	// Open specified config file
	data, err := os.ReadFile(path)
	// If file doesn't open, check if error reads as missing file, if it does load sample config from const
	// If error is different stop execution
	if err != nil {
		match, err2 := regexp.MatchString(".*no such file or directory.*", err.Error())
		if err2 != nil {
			return nil, nil, nil, err2
		}
		if match {
			fmt.Println(err.Error())
			fmt.Println("Loading placeholder config")
			data = []byte(baseYAMLConf)
			yamlMatch = true
		} else {
			return nil, nil, nil, err
		}
	}

	if jsonMatch {
		var unmarshaledJSON map[string]interface{}
		json.Unmarshal(data, &unmarshaledJSON)

		hardware := unmarshaledJSON["hardware"].([]interface{})
		sequence := unmarshaledJSON["sequence"].([]interface{})
		misc := unmarshaledJSON["misc_settings"].(map[string]interface{})
		return hardware, sequence, misc, nil
	} else if yamlMatch {
		var unmarshaledYAML map[string]interface{}
		yaml.Unmarshal(data, &unmarshaledYAML)

		hardware := unmarshaledYAML["hardware"].([]interface{})
		sequence := unmarshaledYAML["sequence"].([]interface{})
		misc := unmarshaledYAML["misc_settings"].(map[string]interface{})
		return hardware, sequence, misc, nil
	} else {
		return nil, nil, nil, errors.New("file extension doesn't match any of supported types")
	}
}

func (c *Config) PrintConfig() {
	fmt.Println("Hardware list:")
	for _, value := range c.hardware {
		fmt.Println("Device name: ", value.(map[string]interface{})["device"])
		fmt.Println("Baud Rate: ", value.(map[string]interface{})["baudrate"])
		fmt.Println("---------------------")
	}

	fmt.Println("\nSequence:")
	for _, value := range c.sequence {
		fmt.Println("Step label: ", value.(map[string]interface{})["step_label"])
		fmt.Println("Device: ", value.(map[string]interface{})["device"])
		fmt.Println("---------------------")
	}
	fmt.Println("\nMisc: ")
	fmt.Println("Sites: ", c.misc["sites"])
	fmt.Println("Stages: ", c.misc["stages"])
}
