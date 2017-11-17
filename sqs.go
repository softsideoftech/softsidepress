package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/h2ik/go-sqs-poller/worker"
)
	func TestSqs() {
		// Make sure that AWS_SDK_LOAD_CONFIG=true is defined as an environment variable before running the application
		// like this

		// create the new client and return the url
		svc := worker.NewSQSClient()
		// set the queue url
		worker.QueueURL = "https://sqs.us-west-2.amazonaws.com/249869178481/softside-ses-q"
		// start the worker
		worker.Start(svc, worker.HandlerFunc(func (msg *sqs.Message) error {
			fmt.Println(aws.StringValue(msg.Body))
			return nil
		}))
	}