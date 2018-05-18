package main

import (
	"net/http"
	"softside/softmail"
	"softside/forwardEmail"
)

func main() {
	runService()
}

func runService() {

	// Start processing SQS messages from SES in the background
	go softmail.StartSqs()

	// Start the SMTP server for forwarding emails
	go forwardEmail.StartSmtpServer()

	// Start the website
	startWebsite()
}

func startWebsite() {
	http.HandleFunc("/yes-please/", softmail.Resubscribe)
	http.HandleFunc("/bye/", softmail.Unsubscribe)
	http.HandleFunc("/join/", softmail.Join)
	http.HandleFunc("/favicon.ico", HandleFavicon)
	http.HandleFunc("/gen_link", softmail.GenerateTrackingLink)
	http.HandleFunc("/ping", softmail.TrackTimeOnPage)
	http.HandleFunc("/", softmail.HandleNormalRequest)
	http.ListenAndServe(":8080", nil)
}

func HandleFavicon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, softmail.FavIconUrl, http.StatusTemporaryRedirect)
}