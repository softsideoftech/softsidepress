package softmail

import (
	"io/ioutil"
	"fmt"
	"gopkg.in/russross/blackfriday.v2"
	"crypto/md5"
	"hash/fnv"
	"strings"
	"regexp"
	"encoding/xml"
	htmlTemplate "html/template"
	"bytes"
	"github.com/jaytaylor/html2text"
	"github.com/veqryn/go-email/email"
	"github.com/sourcegraph/go-ses"
	"log"
	"net/smtp"
	"os"
	"time"
	"net/mail"
)

const emailSuffixMdFile = "emailSuffix.md"
const forwardedEmailSuffixMdFile = "forwardedEmailSuffix.md"
const trackingPixelPath = "/pixie/"

var awsSmtpUsername string = os.Getenv("AWS_SES_SMTP_USER")
var awsSmtpPassword string = os.Getenv("AWS_SES_SMTP_PASSWORD")

var linkRegExString = "\\((https://%s)(.+?)\\)";

var extractSentEmailIdFromUrlEndSlash = regexp.MustCompile("/.*/(.*)")

type EmailTemplateParams struct {
	FirstName          string
	SentEmailId        string
	SiteDomain         string
	SiteOwnerFirstName string
	DestinationUrl     string
	PageTitle          string
	Params             PerRequestParams
}

type LoginEmailTemplateParams struct {
	DestinationUrl string
	PageTitle      string
}

type SendEmailOpts struct {
	UseSuffix      bool
	Login          bool
	DestinationUrl string
	HostName       string
	PageTitle      string
	TemplateParams PerRequestParams
}

type NoSuchListMember struct {
	msg string
}

func (err NoSuchListMember) Error() string {
	return err.msg
}

func emailTemplateToId(subject string, body []byte, recipient string) EmailTemplateId {
	hash := md5.New()
	hash.Write(body)
	hash.Write([]byte(subject))
	hash.Write([]byte(recipient))

	md5Sum := hash.Sum(nil)
	hash64 := fnv.New64()
	hash64.Write(md5Sum)
	return EmailTemplateId(int64(hash64.Sum64())) // make it signed to conform with the Postgres "bigint" type
}

type SendEmailResponse struct {
	MessageId string `xml:"SendEmailResult>MessageId"`
	RequestId string `xml:"ResponseMetadata>RequestId"`
}

func decodeSendMailIdFromUriEnd(path string) SentEmailId {
	submatch := extractSentEmailIdFromUrlEndSlash.FindStringSubmatch(path)
	if submatch == nil {
		return 0
	}
	sentEmailId, err := DecodeId(submatch[1])
	if err != nil {
		log.Printf("Problem parsing SentEmailId from url: %s, error: %v", path, err)
		return 0
	}
	return sentEmailId
}

func FindPartType(msg *email.Message, contentTypePrefix string) ([]*email.Message, error) {
	buf := make([]*email.Message, 0, 1)
	if msg.HasParts() {
		for _, part := range msg.Parts {
			contentType, _, err := part.Header.ContentType()
			if err != nil {
				return buf, err
			}
			if strings.HasPrefix(contentType, contentTypePrefix) {
				buf = append(buf, part)
			} else {
				subBuf, err := FindPartType(part, contentTypePrefix)
				buf = append(buf, subBuf...)
				if err != nil {
					return buf, err
				}
			}
		}
	}
	return buf, nil
}

