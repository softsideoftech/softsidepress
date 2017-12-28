package softmail

import (
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/h2ik/go-sqs-poller/worker"
	"github.com/aws/aws-sdk-go/aws/session"
	"encoding/json"
	"strings"
	"net/mail"
	"bytes"
	"regexp"
)

type SqsMessage struct {
	Type, MessageId, TopicArn, Subject, Timestamp, Message, SignatureVersion, Signature, SigningCertURL, UnsubscribeURL string
}

type SesMessage struct {
	NotificationType string                 `json notificationType`
	Content          string                 `json content`
	Mail             Mail                   `json mail`
	Receipt          map[string]interface{} `json receipt`
}

type Mail struct {
	Timestamp        string                 `json timestamp`
	Source           string                 `json source`
	MessageId        string                 `json messageId`
	HeadersTruncated bool                   `json headersTruncated`
	Destination      []string               `json destination`
	Headers          []map[string]string    `json headers`
	CommonHeaders    map[string]interface{} `json commonHeaders`
}

var extractNameAndEmailRegex = regexp.MustCompile("EMAIL\\[\\[(.+?)]].*FIRSTNAME\\[\\[(.+?)]]");

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

		// Parse the SQS message
		msgString := aws.StringValue(msg.Body)
		var sqsMessage SqsMessage
		err := json.Unmarshal([]byte(msgString), &sqsMessage);
		if (err != nil) {
			return err;
		}

		//EMAIL\[\[(.+?)]].*FIRSTNAME\[\[(.+?)]]

		// Parse the SNS/SES message from the SQS message
		var sesMessage SesMessage
		err = json.Unmarshal([]byte(sqsMessage.Message), &sesMessage);
		if (err != nil) {
			return err;
		}

		// Parse the email content
		reader := strings.NewReader(sesMessage.Content)
		mailMsg, err := mail.ReadMessage(reader)
		if (err != nil) {
			return err;
		}

		// Obtain the email body
		buf := new(bytes.Buffer)
		buf.ReadFrom(mailMsg.Body)
		emailBody := buf.String()

		// Extract the email and firstname out of the body
		fmt.Println(emailBody)
		submatch := extractNameAndEmailRegex.FindStringSubmatch(emailBody)
		email := submatch[1]
		firstName := submatch[2]

		// Save to DB
		newListMember := &ListMember{
			Email:     email,
			FirstName: firstName,
		}
		err = SoftsideDB.Insert(newListMember)
		if err != nil {
			return err;
		}
		return nil
	}))

	return nil;
}