package softside_test

import (
	"testing"
	"strings"
	"softside/tests/sampleEmails"
	"github.com/veqryn/go-email/email"
)


func TestParseEmail2(t *testing.T) {
	msg, err := email.ParseMessage(strings.NewReader(sampleEmails.EmailWithLongBody))
	if err != nil {
		t.Error(err)
	}
	htmlBody := msg.PartsContentTypePrefix("text/html")
	println(string(htmlBody[0].Body))
	println("\n\n\n\n\n")

	textBody := msg.PartsContentTypePrefix("text/plain")
	println(string(textBody[0].Body))
	println("\n\n\n\n\n")

	subject := msg.Header.Get("Subject")
	println(subject)
	println("\n\n\n\n\n")

	//println(string(textBody[0].Bytes()))
	bytes, err := msg.Bytes()
	println(string(bytes))
}