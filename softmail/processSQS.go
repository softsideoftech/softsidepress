package softmail

import (
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/h2ik/go-sqs-poller/worker"
	"github.com/aws/aws-sdk-go/aws/session"
	"encoding/json"
				"time"
	"log"
	"strings"
			"github.com/veqryn/go-email/email"
)

type SqsMessage struct {
	Type, MessageId, TopicArn, Subject, Timestamp, Message, SignatureVersion, Signature, SigningCertURL, UnsubscribeURL string
}

type SesMessage struct {
	NotificationType string                 `json notificationType`
	Content          string                 `json content`
	Mail             Mail                   `json mail`
	Bounce           Bounce                 `json bounce`
	Complaint        Complaint              `json complaint`
	Receipt          map[string]interface{} `json receipt`
}

type Bounce struct {
	BounceType        string          `json bounceType`
	BounceSubType     string          `json bounceSubType`
	BouncedRecipients []MailRecipient `json bouncedRecipients`
	Timestamp         string          `json timestamp`
	FeedbackId        string          `json feedbackId`
	RemoteMtaIp       string          `json remoteMtaIp`
}

type MailRecipient struct {
	EmailAddress string `json emailAddress`
}

type Complaint struct {
	UserAgent             string          `json userAgent`
	ComplainedRecipients  []MailRecipient `json complainedRecipients`
	ComplaintFeedbackType string          `json complaintFeedbackType`
	ArrivalDate           string          `json emailAddress`
	Timestamp             string          `json emailAddress`
	FeedbackId            string          `json emailAddress`
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

func StartSqs() error {

	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})

	if err != nil {
		return err
	}

	svc := sqs.New(sess)

	// set the queue url
	worker.QueueURL = "https://sqs.us-west-2.amazonaws.com/249869178481/softside-ses-q"
	// start the worker
	worker.Start(svc, worker.HandlerFunc(handleSqsMessage))

	return nil
}

type SQSHandlerError struct {
	message string
}

func (e SQSHandlerError) Error() string {
	return e.message
}

func handleSqsMessage(msg *sqs.Message) error {

	// Parse the SQS message
	msgString := aws.StringValue(msg.Body)
	var sqsMessage SqsMessage
	err := json.Unmarshal([]byte(msgString), &sqsMessage)
	if err != nil {
		return err
	}
	
	

	// Parse the SNS/SES message from the SQS message
	var sesMessage SesMessage
	err = json.Unmarshal([]byte(sqsMessage.Message), &sesMessage)
	if err != nil {
		return err
	}

	// todo: this is old dead code for reading incoming emails
	// Parse the email content
	if true {
		contentReader := strings.NewReader(sesMessage.Content)
		log.Printf("Received email content:\n\n%s", sesMessage.Content)
		mailMsg, err := email.ParseMessage(contentReader)
		if err != nil {
			return err
		}
		log.Printf("\n\nMail msg:\n\n%s", mailMsg)
		//
		//// Obtain the email body
		//buf := new(bytes.Buffer)
		//buf.ReadFrom(mailMsg.Body)
		//emailBody := buf.String()
		//
		//// debug statement
		//log.Printf("Received email:\n\n%s", emailBody)
	}

	// Retrieve the sent message id
	var sentEmail = &SentEmail{}
	err = SoftsideDB.Model(sentEmail).Where("third_party_id = ?", sesMessage.Mail.MessageId).Select()
	if err != nil {
		if IsPgSelectEmpty(err) {
			log.Printf("Couldn't find SendEmail with third_party_id: %s. ignoring.", sesMessage.Mail.MessageId)
			return nil
		} else {
			return err
		}
	}

	//VALUES ('sent'), ('delivered'), ('opened'), ('clicked'), ('hard_bounce'), ('soft_bounce'), ('complaint');

	// Handle Each type of notification
	switch sesMessage.NotificationType {
	case "Delivery":
		{
			err = SoftsideDB.Insert(&EmailAction{SentEmailId: sentEmail.Id, Action: "delivered",})
		}
	case "Bounce":
		{
			now := time.Now()
			if sesMessage.Bounce.BounceType == "Transient" {
				// If it's explicitly a Transient (ie 'soft') bounce, then record it but don't unsubscribe the person.
				err = SoftsideDB.Insert(&EmailAction{SentEmailId: sentEmail.Id, Action: "soft_bounce"})
			} else {
				// Assume it's a hard bounce unless explicitly stated otherwise and unsubscribe this person
				err = SoftsideDB.Insert(&EmailAction{SentEmailId: sentEmail.Id, Action: "hard_bounce"})
				if err == nil {
					result, err := SoftsideDB.Model(&ListMember{}).Set("unsubscribed = ?", &now).Where("id = ?", sentEmail.ListMemberId).Update()
					if result == nil || result.RowsAffected() != 1 {
						return SQSHandlerError{(fmt.Sprintf("Problem updating list member to unsubscribed with id: %d, err: %v", sentEmail.ListMemberId, err))}
					}
				}
			}
		}
	case "Complaint":
		{
			now := time.Now()
			err = SoftsideDB.Insert(&EmailAction{SentEmailId: sentEmail.Id, Action: "complaint",})

			// Unsubscribe people if they complain (i.e. mark as spam)
			if err == nil {
				result, err := SoftsideDB.Model(&ListMember{}).Set("unsubscribed = ?", &now).Where("id = ?", sentEmail.ListMemberId).Update()
				if result == nil || result.RowsAffected() != 1 {
					return SQSHandlerError{(fmt.Sprintf("Problem updating list member to unsubscribed with id: %d, err: %v", sentEmail.ListMemberId, err))}
				}
			}
		}
	default:
		log.Printf("Unexpected SES message:\n%s\n\n\n", sqsMessage.Message)
	}

	

	return err
}
