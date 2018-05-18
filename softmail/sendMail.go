package softmail

import (
	"io/ioutil"
	"fmt"
	"gopkg.in/russross/blackfriday.v2"
	"github.com/sourcegraph/go-ses"
	"crypto/md5"
	"hash/fnv"
	"strings"
	"regexp"
	"encoding/xml"
	htmlTemplate "html/template"
	"bytes"
	"github.com/jaytaylor/html2text"
)

const emailSuffixMdFile = "emailSuffix.md" // TODO: make this a relative path

type EmailTemplateParams struct {
	FirstName   string
	SentEmailId string
	SiteDomain  string
}

func emailTemplateToId(subject string, body []byte) int64 {
	hash := md5.New()
	hash.Write(body)
	hash.Write([]byte(subject))
	md5Sum := hash.Sum(nil)
	hash64 := fnv.New64()
	hash64.Write(md5Sum)
	return int64(hash64.Sum64()) // make it signed to conform with the Postgres "bigint" type
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
		fmt.Printf("Problem parsing SentEmailId from url: %s, error: %v", path, err)
		return 0
	}
	return sentEmailId
}

func Sendmail(subject string, templateFile string, fromEmail string, memberGroupName string) error {

	// Load the template file
	markdownEmailBodyBytes, err := ioutil.ReadFile(templateFile)
	markdownEmailBody := string(markdownEmailBodyBytes)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		panic(err)
	}

	// Save the email template in the DB if it doesn't exist
	emailTemplateId := emailTemplateToId(subject, markdownEmailBodyBytes)
	emailTemplate := EmailTemplate{Id: emailTemplateId}
	err = SoftsideDB.Select(&emailTemplate)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			emailTemplate.Subject = subject
			emailTemplate.Body = markdownEmailBody
			err = SoftsideDB.Insert(&emailTemplate)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	// Add the SendEmailId template parameter to all internal links
	markdownEmailBody = linkRegex.ReplaceAllString(markdownEmailBody, "($1$2-{{.SentEmailId}})")

	// Load the suffix template and append it to the markdown template
	suffixEmailBodyBytes, err := ioutil.ReadFile(SoftsideContentPath + "/" + emailSuffixMdFile)
	suffixEmailBody := string(suffixEmailBodyBytes)
	if err != nil {
		panic(err)
	}
	markdownEmailBody += suffixEmailBody

	// Turn the markdown into HTML
	htmlEmailTemplateString := string(blackfriday.Run([]byte(markdownEmailBody)))

	// Create the HTML template (must be HTML and not TEXT to escape user supplied values such as FirstName)
	parsedEmailTempalte, err := htmlTemplate.New(templateFile).Parse(htmlEmailTemplateString)

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
		res, err := ses.EnvConfig.SendEmailHTML(fromEmail, listMember.Email, subject, textEmailString, htmlEmailString)
		if err == nil {
			fmt.Printf("Sent email: %s\n\n\n%s\n\n\n%s\n", htmlEmailString, textEmailString, res)
		} else {
			panic(err)
		}

		// Record the email sent event if we didn't get an error from AWS
		SoftsideDB.Insert(&EmailAction{SentEmailId: sentEmail.Id, Action: "sent"})

		// Unmarshall the AWS XML reponse into a struct
		var sendEmailResponse SendEmailResponse
		err = xml.Unmarshal([]byte(res), &sendEmailResponse)
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

	return nil
}
