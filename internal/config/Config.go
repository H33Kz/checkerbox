package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

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
`

type Config struct {
	path     string
	hardware []map[string]string
	sequence []map[string]string
}

func NewConfig(path string) (*Config, error) {
	hardware, sequence, error := readConfigFile(path)
	if error != nil {
		return nil, error
	}
	return &Config{
		path:     path,
		hardware: hardware,
		sequence: sequence,
	}, nil
}

func (c *Config) GetHardwareConfig() []map[string]string {
	return c.hardware
}

func (c *Config) GetSequenceConfig() []map[string]string {
	return c.sequence
}

func readConfigFile(path string) ([]map[string]string, []map[string]string, error) {
	// Check for json or yaml match
	jsonMatch, _ := regexp.MatchString("\\.json$", path)
	yamlMatch, _ := regexp.MatchString("\\.(yaml|yml)", path)
	// Open specified config file
	data, err := os.ReadFile(path)
	// If file doesn't open, check if error reads as missing file, if it does load sample config from const
	// Otherwise propagate error
	if err != nil {
		match, err2 := regexp.MatchString(".*no such file or directory.*", err.Error())
		if err2 != nil {
			return nil, nil, err2
		}
		if match {
			fmt.Println(err.Error())
			fmt.Println("Loading placeholder config")
			data = []byte(baseYAMLConf)
			yamlMatch = true
		} else {
			return nil, nil, err
		}
	}

	// Variables for holding initial umarshaled config
	var hardware []interface{}
	var sequence []interface{}

	// Unmarshal based on file extension or propagate error of unsupported file type
	if jsonMatch {
		var unmarshaledJSON map[string]interface{}
		json.Unmarshal(data, &unmarshaledJSON)

		hardware = unmarshaledJSON["hardware"].([]interface{})
		sequence = unmarshaledJSON["sequence"].([]interface{})
	} else if yamlMatch {
		var unmarshaledYAML map[string]interface{}
		yaml.Unmarshal(data, &unmarshaledYAML)

		hardware = unmarshaledYAML["hardware"].([]interface{})
		sequence = unmarshaledYAML["sequence"].([]interface{})
	} else {
		return nil, nil, errors.New("file extension doesn't match any of supported types")
	}

	// Remap unmarshaled config to be map of strings keyed by strings or arrays of this maps
	finalHardwareMap := make([]map[string]string, 0)
	for _, value := range hardware {
		intermediateMapNode := make(map[string]string)
		for key, mapVal := range value.(map[string]interface{}) {
			strKey := fmt.Sprintf("%v", key)
			strVal := fmt.Sprintf("%v", mapVal)
			strKey = strings.ToLower(strKey)
			intermediateMapNode[strKey] = strVal
		}
		finalHardwareMap = append(finalHardwareMap, intermediateMapNode)
	}

	finalSequenceMap := make([]map[string]string, 0)
	for _, value := range sequence {
		intermediateMapNode := make(map[string]string)
		for key, mapVal := range value.(map[string]interface{}) {
			strKey := fmt.Sprintf("%v", key)
			strVal := fmt.Sprintf("%v", mapVal)
			strKey = strings.ToLower(strKey)
			intermediateMapNode[strKey] = strVal
		}
		finalSequenceMap = append(finalSequenceMap, intermediateMapNode)
	}

	return finalHardwareMap, finalSequenceMap, nil
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
}
