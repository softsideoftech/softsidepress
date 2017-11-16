package main

import (
	"fmt"
	"github.com/sourcegraph/go-ses"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"os"
	"net/http"
	"regexp"
	"github.com/go-pg/pg"
	"time"
)


type Page struct {
	Title string
	Body  []byte
}

func main() {
	testDB()
	//testRedirect()
	//testEmail()
}

type EmailTemplate struct {
	Id     uint32
	Subject   string
	Body string
}
func (u EmailTemplate) String() string {
	return fmt.Sprintf("EmailTemplate<%d %s %v>", u.Id, u.Subject, u.Body)
}

func testDB() {
	db := pg.Connect(&pg.Options{
		User: "vlad",
	})

	emailTemplate1 := &EmailTemplate{
		Subject:   "test subject" + time.Now().String(),
		Body: "this is a test body",
	}
	err := db.Insert(emailTemplate1)
	if err != nil {
		panic(err)
	}


	emailTemplate := EmailTemplate{Id: emailTemplate1.Id}
	err = db.Select(&emailTemplate)
	if err != nil {
		panic(err)
	}
	fmt.Println(emailTemplate)


	// Select all email templates.
	var emailTemplates []EmailTemplate
	err = db.Model(&emailTemplates).Select()
	if err != nil {
		panic(err)
	}
	fmt.Println(emailTemplates)
}


func testRedirect() {
	http.HandleFunc("/redirect/", makeHandler(handleRedirect))
	http.ListenAndServe(":8080", nil)
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func handleRedirect(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "https://softsideoftech.com?t=" + title, http.StatusFound)
}

var validPath = regexp.MustCompile("^/(redirect)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[1])
	}
}

func testEmail() {
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
	to := "vgiverts@gmail.com"
	// EnvConfig uses the AWS credentials in the environment variables $AWS_ACCESS_KEY_ID and
	// $AWS_SECRET_KEY.
	res, err := ses.EnvConfig.SendEmailHTML(from, to, "Hello, world 2", string(markdownEmailBody), htmlEmailString)
	if err == nil {
		fmt.Printf("Sent email: %s...\n", res[:32])
	} else {
		fmt.Printf("Error sending email: %s\n", err)
	}
}
