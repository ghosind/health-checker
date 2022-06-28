package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		printHelp()
		os.Exit(1)
	}

	configPath := os.Args[1]
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	messages := checkInstances(config)
	if len(messages) > 0 {
		sendEmail(config, strings.Join(messages, "\n\n"))
	}
}

// printHelp prints application version and usage.
func printHelp() {
	fmt.Fprintf(os.Stderr, `health-checker %s

Usages: health-checker <config_file>
`, Version)
}
