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
)

const emailSuffixMdFile = "emailSuffix.md"
const forwardedEmailSuffixMdFile = "forwardedEmailSuffix.md"
const trackingPixelPath = "/pixie/"
const trackingPixelMarkdown = "![](https://{{.SiteDomain}}" + trackingPixelPath + "{{.SentEmailId}}.png)"
var trackingPrefixTemplate, _ = htmlTemplate.New("tracking-prefix").Parse(trackingPixelMarkdown)

type EmailTemplateParams struct {
	FirstName   string
	SentEmailId string
	SiteDomain  string
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

var linkRegex = regexp.MustCompile("\\((https://)(.+?)\\)")
var extractSentEmailIdFromUrlEndSlash = regexp.MustCompile("/.*/(.*)")

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

func ForwardEmail(recipient string, msg *email.Message) {

	var textEmailBody string
	var htmlEmailBody string

	htmlMessages := msg.PartsContentTypePrefix("text/html")
	if len(htmlMessages) > 0 {
		htmlEmailBodyBytes := htmlMessages[len(htmlMessages) - 1].Body
		htmlEmailBody = string(htmlEmailBodyBytes)
	}

	textMessages := msg.PartsContentTypePrefix("text/plain")
	if len(textMessages) > 0 {
		textEmailBodyBytes := textMessages[len(textMessages) - 1].Body
		textEmailBody = string(textEmailBodyBytes)
	}

	subject := msg.Header.Get("Subject")


	// Obtain the emailTemplateId
	emailTemplateId := obtainEmailTemplateId(subject, textEmailBody, recipient)

	// Create a SentEmail record
	sentEmail := &SentEmail{
		EmailTemplateId: emailTemplateId,
		ListMemberId:    0, // This is a direct email, not a list email, so a ListMemberId might not exist. Using the phantom list member "0" to represent this case.
		// todo: save the email address being sent to 
		// (maybe insert a new type of non-subscribed user into list members?)
	}
	err := SoftsideDB.Insert(sentEmail)
	if err != nil {
		panic(err)
	}
	// Base64 encode the SentEmail id
	encodedSentEmailId := EncodeId(sentEmail.Id)

	// Load the suffix template
	renderedPrefix := obtainTrackingPrefix(encodedSentEmailId)

	// Append the to only the html email
	htmlEmailBody = renderedPrefix + htmlEmailBody

	// Place the bytes back into the message
	if len(htmlMessages) > 0 {
		htmlMessages[len(htmlMessages)-1].Body = []byte(htmlEmailBody)
	}

	// Place the bytes back into the message (todo: this may be unnecessary now that we're not putting tracking into text emails)
	if len(textMessages) > 0 {
		textMessages[len(textMessages)-1].Body = []byte(textEmailBody)
	}

	// Actually send the email
	msgBytes, err := msg.Bytes()
	if err != nil {
		panic(err)
	}
	awsResponse, err := ses.EnvConfig.SendRawEmail(msgBytes)

	// Perfom the bookkeeping
	processSentEmail(err, htmlEmailBody, textEmailBody, awsResponse, sentEmail)
}

// todo: not using this right now
func obtainTrackingSuffix(encodedSentEmailId string) string {
	suffixEmailBodyBytes, err := ioutil.ReadFile(SoftsideContentPath + "/emails/" + forwardedEmailSuffixMdFile)
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
		SentEmailId: encodedSentEmailId,
		SiteDomain:  siteDomain,
	})
	renderedSuffix := buffer.String()
	return renderedSuffix
}

func obtainTrackingPrefix(encodedSentEmailId string) string {
	// Parse and render the suffix template
	return fmt.Sprintf("<img src=\"https://softsideoftech.com/pixie/%s.png\"/>", encodedSentEmailId);
}


