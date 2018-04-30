package main

import (
	"fmt"
	"net/http"
	"softside/softmail"
	"io/ioutil"
)

func main() {
	runService()
}

func runService() {

	// Start processing SQS messages from SES in the background
	go softmail.StartSqs()

	http.HandleFunc("/yes-please/", softmail.Resubscribe)
	http.HandleFunc("/bye/", softmail.Unsubscribe)
	http.HandleFunc("/join/", softmail.Join)
	http.HandleFunc("/favicon.ico", HandleFavicon)
	http.HandleFunc("/gen_link", softmail.GenerateTrackingLink)
	http.HandleFunc("/ping", softmail.TrackTimeOnPage)
	http.HandleFunc("/lkdcnt", getLinkedInCounts)
	http.HandleFunc("/", softmail.HandleNormalRequest)
	http.ListenAndServe(":8080", nil)
}

func HandleFavicon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, softmail.FavIconUrl, http.StatusTemporaryRedirect)
}

// Proxy for LinedIn counts because calling LinkedIn from the browser causes a CORS error.
func getLinkedInCounts(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query()["url"][0]
	resp, err := http.Get("https://www.linkedin.com/countserv/count/share?format=json&url=" + url)
	if err == nil {
		defer resp.Body.Close()
		body, err2 := ioutil.ReadAll(resp.Body)
		w.Write(body)
		err = err2
	}

	if (err != nil) {
		fmt.Printf("Couldn't obtain linked count for url: %s, error: %v", url, err)
	}
}
