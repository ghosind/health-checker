package main

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// CheckResult is the result for check instance status.
type CheckResult struct {
	Message string
	Status  bool
}

// checkInstances gets the statuses of instances, and return the unreachable
// instance list.
func checkInstances(config *Config) []string {
	ch := make(chan CheckResult)
	expect := getInstanceCount(config)
	actual := 0
	messages := make([]string, 0, expect)

	if config.Instances != nil {
		for _, instance := range config.Instances {
			go checkInstance(instance, config, ch)
		}
	}

	if config.Groups != nil {
		for _, group := range config.Groups {
			go checkGroup(group, config, ch)
		}
	}

	for res := range ch {
		if !res.Status {
			messages = append(messages, res.Message)
		}

		actual++
		if expect == actual {
			close(ch)
		}
	}

	return messages
}

// getInstanceCount gets the count of instances including groups.
func getInstanceCount(config *Config) int {
	count := 0

	if config.Instances != nil {
		count += len(config.Instances)
	}

	if config.Groups != nil {
		count += len(config.Groups)
	}

	return count
}

// checkGroup gets the statuses of instances of the group.
func checkGroup(group InstanceGroup, config *Config, ch chan CheckResult) {
	instancesChan := make(chan CheckResult)
	num := 0
	failed := 0
	messages := make([]string, 0, len(group.Instances))

	var result CheckResult
	result.Status = true

	for _, instance := range group.Instances {
		go checkInstance(instance, config, instancesChan)
	}

	for res := range instancesChan {
		if !res.Status {
			messages = append(messages, res.Message)
			failed++
		}

		num++

		if num == len(group.Instances) {
			close(instancesChan)
		}
	}

	if group.Type != "all" && group.Type != "any" {
		// The default type is 'all'
		group.Type = "all"
	}

	switch group.Type {
	case "any":
		if failed > 0 {
			// The status of the group will be false when some instances were failed.
			result.Status = false
		}
	case "all":
		if failed == len(group.Instances) {
			// The status of the group will be false when all instances were failed.
			result.Status = false
		}
	}

	if !result.Status {
		result.Message = "Check group " + group.Name + " failed:\n\t" + strings.Join(messages, "\n\t")
	}

	ch <- result
}

// checkInstance sends request to specifc server and creates report if it has failed.
func checkInstance(
	instance Instance,
	config *Config,
	ch chan CheckResult,
) {
	var result CheckResult

	// Set to global uri if no specific uri in instance config.
	url := getRequestUrl(instance, config)

	client := http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	result.Status = true

	_, err := client.Get(url)
	if err != nil {
		log.Printf("Failed to open url %s: %v", url, err)
		result.Status = false
		result.Message = "Check instance " + instance.Addr + " failed (error: " + err.Error() + ")"
	}

	ch <- result
}

// getRequestUrl make destination url by instance or global config.
func getRequestUrl(instance Instance, config *Config) string {
	url := ""

	if instance.Scheme != "" {
		url = instance.Scheme + "://"
	} else if config.Scheme != "" {
		url = config.Scheme + "://"
	} else {
		url = "http://"
	}

	if instance.URI != "" {
		url += instance.Addr + instance.URI
	} else {
		url += instance.Addr + config.URI
	}

	return url
}