func SendEmailToGroup(subject string, templateFileName string, fromEmail string, memberGroupName string) error {

	// Load the template file
	markdownEmailBodyBytes, err := ioutil.ReadFile(templateFileName)
	markdownEmailBody := string(markdownEmailBodyBytes)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		panic(err)
	}

	emailTemplateId := obtainEmailTemplateId(subject, markdownEmailBody, "")

	// Add the SendEmailId template parameter to all internal links
	markdownEmailBody = linkRegex.ReplaceAllString(markdownEmailBody, "($1$2-{{.SentEmailId}})")

	// Load the suffix template and append it to the markdown template
	suffixEmailBodyBytes, err := ioutil.ReadFile(SoftsideContentPath + "/emails/" + emailSuffixMdFile)
	suffixEmailBody := string(suffixEmailBodyBytes)
	if err != nil {
		panic(err)
	}
	markdownEmailBody += suffixEmailBody

	// Turn the markdown into HTML
	htmlEmailTemplateString := string(blackfriday.Run([]byte(markdownEmailBody)))

	// Create the HTML template (must be HTML and not TEXT to escape user supplied values such as FirstName)
	parsedEmailTempalte, err := htmlTemplate.New(templateFileName).Parse(htmlEmailTemplateString)

	// Load the email list
	//var listMembers []ListMember
	//err = SoftsideDB.Model(&listMembers).Select()

	var listMembers []ListMember

	if memberGroupName == "all" {
		// Select all members where unsubscribed is nil (ie, they never explicitly unsubscribed)
		err = SoftsideDB.Model(&listMembers).Where("unsubscribed IS NULL", nil).Select()
	} else {
		_, err = SoftsideDB.Query(&listMembers, `
	select l.* from list_members l, member_groups g 
	where l.id = g.list_member_id and g.name = ? AND l.unsubscribed IS NULL`, memberGroupName)
	}

	fmt.Printf("listMembers: %v\n", listMembers)
	if err != nil {
		panic(err)
	}

	// For each member
	// 		Create a SentEmail record
	//		Templatize the name and links. Include the email_sent id and member_id
	//		Generate an html version
	//		Send the email
	//		Record the fact the email was sent
	for _, listMember := range listMembers {
		sendEmailToListMember(emailTemplateId, listMember, parsedEmailTempalte, fromEmail, subject)
	}

	return nil
}

func obtainEmailTemplateId(subject string, emailBody string, recipient string) (EmailTemplateId) {
	// Save the email template in the DB if it doesn't exist
	emailTemplateId := emailTemplateToId(subject, []byte(emailBody), recipient)
	emailTemplate := EmailTemplate{Id: emailTemplateId}
	err := SoftsideDB.Select(&emailTemplate)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			emailTemplate.Subject = subject
			emailTemplate.Body = emailBody
			err = SoftsideDB.Insert(&emailTemplate)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	return emailTemplateId
}

func sendEmailToListMember(emailTemplateId EmailTemplateId, listMember ListMember, parsedEmailTempalte *htmlTemplate.Template, fromEmail string, subject string) {
	// Create a SentEmail record
	sentEmail := &SentEmail{
		EmailTemplateId: emailTemplateId,
		ListMemberId:    listMember.Id,
	}
	err := SoftsideDB.Insert(sentEmail)
	if err != nil {
		panic(err)
	}
	// Base64 encode the SentEmail id
	encodedSentEmailId := EncodeId(sentEmail.Id)
	// Render the HTML template
	buffer := &bytes.Buffer{}
	err = parsedEmailTempalte.Execute(buffer, &EmailTemplateParams{
		SentEmailId: encodedSentEmailId,
		FirstName:   listMember.FirstName,
		SiteDomain:  siteDomain,
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
	processSentEmail(err, htmlEmailString, textEmailString, awsResponse, sentEmail)
}

func processSentEmail(err error, htmlEmailString string, textEmailString string, awsResponse string, sentEmail *SentEmail) {
	if err == nil {
		fmt.Printf("Sent email: %s\n\n\n%s\n\n\n%s\n", htmlEmailString, textEmailString, awsResponse)
	} else {
		panic(err)
	}
	// Record the email sent event if we didn't get an error from AWS
	SoftsideDB.Insert(&EmailAction{SentEmailId: sentEmail.Id, Action: "sent"})
	// Unmarshall the AWS XML reponse into a struct
	var sendEmailResponse SendEmailResponse
	err = xml.Unmarshal([]byte(awsResponse), &sendEmailResponse)
	if err != nil {
		panic(err)
	}
	// Update the sent email with the AWS MessageId
	sentEmail.ThirdPartyId = sendEmailResponse.MessageId
	err = SoftsideDB.Update(sentEmail)
	if err != nil {
		panic(err)
	}
}
