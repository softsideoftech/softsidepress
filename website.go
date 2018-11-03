package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"softside/forwardEmail"
	"softside/softmail"
)

func main() {
	runService()
}

func runService() {
	log.Println("Starting website")

	rawCtx := softmail.NewRawRequestCtx()
	if !rawCtx.DevMode {
		// Start processing SQS messages from SES in the background
		go softmail.StartSqs()

		// Start the SMTP server for forwarding emails
		go forwardEmail.StartSmtpServer()
	}
	
	// We send scheduled emails even in DevMode.
	go rawCtx.StartEmailScheduler()

	// Start the website
	startWebsite()
}

func startWebsite() {
	registerHandler("/yes-please/", softmail.Resubscribe, false)
	registerHandler("/bye/", softmail.Unsubscribe, false)
	registerHandler("/join/", softmail.Join, false)
	registerHandler("/request-login-link/", softmail.RequestLoginLink, true)
	registerHandler("/favicon.ico", HandleFavicon, false)
	registerHandler("/gen_link", softmail.GenerateTrackingLink, false)
	registerHandler("/ping", softmail.TrackTimeOnPage, false)
	registerHandler("/", softmail.HandleNormalRequest, false)
	http.ListenAndServe(":8080", nil)
}

func HandleFavicon(ctx *softmail.RequestContext) {
	http.Redirect(ctx.W, ctx.R, softmail.FavIconUrl, http.StatusTemporaryRedirect)
}

func registerHandler(pattern string, httpHandler func(ctx *softmail.RequestContext), initCtx bool) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := softmail.NewRequestCtx(w, r, initCtx)
		defer func() {
			stack := debug.Stack()
			if r := recover(); r != nil {
				ctx.SendUserFacingError("ERROR processing pattern: " + pattern, fmt.Sprintf("%v\n%s", r, string(stack)))
			}
		}()

		httpHandler(ctx)
	});
}
