package softside_test

import (
	"testing"
	"strings"
	"softside/tests/sampleEmails"
	"github.com/veqryn/go-email/email"
	"softside/softmail"
)

func TestEmailWithLongBody(t *testing.T) {
	//testEmail(sampleEmails.EmailWithLongBody, t)
}
func TestEmailWithMultiCc(t *testing.T) {
	testEmail(sampleEmails.EmailWithMultiCc, t)
}

func testEmail(emailMessageString string, t *testing.T) {
	msg, err := email.ParseMessage(strings.NewReader(emailMessageString))
	htmlBody := msg.PartsContentTypePrefix("text/html")
	htmlBodyStr := string(htmlBody[0].Body)
	println(htmlBodyStr)
	println("\n\n\n\n\n")
	textBody := msg.PartsContentTypePrefix("text/plain")
	textBodyStr := string(textBody[0].Body)
	println(textBodyStr)
	println("\n\n\n\n\n")
	subject := msg.Header.Get("Subject")
	println(subject)
	println("\n\n\n\n\n")
	//println(string(textBody[0].Bytes()))
	bytes, err := msg.Bytes()
	println(string(bytes))
	if softmail.DevelopmentMode {
		softmail.ForwardEmail("vlad@softsideoftech.com", "stacyap@gmail.com", msg)
	}
	if err != nil {
		t.Error(err)
	}
}