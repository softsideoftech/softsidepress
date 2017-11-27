package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"softside/softmail"
)

type Page struct {
	Title string
	Body  []byte
}

func main() {
	//db := pg.Connect(&pg.Options{
	//	User: "vlad",
	//})

		testEmailTracker()

	//testSendMail()
	//softmail.StartSqs()

	//testRedirect()
}
func testSendMail() {
	err := softmail.Sendmail("test body", "/Users/vlad/go/src/softside/testemail.txt", "vlad@softsideoftech.com")
	if (err != nil) {
		fmt.Println(err)
	}
}

//func (u softmail.EmailTemplate) String() string {
//	return fmt.Sprintf("EmailTemplate<%d %s %v>", u.Id, u.Subject, u.Body)
//}

func testEmailTracker() {
	//TODO: route requests for "/favicon.ico"
	http.HandleFunc("/", softmail.HandleEmailOpen)
	http.ListenAndServe(":8080", nil)
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
	http.Redirect(w, r, "https://softsideoftech.com?t="+title, http.StatusFound)
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
