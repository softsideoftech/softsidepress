package softside_test

import (
	"testing"
	"strings"
	"softside/tests/sampleEmails"
	"github.com/veqryn/go-email/email"
	"softside/softmail"
	"github.com/jhillyerd/enmime"
	"fmt"
)

func TestEmailWithLongBody(t *testing.T) {
	//testEmail(sampleEmails.EmailWithLongBody, t)
}

func TestEmailWithMultiCc(t *testing.T) {
	//testEmail(sampleEmails.EmailWithMultiCc, t)
}

func TestEmailWithBigAttachment(t *testing.T) {
	testEmail(sampleEmails.EmailSampleWithAttachment, t)
}

func testEmail(emailMessageString string, t *testing.T) {
	msg, err := email.ParseMessage(strings.NewReader(emailMessageString))
	checkErr(err, t)

	htmlBody, err := softmail.FindPartType(msg, "text/html")
	checkErr(err, t)
	textBody, err := softmail.FindPartType(msg, "text/plain")
	checkErr(err, t)
	
	fmt.Printf("%s\n\n\n\n\n", string(htmlBody[0].Body))
	fmt.Printf("%s\n\n\n\n\n", string(textBody[0].Body))
	fmt.Printf("SUBJECT: %", msg.Header.Get("Subject"))

	bytes, err := msg.Bytes()
	println(string(bytes))
	if softmail.DevelopmentMode {
		softmail.ForwardEmail("vlad@softsideoftech.com", "vgiverts+123@gmail.com", msg)
	}
	if err != nil {
		t.Error(err)
	}
}

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}

func testEmail2(emailMessageString string, t *testing.T) {
	env, err := enmime.ReadEnvelope(strings.NewReader(emailMessageString))
	if err != nil {
		panic(err)
	}

	// Headers can be retrieved via Envelope.GetHeader(name).
	fmt.Printf("From: %v\n", env.GetHeader("From"))

	// Address-type headers can be parsed into a list of decoded mail.Address structs.
	alist, _ := env.AddressList("To")
	for _, addr := range alist {
		fmt.Printf("To: %s <%s>\n", addr.Name, addr.Address)
	}

	// enmime can decode quoted-printable headers.
	fmt.Printf("Subject: %v\n", env.GetHeader("Subject"))

	// The plain text body is available as mime.Text.
	fmt.Printf("Text Body: %v chars\n\n\n", (env.Text))

	// The HTML body is stored in mime.HTML.
	fmt.Printf("HTML Body: %v chars\n\n\n", (env.HTML))

	// mime.Inlines is a slice of inlined attacments.
	fmt.Printf("Inlines: %v\n\n\n", (env.Inlines))

	// mime.Attachments contains the non-inline attachments.
	fmt.Printf("Attachments: %v\n", len(env.Attachments))
}