func (ctx RequestContext) ForwardEmail(sender string, recipient string, msg *email.Message) {

	var textEmailBody string
	var htmlEmailBody string

	htmlMessages, err := FindPartType(msg, "text/html")
	if err != nil {
		log.Printf("ERROR finding content type 'text/html': %v\n", err)
	}
	if len(htmlMessages) > 0 {
		htmlEmailBodyBytes := htmlMessages[len(htmlMessages)-1].Body
		htmlEmailBody = string(htmlEmailBodyBytes)
	}

	// Obtain the body so we could save it in the DB
	textMessages, err := FindPartType(msg, "text/plain")
	if err != nil {
		log.Printf("ERROR finding content type 'text/plain': %v\n", err)
	}
	if len(textMessages) > 0 {
		textEmailBodyBytes := textMessages[len(textMessages)-1].Body
		textEmailBody = string(textEmailBodyBytes)
	} else {
		textEmailBody = ""
	}

	subject := msg.Header.Get("Subject")

	// Obtain the emailTemplateId
	emailTemplateId := obtainEmailTemplateId(subject, textEmailBody, recipient)

	var listMember *ListMember
	recipients := append(append(msg.Header.To(), msg.Header.Cc()...), msg.Header.Bcc()...)
	for _, fullEmail := range recipients {
		address, err := mail.ParseAddress(fullEmail)
		if fullEmail != "" && address.Address == recipient {
			var exists bool
			listMember, exists, err = ctx.GetListMemberByEmail(recipient)
			if err != nil {
				log.Printf("ERROR retrieving list member while forwaring email: %s, %v\n", fullEmail, err)
			}
			if !exists {
				// If didn't find this list member, then create a new 'unsubscribed' list member.
				listMember = createListMember(address, false)
			}
		}
	}

	if err != nil {
		log.Printf("ERROR retrieving list member by email during email forwarding: %v\n", err)
	}

	// Create a SentEmail record
	sentEmail := &SentEmail{
		EmailTemplateId: emailTemplateId,
		ListMemberId:    listMember.Id,
	}
	err = SoftsideDB.Insert(sentEmail)
	if err != nil {
		log.Printf("ERROR inserting sendEmail into DB: %v\n", err)
	}
	// Base64 encode the SentEmail id
	encodedSentEmailId := EncodeId(sentEmail.Id)

	// Load the tracking prefix
	renderedPrefix := obtainTrackingPrefix(encodedSentEmailId)

	// Append the to only the html email
	htmlEmailBody = renderedPrefix + htmlEmailBody

	// Place the html bytes back into the message
	if len(htmlMessages) > 0 {
		htmlMessages[len(htmlMessages)-1].Body = []byte(htmlEmailBody)
	}

	// Re-set the Received header to make sure the recipient is the only thing there
	msg.Header.Set("Received", fmt.Sprintf("by softsideoftech.com with SMTP for <%s>", recipient))

	// Obtain the message bytes
	msgBytes, err := msg.Bytes()
	if err != nil {
		log.Panicf("ERROR retrieving emai message bytes: %v", err)
	}

	// Actually send the email
	auth := smtp.PlainAuth("", awsSmtpUsername, awsSmtpPassword, "email-smtp.us-west-2.amazonaws.com")
	awsResponse, err := SendMail("email-smtp.us-west-2.amazonaws.com:587", auth, sender, []string{recipient}, msgBytes)
	log.Printf("\nAWS SMTP RESPONSE:%s,%v\n:", awsResponse, err);
	processSentEmail(err, htmlEmailBody, textEmailBody, awsResponse[3:], sentEmail, listMember)
}

func createListMember(address *mail.Address, subscribed bool) *ListMember {
	now := time.Now()
	var listMember ListMember
	if !subscribed {
		listMember.Unsubscribed = &now
	}

	var firstName = strings.Split(address.Name, " ")[0]
	if firstName == "" {
		// If the email didn't contain a name, then try to extract the first 
		// name from the email address using a couple common name separators
		firstName = strings.Split(address.Address, "@")[0]
		firstName = strings.Split(firstName, ".")[0]
		firstName = strings.Split(firstName, "-")[0]
		firstName = strings.Split(firstName, "+")[0]
	}
	listMember.FirstName = strings.Title(strings.ToLower(firstName))
	listMember.Email = address.Address
	err := SoftsideDB.Insert(&listMember)
	if err != nil {
		log.Printf("ERROR inserting ListMember while forwarding email: %s, %v\n", address.Address, err)
	} else {
		log.Printf("Added new ListMember: %v\n", address)
	}
	return &listMember
}

// todo: not using this right now
func (ctx *RequestContext) obtainTrackingSuffix(encodedSentEmailId string) string {
	suffixEmailBodyBytes, err := ioutil.ReadFile(ctx.GetFilePath("/emails/" + forwardedEmailSuffixMdFile))
	if err != nil {
		panic(err)
	}
	suffixEmailBody := string(suffixEmailBodyBytes)
	// Parse and render the suffix template
	template, err := htmlTemplate.New(forwardedEmailSuffixMdFile).Parse(suffixEmailBody)
	if err != nil {
		panic(err)
	}
	buffer := &bytes.Buffer{}
	template.Execute(buffer, &EmailTemplateParams{
		SentEmailId:        encodedSentEmailId,
		SiteDomain:         siteDomain,
		SiteOwnerFirstName: ownerFirstName,
	})
	renderedSuffix := buffer.String()
	return renderedSuffix
}

