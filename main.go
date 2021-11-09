package main

import (
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		log.Println("Usages: health-checker <config_file>")
		os.Exit(1)
	}

	configPath := os.Args[1]
	config := loadConfig(configPath)

	messages := checkInstances(*config)
	if len(messages) > 0 {
		sendEmail(*config, strings.Join(messages, "\n\n"))
	}
}
