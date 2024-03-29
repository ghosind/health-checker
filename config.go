package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
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
	Addr   string
	URI    string
	Scheme string
}

// InstanceGroup The configurations of the group of instances.
type InstanceGroup struct {
	Instances []Instance
	Type      string
	Name      string
}

// Config is the configuration of the health checker.
type Config struct {
	Instances []Instance
	Groups    []InstanceGroup
	AWS       AWSConfig
	URI       string
	Scheme    string
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
		return parseJSONFile(file)
	case "yaml", "yml":
		return parseYAMLFile(file)
	default:
		return nil, fmt.Errorf("unsupported %s file format", extension)
	}
}

func parseJSONFile(file *os.File) (*Config, error) {
	cfg := new(Config)

	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseYAMLFile(file *os.File) (*Config, error) {
	cfg := new(Config)

	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getFileExtension(name string) string {
	arr := strings.Split(name, ".")

	return strings.ToLower(arr[len(arr)-1])
}
