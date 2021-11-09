package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		printHelp()
		os.Exit(1)
	}

	configPath := os.Args[1]
	config := loadConfig(configPath)

	messages := checkInstances(*config)
	if len(messages) > 0 {
		sendEmail(*config, strings.Join(messages, "\n\n"))
	}
}

func printHelp() {
	fmt.Fprintf(os.Stderr, `health-checker %s

Usages: health-checker <config_file>
`, Version)
}
