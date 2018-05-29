package softside_test

import (
	"testing"
	"strings"
	"softside/tests/sampleEmails"
	"github.com/veqryn/go-email/email"
	"softside/softmail"
)


func TestParseEmail2(t *testing.T) {
	msg, err := email.ParseMessage(strings.NewReader(sampleEmails.EmailWithLongBody))

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


	if false {
		softmail.ForwardEmail("vgiverts@gmail.com", msg)
	}
	if err != nil {
		t.Error(err)
	}
}