func obtainTrackingPrefix(encodedSentEmailId string) string {
	// Parse and render the suffix template
	return fmt.Sprintf("<img src=\"https://softsideoftech.com/pixie/%s.png\"/>", encodedSentEmailId);
}

func (ctx *RequestContext) SendTemplatedEmail(subject string, templateFileName string, memberEmailOrGroupName string, opts SendEmailOpts) []ListMember {

	log.Printf("In SendTemplatedEmail(...) with opts: %v: ", opts)

	// Load the template file
	markdownEmailBodyBytes, err := ioutil.ReadFile(ctx.GetFilePath(templateFileName))
	markdownEmailBody := string(markdownEmailBodyBytes)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		panic(err)
	}

	emailTemplateId := obtainEmailTemplateId(subject, markdownEmailBody, "")

	// Add the SendEmailId template parameter to all internal links
	var linkRegex = regexp.MustCompile(fmt.Sprintf(linkRegExString, siteDomain))
	markdownEmailBody = linkRegex.ReplaceAllString(markdownEmailBody, "($1$2-{{.SentEmailId}})")

	if opts.UseSuffix {
		// Load the suffix template and append it to the markdown template
		suffixEmailBodyBytes, err := ioutil.ReadFile(ctx.GetFilePath("/emails/" + emailSuffixMdFile))
		suffixEmailBody := string(suffixEmailBodyBytes)
		if err != nil {
			panic(err)
		}
		markdownEmailBody += suffixEmailBody
	} else {
		trackingPrefix := obtainTrackingPrefix("{{.SentEmailId}}")
		markdownEmailBody = trackingPrefix + markdownEmailBody
	}

	// Turn the markdown into HTML
	htmlEmailTemplateString := string(blackfriday.Run([]byte(markdownEmailBody)))

	// Create the HTML template (must be HTML and not TEXT to escape user supplied values such as FirstName)
	parsedEmailTempalte, err := htmlTemplate.New(templateFileName).Parse(htmlEmailTemplateString)

	// Load the email list
	listMembers := ctx.obtainListMembers(memberEmailOrGroupName, opts)

	fmt.Printf("Sending email to listMembers: %v\n", listMembers)

	// For each member
	// 		Create a SentEmail record
	//		Templatize the name and links. Include the email_sent id and member_id
	//		Generate an html version
	//		Send the email
	//		Record the fact the email was sent
	var fromEmail = fmt.Sprintf("%s %s<%s>", ownerFirstName, ownerLastName, ownerEmail)
	for _, listMember := range listMembers {
		ctx.sendEmailToListMember(emailTemplateId, listMember, parsedEmailTempalte, fromEmail, subject, opts)
	}

	return listMembers
}

func (ctx *RequestContext) obtainListMembers(memberEmailOrGroupName string, opts SendEmailOpts) []ListMember {
	var err error = nil
	var listMembers []ListMember
	if memberEmailOrGroupName == "all" {
		// Select all members where unsubscribed is nil (ie, they never explicitly unsubscribed)
		err = SoftsideDB.Model(&listMembers).Where("unsubscribed IS NULL", nil).Select()
	} else if strings.Contains(memberEmailOrGroupName, "@") {
		address, _ := mail.ParseAddress(memberEmailOrGroupName)

		// If an email was supplied, then select that member.
		listMember, found, err := ctx.GetListMemberByEmail(address.Address)
		if err != nil {
			panic(err)
		}

		if !found {
			// Add the email to list_members if it doesn't already exist.
			log.Printf("Couldn't find a member with the email: %s. Creating one now.", address.Address)

			// Only subscribe the list member if they are logging in (that means they're interested)
			listMember = createListMember(address, opts.Login)
		}
		listMembers = append(listMembers, *listMember)
	} else {
		_, err = SoftsideDB.Query(&listMembers, `
		select l.* from list_members l, member_groups g 
		where l.id = g.list_member_id and g.name = ? AND l.unsubscribed IS NULL`, memberEmailOrGroupName)
	}

	// Not ok to have an error here. Just do a hard failure.
	if err != nil {
		panic(err)
	}

	return listMembers
}

