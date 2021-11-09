package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// AWSConfig The configuration of AWS client.
type AWSConfig struct {
	ClientID     string
	ClientSecret string
	Region       string
	Sender       string
}

// Instance The configurations of server instance.
type Instance struct {
	Addr string
	URI  string
}

// InstanceGroup The configurations of the group of instances.
type InstanceGroup struct {
	Instances []Instance
	Type      string
	Name      string
}

// CheckerConfig The Configrations of health checker.
type CheckerConfig struct {
	Instances []Instance
	Groups    []InstanceGroup
	AWS       AWSConfig
	URI       string
	Timeout   int
	Receivers *[]*string
	Receiver  *string
}

// loadConfig gets configurations from file.
func loadConfig(path string) *CheckerConfig {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Printf("Failed to open config file %s", err)
		os.Exit(1)
	}

	var cfg CheckerConfig

	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		log.Printf("Failed to decode config file: %v", err)
		os.Exit(1)
	}

	return &cfg
}
