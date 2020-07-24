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
}

// CheckerConfig The Configrations of health checker.
type CheckerConfig struct {
	Instances []Instance
	AWS       AWSConfig
	URI       string
	Timeout   int
	Recipient string
}

// CheckResult The result for check instance status.
type CheckResult struct {
	URL    string
	Status bool
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
		sendEmail(*config, strings.Join(messages, "\n"))
	}
}

// sendEmail Send check report to specified email.
func sendEmail(config CheckerConfig, content string) {
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.AWS.Region),
		Credentials: credentials.NewStaticCredentials(config.AWS.ClientID, config.AWS.ClientSecret, ""),
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

// checkInstances Get the statuses of instances, and return the unreachable instance list.
func checkInstances(config CheckerConfig) []string {
	ch := make(chan CheckResult)
	expect := len(config.Instances)
	actual := 0
	messages := make([]string, 0, expect)

	for _, instance := range config.Instances {
		go checkInstance(instance, config, ch)
	}

	for v := range ch {
		if !v.Status {
			msg := "Check instance " + v.URL + " failed."
			messages = append(messages, msg)
		}

		actual++
		if expect == actual {
			close(ch)
		}
	}

	return messages
}

// checkInstance Get the status of specified instance.
func checkInstance(instance Instance, config CheckerConfig, ch chan CheckResult) {
	var result CheckResult

	url := instance.Addr + config.URI

	client := http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	result.Status = true
	result.URL = url

	_, err := client.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		result.Status = false
	}

	ch <- result
}
