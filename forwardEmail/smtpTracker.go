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
)

var softmailPassword string = os.Getenv("SOFTMAIL_FORWARDING_PASSWORD")
type Backend struct{}

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
	return nil, smtp.ErrAuthRequired
}

type User struct{}

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