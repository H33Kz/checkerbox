package main

import (
	"checkerbox/internal/config"
	"fmt"
	"log"
)

func main() {
	hardware, sequence, misc, error := config.ReadConfigFile("config/config.yaml")
	if error != nil {
		log.Fatal(error)
	} else {
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
}
