package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
)

const baseJSONConf = `{
  "hardware": [
    {
      "site": "1",
      "device": "genericuart",
      "address": "ttyUSB0",
      "baudrate": "9600"
    },
    {
      "site": "2",
      "device": "modbus",
      "address": "COM6",
      "baudrate": "4800"
    }
  ],
  "sequence": [
    {
      "step_label": "STAGE1"
    },
    {
      "step_label": "Do something",
      "device": "name",
      "timeout": "1000",
      "function": "name",
      "threshold": "name"
    },
    {
      "step_label": "STAGE2"
    },
    {
      "step_label": "Do other thing",
      "device": "name",
      "timeout": "3000",
      "function": "name",
      "threshold": "name"
    }
  ],
  "misc_settings": {
    "sites": 2,
    "stages": 2
  }
}`

func main() {
	hardware, sequence, misc := readConfigFile("config/config.json")

	fmt.Println("Hardware list:")
	for _, value := range hardware {
		fmt.Println("Device name: ", value.(map[string]interface{})["device"])
		fmt.Println("Baud Rate: ", value.(map[string]interface{})["baudrate"])
		fmt.Println("---------------------")
	}

	fmt.Println("\nSequence:")
	for _, value := range sequence {
		fmt.Println("Step label: ", value.(map[string]interface{})["step_label"])
		fmt.Println("Device: ", value.(map[string]interface{})["device"])
		fmt.Println("---------------------")
	}
	fmt.Println("\nMisc: ")
	fmt.Println("Sites: ", misc["sites"])
	fmt.Println("Stages: ", misc["stages"])
}

func readConfigFile(path string) ([]interface{}, []interface{}, map[string]interface{}) {
	data, err := os.ReadFile(path)
	if err != nil {
		match, err2 := regexp.MatchString(".*no such file or directory.*", err.Error())
		if err2 != nil {
			log.Fatal(err2.Error())
		}
		if match {
			fmt.Println(err.Error())
			fmt.Println("Loading placeholder config")
			data = []byte(baseJSONConf)
		} else {
			log.Fatal(err.Error())
		}
	}

	var unmarshaledJSON map[string]interface{}
	json.Unmarshal(data, &unmarshaledJSON)

	var hardware []interface{} = unmarshaledJSON["hardware"].([]interface{})
	var sequence []interface{} = unmarshaledJSON["sequence"].([]interface{})
	var misc map[string]interface{} = unmarshaledJSON["misc_settings"].(map[string]interface{})
	return hardware, sequence, misc
}
