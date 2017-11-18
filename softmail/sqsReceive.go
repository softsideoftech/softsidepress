package softmail

import (
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/h2ik/go-sqs-poller/worker"
	"github.com/aws/aws-sdk-go/aws/session"
	"encoding/json"
	"strings"
)

type SqsMessage struct {
	Type, MessageId, TopicArn, Subject, Timestamp, Message, SignatureVersion, Signature, SigningCertURL, UnsubscribeURL string

}

type SesMessage struct {
	NotificationType string `json notificationType`
	Content string `json content`
	Mail Mail `json mail`
	Receipt map[string]interface{}  `json receipt`
}

type Mail struct {
	Timestamp string `json timestamp`
	Source string `json source`
	MessageId string `json messageId`
	HeadersTruncated bool `json headersTruncated`
	Destination []string `json destination`
	Headers []map[string]string `json headers`
	CommonHeaders map[string]interface{} `json commonHeaders`
}

func StartSqs() error {

	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})

	if (err != nil) {
		return err;
	}

	svc := sqs.New(sess)

	// set the queue url
	worker.QueueURL = "https://sqs.us-west-2.amazonaws.com/249869178481/softside-ses-q"
	// start the worker
	worker.Start(svc, worker.HandlerFunc(func(msg *sqs.Message) error {
		msgString := aws.StringValue(msg.Body)
		fmt.Println(msgString)
		msgString = strings.Replace(msgString, "\\\"","\"", 0)
		fmt.Println(msgString)
		var sqsMessage SqsMessage
		err := json.Unmarshal([]byte(msgString), &sqsMessage);
		if (err != nil) {
			return err;
		}


		//EMAIL\[\[(.+?)]].*FIRSTNAME\[\[(.+?)]]


		//var sesMessage map[string]interface{}
		var sesMessage SesMessage
		err = json.Unmarshal([]byte(sqsMessage.Message), &sesMessage);
		if (err != nil) {
			return err;
		}

		fmt.Println(sesMessage.Content)

		return nil
	}))

	return nil;
}
