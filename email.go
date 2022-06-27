package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// sendEmail sends report to specific emails.
func sendEmail(config *Config, content string) {
	cli := getSesClient(config)
	receivers := getReceivers(config)

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: receivers,
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

	result, err := cli.SendEmail(input)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		os.Exit(1)
	}

	log.Printf("Send email result: %v", result)
}

// getReceivers gets receivers email from config.
func getReceivers(config *Config) []*string {
	receivers := make([]*string, 0)

	if config.Receivers != nil && len(*config.Receivers) > 0 {
		for _, receiver := range *config.Receivers {
			if receiver == nil || len(*receiver) == 0 {
				continue
			}

			receivers = append(receivers, receiver)
		}
	} else if config.Receiver != nil && len(*config.Receiver) > 0 {
		receivers = append(receivers, config.Receiver)
	}

	if len(receivers) == 0 {
		log.Printf("No receiver")
		os.Exit(1)
	}

	return receivers
}

// getSesClient creates new AWS session and returns AWS SES client.
func getSesClient(config *Config) *ses.SES {
	session, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWS.Region),
		Credentials: credentials.NewStaticCredentials(
			config.AWS.ClientID,
			config.AWS.ClientSecret,
			"", // token is optional parameter, just leave it as empty string
		),
	})
	if err != nil {
		log.Printf("Failed to create new AWS session: %v", err)
		os.Exit(1)
	}

	cli := ses.New(session)

	return cli
}
