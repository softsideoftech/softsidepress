package softmail

import (
	"io/ioutil"
	"fmt"
	"gopkg.in/russross/blackfriday.v2"
	"github.com/sourcegraph/go-ses"
	"crypto/md5"
	"hash/fnv"
	"github.com/go-pg/pg"
	"strings"
	"regexp"
)

func emailTemplateToId(subject string, body []byte) int64 {
	hash := md5.New()
	hash.Write(body)
	hash.Write([]byte(subject))
	md5Sum := hash.Sum(nil)
	hash64 := fnv.New64()
	hash64.Write(md5Sum)
	return int64(hash64.Sum64()) // make it signed to conform with the Postgres "bigint" type
}

var firstNameRegex = regexp.MustCompile("(\\{\\{first_name\\}\\})")
var sentEmailIdRegex = regexp.MustCompile("(\\{\\{sent_email_id\\}\\})")

var linkRegex = regexp.MustCompile("\\((https://)(.+?)\\)")
var extractSentEmailIdFromUrlEnd = regexp.MustCompile("/.*/(.*)")

// todo: make these configurable
const trackingSubDomain = "www"

func decodeSendMailIdFromUriEnd(path string) SentEmailId {
	submatch := extractSentEmailIdFromUrlEnd.FindStringSubmatch(path)
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

func Sendmail(subject string, templateFile string, fromEmail string) error {

	// Connect to the DB
	// TODO: Replace the naive DB connection with connection pooling and a config driven connection string
	db := pg.Connect(&pg.Options{
		User: "vlad",
	})

	// Load the template file
	markdownEmailBodyBytes, err := ioutil.ReadFile(templateFile)
	markdownEmailBody := string(markdownEmailBodyBytes)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return err
	}

	// Save the email template in the DB if it doesn't exist
	emailTemplateId := emailTemplateToId(subject, markdownEmailBodyBytes)
	emailTemplate := EmailTemplate{Id: emailTemplateId}
	err = db.Select(&emailTemplate)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			emailTemplate.Subject = subject
			emailTemplate.Body = markdownEmailBody
			err = db.Insert(&emailTemplate)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Load the email list
	var listMembers []ListMember
	err = db.Model(&listMembers).Select()
	if err != nil {
		return err
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
		err := db.Insert(sentEmail)
		if err != nil {
			return err
		}

		// Base64 encode the SentEmail id
		encodedSentEmailId := EncodeId(sentEmail.Id)

		// Templatize the first name and sent email id
		// todo: replace with go templates
		markdownEmailBody = templatizeParams(markdownEmailBody, &listMember, encodedSentEmailId)

		// Templatize the links
		markdownEmailBody := linkRegex.ReplaceAllString(markdownEmailBody, "($1$2-"+encodedSentEmailId+")")

		// Generate the html body
		htmlEmailString := string(blackfriday.Run([]byte(markdownEmailBody)))

		// Append the tracking pixel
		htmlEmailString += "<img src='http://" + trackingSubDomain + "." + siteDomain + "/" + encodedSentEmailId + ".jpg'/>"

		// Send the email
		res, err := ses.EnvConfig.SendEmailHTML(fromEmail, listMember.Email, subject, markdownEmailBody, htmlEmailString)
		if err == nil {
			fmt.Printf("Sent email: %s...\n", res)
		} else {
			fmt.Printf("Error sending email: %s\n", err)
		}

		// TODO: Save id that comes back from SES so we could track bounces and complaints
		// TODO: Return an actual image so email clients don't keep re-requesting
	}

	return nil
}

func templatizeParams(markdownEmailBody string, listMember *ListMember, encodedSentEmailId string) string {
	markdownEmailBody = firstNameRegex.ReplaceAllString(markdownEmailBody, listMember.FirstName)
	return sentEmailIdRegex.ReplaceAllString(markdownEmailBody, encodedSentEmailId)
}