func obtainEmailTemplateId(subject string, emailBody string, recipient string) (EmailTemplateId) {
	// Save the email template in the DB if it doesn't exist
	emailTemplateId := emailTemplateToId(subject, []byte(emailBody), recipient)
	emailTemplate := EmailTemplate{Id: emailTemplateId}
	err := SoftsideDB.Select(&emailTemplate)
	if err != nil {
		if IsPgSelectEmpty(err) {
			emailTemplate.Subject = subject
			emailTemplate.Body = emailBody
			err = SoftsideDB.Insert(&emailTemplate)
			if err != nil {
				panic(err)
			}
		} else {
			log.Panicf("Problem obtaining email template ID: %v", err)
		}
	}
	return emailTemplateId
}

func (ctx RequestContext) sendEmailToListMember(emailTemplateId EmailTemplateId, listMember ListMember, parsedEmailTempalte *htmlTemplate.Template, fromEmail string, subject string, opts SendEmailOpts) {
	// Create a SentEmail record
	sentEmail := &SentEmail{
		EmailTemplateId: emailTemplateId,
		ListMemberId:    listMember.Id,
	}
	err := ctx.DB.Insert(sentEmail)
	if err != nil {
		panic(err)
	}

	// If we're logging in, then create a unique URL for the link.
	var destinationUrl string
	if opts.Login {
		destinationUrl, err = ctx.TryToCreateShortTrackedUrl(opts.DestinationUrl, opts.HostName, sentEmail.Id, opts.Login)
		if err != nil {
			panic(fmt.Sprintf("ERROR obtaining TrackedUrl: %v", err))
		}
	} else {
		destinationUrl = opts.DestinationUrl
	}

	// Base64 encode the SentEmail id
	encodedSentEmailId := EncodeId(sentEmail.Id)

	// Render the HTML template
	buffer := &bytes.Buffer{}
	err = parsedEmailTempalte.Execute(buffer, &EmailTemplateParams{
		SentEmailId:        encodedSentEmailId,
		FirstName:          listMember.FirstName,
		SiteDomain:         siteDomain,
		SiteOwnerFirstName: ownerFirstName,
		PageTitle:          opts.PageTitle,
		DestinationUrl:     destinationUrl,
		Params:             opts.TemplateParams,
	})
	if err != nil {
		panic(err)
	}
	htmlEmailString := buffer.String()
	// Convert the HTML to plaintext
	textEmailString, err := html2text.FromString(htmlEmailString)
	if err != nil {
		panic(err)
	}
	// Send the email
	awsResponse, err := ses.EnvConfig.SendEmailHTML(fromEmail, listMember.Email, subject, textEmailString, htmlEmailString)

	// Unmarshall the response
	err, awsMessageId := unmarshallAwsResponse(err, awsResponse)
	if err != nil {
		panic(err)
	}

	processSentEmail(err, htmlEmailString, textEmailString, awsMessageId, sentEmail, &listMember)
}

func processSentEmail(err error, htmlEmailString string, textEmailString string, awsMessageId string, sentEmail *SentEmail, listMember *ListMember) {
	if err == nil {
		fmt.Printf("Sent email to: %s,\n\nhtml: %s\n\n\n%s\n\n\nawsMessageId: %s\n", listMember.Email, htmlEmailString, textEmailString, awsMessageId)
	} else {
		panic(fmt.Sprintf("ERROR sending email: %v\n", err))
	}
	// Record the email sent event if we didn't get an error from AWS
	SoftsideDB.Insert(&EmailAction{SentEmailId: sentEmail.Id, Action: "sent"})

	// Update the sent email with the AWS MessageId
	sentEmail.ThirdPartyId = awsMessageId
	err = SoftsideDB.Update(sentEmail)
	if err != nil {
		log.Printf("ERROR updating SentEmail.ThirdPartyId with AWS MessageId: %v\n", err)
	}
}

func unmarshallAwsResponse(err error, awsResponse string) (error, string) {
	// Unmarshall the AWS XML reponse into a struct
	var sendEmailResponse SendEmailResponse
	err = xml.Unmarshal([]byte(awsResponse), &sendEmailResponse)
	if err != nil {
		log.Printf("ERROR umarshalling AWS SendEmailResponse: %v\n", err)
	}
	return err, sendEmailResponse.MessageId
}
