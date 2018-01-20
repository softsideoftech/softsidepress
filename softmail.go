package main

import (
	"softside/softmail"
	"fmt"
)

func main() {
	testSendMail()
}

func testSendMail() {
	err := softmail.Sendmail("test body", "/Users/vlad/go/src/softside/emails/testemail.md", "vlad@softsideoftech.com")
	if err != nil {
		fmt.Println(err)
	}
}