package softmail

import (
	"time"
	"bytes"
	"github.com/go-pg/pg"
	"net/http"
	"fmt"
)

const unsubscribeTemplate = "src/softside/mgmt-pages/unsubscribe.md"
const resubscribeTemplate = "src/softside/mgmt-pages/resubscribe.md"
const errorTemplate = "src/softside/mgmt-pages/error.md"
const baseHtmlTemplate = "src/softside/html/base.html"
const owner = "Vlad"

type ListMemberParams struct {
	FirstName string
	EncodedId string
}

type SiteOwner struct {
	OwnerName string
}

func sendUserFacingError(logMessage string, err error, w http.ResponseWriter) {
	fmt.Printf(logMessage + "err: %v", err)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
	renderMarkdownToHtmlTemplate(w, baseHtmlTemplate, "Something isn't right...", errorTemplate, SiteOwner{owner})
}

func (ctx *RequestContext) someScribe(w http.ResponseWriter, r *http.Request, templateFile string, pageTitle string) *ListMember {

	// Find the SentEmailId in the url
	sentEmailId := decodeSendMailIdFromUriEnd(r.URL.Path)
	if sentEmailId == 0 {
		sendUserFacingError("Couldn't find SentEmailId in url: %v", nil, w)
		return nil
	}

	// Load the ListMember from the DB
	listMemberId, err := ctx.getListMemberIdFromSentEmail(sentEmailId)
	if (err != nil) {
		sendUserFacingError("", err, w)
		return nil
	}
	listMember := &ListMember{Id: listMemberId}
	err = ctx.db.Select(listMember)
	if err != nil {
		sendUserFacingError("Couldn't find list member in url: %v", err, w)
		return nil
	}

	// Run the template
	buffer := &bytes.Buffer{}
	renderMarkdownToHtmlTemplate(buffer, baseHtmlTemplate, pageTitle, templateFile, ListMemberParams{listMember.FirstName, EncodeId(sentEmailId)})


	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "foo bar!", time.Now(), bytes.NewReader(buffer.Bytes()))

	return listMember
}

func Resubscribe(w http.ResponseWriter, r *http.Request) {
	// Connect to the DB
	// TODO: Replace the naive DB connection with connection pooling and a config driven connection string
	ctx := &RequestContext{
		db: pg.Connect(&pg.Options{
			User: "vlad",
		}),
	}

	listMember := ctx.someScribe(w, r, resubscribeTemplate, "Welcome back :)")

	// Update the unsubscribe status
	if (listMember != nil) {
		listMember.Unsubscribed = nil
		ctx.db.Update(listMember)
	}
}

func Unsubscribe(w http.ResponseWriter, r *http.Request) {
	// Connect to the DB
	// TODO: Replace the naive DB connection with connection pooling and a config driven connection string
	ctx := &RequestContext{
		db: pg.Connect(&pg.Options{
			User: "vlad",
		}),
	}

	listMember := ctx.someScribe(w, r, unsubscribeTemplate, "Have a good one!")

	// Update the unsubscribe status
	if (listMember != nil) {
		now := time.Now()
		listMember.Unsubscribed = &now
		ctx.db.Update(listMember)
		// todo: send an email to unsubscribers to let them they're off
	}
}
