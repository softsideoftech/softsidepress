package main

import (
	"fmt"
	"github.com/sourcegraph/go-ses"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"os"
)

func main() {

	markdownEmailBody, err := ioutil.ReadFile("/Users/vlad/go/src/hello/testemail.txt")

	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		os.Exit(1)
	}

	htmlEmailBytes := blackfriday.Run(markdownEmailBody)
	htmlEmailString := string(htmlEmailBytes)

	fmt.Print(htmlEmailString)

	// Change the From address to a sender address that is verified in your Amazon SES account.
	from := "vladgiverts@softsideoftech.com"
	to := "vladgiverts@softsideoftech.com"

	// EnvConfig uses the AWS credentials in the environment variables $AWS_ACCESS_KEY_ID and
	// $AWS_SECRET_KEY.
	res, err := ses.EnvConfig.SendEmailHTML(from, to, "Hello, world 2", string(markdownEmailBody), htmlEmailString)
	if err == nil {
		fmt.Printf("Sent email: %s...\n", res[:32])
	} else {
		fmt.Printf("Error sending email: %s\n", err)
	}
}
