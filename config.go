package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
type Config struct {
	Instances []Instance
	Groups    []InstanceGroup
	AWS       AWSConfig
	URI       string
	Timeout   int
	Receivers *[]*string
	Receiver  *string
}

// loadConfig gets configurations from file.
func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	extension := getFileExtension(file.Name())
	switch extension {
	case "json":
		return parseJsonFile(file)
	default:
		return nil, fmt.Errorf("unsupported %s file format", extension)
	}
}

func parseJsonFile(file *os.File) (*Config, error) {
	cfg := new(Config)

	err := json.NewDecoder(file).Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func getFileExtension(name string) string {
	arr := strings.Split(name, ".")

	return strings.ToLower(arr[len(arr)-1])
}
