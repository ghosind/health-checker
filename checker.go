package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
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
	Recipient string
}

// CheckResult The result for check instance status.
type CheckResult struct {
	Message string
	Status  bool
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: checker <config_file>")
		os.Exit(1)
	}

	configPath := os.Args[1]
	config := readConfig(configPath)

	messages := checkInstances(*config)
	if len(messages) > 0 {
		sendEmail(*config, strings.Join(messages, "\n\n"))
	}
}

// sendEmail Send check report to specified email.
func sendEmail(config CheckerConfig, content string) {
	session, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWS.Region),
		Credentials: credentials.NewStaticCredentials(
			config.AWS.ClientID,
			config.AWS.ClientSecret,
			"", // token is optional parameter.
		),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	service := ses.New(session)

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(config.Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("utf-8"),
					Data:    aws.String(content),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("utf-8"),
				Data:    aws.String("Check instance(s) failed"),
			},
		},
		Source: aws.String(config.AWS.Sender),
	}

	result, err := service.SendEmail(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Send email result: %s\n", result.String())
}

// readConfig Read checker configuration from file that passed by argument.
func readConfig(path string) *CheckerConfig {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var cfg CheckerConfig

	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return &cfg
}

// checkInstances Get the statuses of instances, and return the unreachable
// instance list.
func checkInstances(config CheckerConfig) []string {
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

// getInstanceCount Gets the count of instances including groups.
func getInstanceCount(config CheckerConfig) int {
	count := 0

	if config.Instances != nil {
		count += len(config.Instances)
	}

	if config.Groups != nil {
		count += len(config.Groups)
	}

	return count
}

// checkGroup Gets the statuses of instances of the group.
func checkGroup(group InstanceGroup, config CheckerConfig, ch chan CheckResult) {
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

	if result.Status == false {
		result.Message = "Check group " + group.Name + " failed:\n\t" + strings.Join(messages, "\n\t")
	}

	ch <- result
}

// checkInstance Get the status of specified instance.
func checkInstance(
	instance Instance,
	config CheckerConfig,
	ch chan CheckResult,
) {
	var result CheckResult

	// Use global uri if no special uri specified.
	var url string
	if instance.URI != "" {
		url = instance.Addr + instance.URI
	} else {
		url = instance.Addr + config.URI
	}

	client := http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	result.Status = true

	_, err := client.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		result.Status = false
		result.Message = "Check instance " + instance.Addr + " failed (error: " + err.Error() + ")"
	}

	ch <- result
}
