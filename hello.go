package main

import (
	"fmt"
	"net/http"
	"softside/softmail"
)


func main() {
	runService()

	//testSendMail()
	//softmail.StartSqs()
}

func testSendMail() {
	err := softmail.Sendmail("test body", "/Users/vlad/go/src/softside/emails/testemail.md", "vlad@softsideoftech.com")
	if (err != nil) {
		fmt.Println(err)
	}
}

func runService() {
	http.HandleFunc("/yes-please/", softmail.Resubscribe)
	http.HandleFunc("/bye/", softmail.Unsubscribe)
	http.HandleFunc("/join/", softmail.Join)
	http.HandleFunc("/favicon.ico", HandleFavicon)
	http.HandleFunc("/gen_link", softmail.GenerateTrackingLink)
	http.HandleFunc("/", softmail.TrackRequest)
	http.ListenAndServe(":8080", nil)
}

func HandleFavicon(w http.ResponseWriter, r *http.Request) {
	// todo: make this configurable
	favIconUrl := "https://s3-us-west-2.amazonaws.com/static.softsideoftech.com/favicon.ico"
	http.Redirect(w, r, favIconUrl, http.StatusTemporaryRedirect)
}
