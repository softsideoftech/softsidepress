package forwardEmail

import (
	"errors"
	"io/ioutil"
	"log"
	"github.com/emersion/go-smtp"
	"io"
	"softside/softmail"
	"github.com/veqryn/go-email/email"
	"bytes"
	"os"
	"strings"
	"github.com/sourcegraph/go-ses"
)

var softmailPassword string = os.Getenv("SOFTMAIL_FORWARDING_PASSWORD")
type Backend struct{}

type User struct{}

type AnnonymousUser struct {
	*User
}

func (bkd *Backend) Login(username, password string) (smtp.User, error) {
	log.Println("received email from username: %v, password: %v", username, password)
	if true {
		return &User{}, nil	
	}
	if username != "username" || password != softmailPassword {
		return nil, errors.New("Invalid username or password")
	}
	return &User{}, nil
}

// Require clients to authenticate using SMTP AUTH before sending emails
func (bkd *Backend) AnonymousLogin() (smtp.User, error) {
	return AnnonymousUser{}, nil
}


func (u AnnonymousUser) Send(from string, to []string, r io.Reader) error {
	log.Println("Translating message:", from, to)

	if b, err := ioutil.ReadAll(r); err != nil {
		return err
	} else {
		log.Println("Data:", string(b))
		msg, err := email.ParseMessage(bytes.NewReader(b))
		if err != nil {
			return err
		}



		htmlMessages, err := softmail.FindPartType(msg, "text/html")
		if err != nil || len(htmlMessages) == 0 {
			log.Printf("ERROR finding content type 'text/html': %v\n\nMESSAGE: %v\v", string(b), err)
			return errors.New("ERROR finding content type 'text/html'")
		}
		htmlEmailBodyBytes := htmlMessages[len(htmlMessages)-1].Body
		htmlEmailBody := string(htmlEmailBodyBytes)

		translation, err := softmail.TranslateHtml(htmlEmailBody)
		if err != nil {
			log.Printf("ERROR translating email body:\n\n%v\n\nERROR MESSAGE: %v\n\n", htmlEmailBody, err)
			return errors.New("ERROR translating email body")
		}
		if translation != "" {
			// Add the translation tag into the subject
			subject := "[TRNS] " + msg.Header.Get("Subject")

			// Use the Return-Path header for the recipient
			recipient := strings.Trim(msg.Header.Get("From"), "<>")
			// todo vg: make all this stuff configurable, particularly the sender 
			sender := "vlad@softsideoftech.com"

			awsResponse, err := ses.EnvConfig.SendEmailHTML(sender, recipient, subject, "", translation)

			log.Printf("\nAWS SMTP RESPONSE:%s,%v\n:", awsResponse, err);

		}
	}
	return nil

}

func (u *User) Send(from string, to []string, r io.Reader) error {
	log.Println("Sending message:", from, to)

	if b, err := ioutil.ReadAll(r); err != nil {
		return err
	} else {
		log.Println("Data:", string(b))
		msg, err := email.ParseMessage(bytes.NewReader(b))
		if err != nil {
			return err
		}
		softmail.ForwardEmail(from, to[0], msg)
	}
	return nil
}

func (u *User) Logout() error {
	return nil
}

func StartSmtpServer() {
	be := &Backend{}

	s := smtp.NewServer(be)

	s.Addr = ":25"
	s.Domain = "localhost"
	s.MaxIdleSeconds = 300
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	log.Println("Starting server at", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}