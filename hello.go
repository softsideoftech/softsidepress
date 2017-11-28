package main

import (
	"fmt"
	"net/http"
	"softside/softmail"
	"strings"
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

}

func testSendMail() {
	err := softmail.Sendmail("test body", "/Users/vlad/go/src/softside/testemail.txt", "vlad@softsideoftech.com")
	if (err != nil) {
		fmt.Println(err)
	}
}

func testEmailTracker() {
	http.HandleFunc("/", HandleRequest)
	http.ListenAndServe(":8080", nil)
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Processing request: &s", r.RequestURI)
	if (strings.HasSuffix(r.RequestURI, "/favicon.ico")) {
		// TODO: make this configurable?
		favIconUrl := "http://static.softsideoftech.com/favicon.ico"
		http.Redirect(w, r, favIconUrl, http.StatusTemporaryRedirect)
	} else {
		softmail.HandleEmailOpen(w, r)
	}
}
