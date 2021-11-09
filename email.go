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
func sendEmail(config CheckerConfig, content string) {
	cli := getSesClient(config)

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

	result, err := cli.SendEmail(input)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		os.Exit(1)
	}

	log.Printf("Send email result: %v", result)
}

// getSesClient creates new AWS session and returns AWS SES client.
func getSesClient(config CheckerConfig) *ses.SES {
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
