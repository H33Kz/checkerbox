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
	hardware []map[string]string
	sequence []map[string]string
	misc     map[string]string
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

func (c *Config) GetHardwareConfig() []map[string]string {
	return c.hardware
}

func (c *Config) GetSequenceConfig() []map[string]string {
	return c.sequence
}

func (c *Config) GetMiscConfig() map[string]string {
	return c.misc
}

func readConfigFile(path string) ([]map[string]string, []map[string]string, map[string]string, error) {
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

	var hardware []interface{}
	var sequence []interface{}
	var misc map[string]interface{}

	if jsonMatch {
		var unmarshaledJSON map[string]interface{}
		json.Unmarshal(data, &unmarshaledJSON)

		hardware = unmarshaledJSON["hardware"].([]interface{})
		sequence = unmarshaledJSON["sequence"].([]interface{})
		misc = unmarshaledJSON["misc_settings"].(map[string]interface{})
	} else if yamlMatch {
		var unmarshaledYAML map[string]interface{}
		yaml.Unmarshal(data, &unmarshaledYAML)

		hardware = unmarshaledYAML["hardware"].([]interface{})
		sequence = unmarshaledYAML["sequence"].([]interface{})
		misc = unmarshaledYAML["misc_settings"].(map[string]interface{})
	} else {
		return nil, nil, nil, errors.New("file extension doesn't match any of supported types")
	}

	finalHardwareMap := make([]map[string]string, 1)
	for _, value := range hardware {
		intermediateMapNode := make(map[string]string)
		for key, mapVal := range value.(map[string]interface{}) {
			strKey := fmt.Sprintf("%v", key)
			strVal := fmt.Sprintf("%v", mapVal)
			intermediateMapNode[strKey] = strVal
		}
		finalHardwareMap = append(finalHardwareMap, intermediateMapNode)
	}

	finalSequenceMap := make([]map[string]string, 1)
	for _, value := range sequence {
		intermediateMapNode := make(map[string]string)
		for key, mapVal := range value.(map[string]interface{}) {
			strKey := fmt.Sprintf("%v", key)
			strVal := fmt.Sprintf("%v", mapVal)
			intermediateMapNode[strKey] = strVal
		}
		finalSequenceMap = append(finalSequenceMap, intermediateMapNode)
	}

	finalMiscMap := make(map[string]string)
	for key, value := range misc {
		strKey := fmt.Sprintf("%v", key)
		strVal := fmt.Sprintf("%v", value)
		finalMiscMap[strKey] = strVal
	}
	return finalHardwareMap, finalSequenceMap, finalMiscMap, nil
}

func (c *Config) PrintConfig() {
	fmt.Println("Hardware list:")
	for _, value := range c.hardware {
		fmt.Println("Device name: ", value["device"])
		fmt.Println("Baud Rate: ", value["baudrate"])
		fmt.Println("---------------------")
	}

	fmt.Println("\nSequence:")
	for _, value := range c.sequence {
		fmt.Println("Step label: ", value["step_label"])
		fmt.Println("Device: ", value["device"])
		fmt.Println("---------------------")
	}
	fmt.Println("\nMisc: ")
	fmt.Println("Sites: ", c.misc["sites"])
	fmt.Println("Stages: ", c.misc["stages"])
}
