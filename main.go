package main

import (
	"checkerbox/internal/config"
	"log"
)

func main() {
	config, error := config.NewConfig("config/config.yml")
	if error != nil {
		log.Fatal(error.Error())
	}
	config.PrintConfig()
}
