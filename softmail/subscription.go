package softmail

import (
	"time"
	"bytes"
	"net/http"
	"fmt"
	"strings"
)

const errorTemplate = "src/softside/mgmt-pages/error.md"
const pagesHtmlTemplate = "src/softside/html/pages-tmpl.html"
const homePageHtmlTemplate = "src/softside/html/home-page-tmpl.html"
const homePageMdTemplate = "src/softside/pages/whats-purposeful-leadership-coaching.md"
const mgmtPagesHtmlTemplate = "src/softside/html/mgmt-pages-tmpl.html"
const owner = "Vlad"

type SubscriptionTemplateParams struct {
	FirstName string
	EncodedId string
	OwnerName string
}

func sendUserFacingError(logMessage string, err error, w http.ResponseWriter) {
	fmt.Printf(logMessage+"err: %v", err)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
	renderMarkdownToHtmlTemplate(w, mgmtPagesHtmlTemplate, "Something isn't right...", errorTemplate, SubscriptionTemplateParams{OwnerName: owner})
}

func (ctx *RequestContext) someScribe(templateFile string, pageTitle string) *ListMember {

	// Find the SentEmailId in the url
	sentEmailId := decodeSendMailIdFromUriEnd(ctx.r.URL.Path)
	if sentEmailId == 0 {
		sendUserFacingError("Couldn't find SentEmailId in url: %v", nil, ctx.w)
		return nil
	}

	// Load the ListMember from the DB
	listMemberId, err := ctx.getListMemberIdFromSentEmail(sentEmailId)
	if (err != nil) {
		sendUserFacingError("", err, ctx.w)
		return nil
	}
	listMember := &ListMember{Id: listMemberId}
	err = ctx.db.Select(listMember)
	if err != nil {
		sendUserFacingError("Couldn't find list member in url: %v", err, ctx.w)
		return nil
	}

	renderMgmtPage(ctx.w, ctx.r, templateFile, pageTitle, sentEmailId, listMember)

	return listMember
}

func renderMgmtPage(w http.ResponseWriter, r *http.Request, templateName string, pageTitle string, sentEmailId SentEmailId, listMember *ListMember) {
	// Run the template
	buffer := &bytes.Buffer{}
	listMemberParams := SubscriptionTemplateParams{listMember.FirstName, EncodeId(sentEmailId), owner}
	renderMarkdownToHtmlTemplate(buffer, mgmtPagesHtmlTemplate, pageTitle, "src/softside/mgmt-pages/"+templateName+".md", listMemberParams)

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(buffer.Bytes()))
}

func Resubscribe(w http.ResponseWriter, r *http.Request) {
	ctx := NewRequestCtx(w, r)

	listMember := ctx.someScribe("resubscribe", "Welcome back :)")

	// Update the unsubscribe status
	if (listMember != nil) {
		listMember.Unsubscribed = nil
		ctx.db.Update(listMember)
	}
}

func Unsubscribe(w http.ResponseWriter, r *http.Request) {
	// Connect to the DB
	ctx := NewRequestCtx(w, r)

	listMember := ctx.someScribe("unsubscribe", "Goodbye, {{.FirstName}}")

	// Update the unsubscribe status
	if (listMember != nil) {
		now := time.Now()
		listMember.Unsubscribed = &now
		ctx.db.Update(listMember)
		// todo: send an email to unsubscribers to let them know they're off
	}
}

func Join(w http.ResponseWriter, r *http.Request) {
	ctx := NewRequestCtx(w, r)

	firstName := r.FormValue("first-name")
	email := r.FormValue("email")

	// todo: validate the input params

	listMember := &ListMember{}
	err := ctx.db.Model(listMember).Column("list_member.*").Where("list_member.email = ?", email).Select()

	// If it's an error other than no rows returned, then log it
	if (err != nil && !strings.Contains(err.Error(), "no rows in result set")) {
		fmt.Printf("Selecting member from list. FirstName: %s, Email: %s, err: %v", firstName, email, err)
	}

	// Update all the fields whether or not the record exists. Updating email is idempotent.
	listMember.Email = email
	listMember.FirstName = strings.Title(strings.ToLower(firstName))
	now := time.Now()
	if (listMember.Subscribed == nil) {
		listMember.Subscribed = &now
	}
	if (listMember.Unsubscribed != nil) {
		listMember.Unsubscribed = nil
	}
	listMember.Updated = now

	if (listMember.Id > 0) {
		err = ctx.db.Update(listMember)
	} else {
		err = ctx.db.Insert(listMember)
	}

	if (err != nil) {
		sendUserFacingError(fmt.Sprintf("Problem adding member to list. FirstName: %s, Email: %s", firstName, email), err, w)
	} else {
		renderMgmtPage(w, r, "join", "Welcome, {{.FirstName}}", SentEmailId(0), listMember)
	}
}
