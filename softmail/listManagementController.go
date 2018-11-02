package softmail

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)


func (ctx *RequestContext) SendUserFacingError(logMessage string, err interface{}) {
	log.Printf(logMessage + ", error: %v\n\n", err)
	ctx.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx.W.Header().Set("X-Content-Type-Options", "nosniff")
	ctx.W.WriteHeader(http.StatusInternalServerError)

	ctx.renderMarkdownToHtmlTemplate(&MarkdownTemplateConfig{
		BaseHtmlFile:     mgmtPagesHtmlTemplate,
		HtmlTitle:        "Something isn't right...",
		MarkdownFile:     mdTemplateError,
		PerRequestParams: MdMessageParams(""),
	})
}

func (ctx *RequestContext) someScribe(templateFile string, pageTitle string) *ListMember {

	// Find the SentEmailId in the url
	sentEmailId := decodeSendMailIdFromUriEnd(ctx.R.URL.Path)
	if sentEmailId == 0 {
		ctx.SendUserFacingError("Couldn't find SentEmailId in url: %v", nil)
		return nil
	}

	// Load the SentEmail from the DB
	sentEmail, err := ctx.GetSentEmail(sentEmailId)
	if (err != nil) {
		ctx.SendUserFacingError("", err)
		return nil
	}
	listMember := &ListMember{Id: sentEmail.ListMemberId}
	err = ctx.DB.Select(listMember)
	if err != nil {
		ctx.SendUserFacingError("Couldn't find list member in url: %v", err)
		return nil
	}

	ctx.renderMgmtPage(templateFile, pageTitle, sentEmailId, listMember, "")

	return listMember
}

func (ctx *RequestContext) renderMgmtPage(templateName string, pageTitle string, sentEmailId SentEmailId, listMember *ListMember, message string) {

	ctx.W.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Run the template
	listMemberParams := CommonMdTemplateParams{
		listMember.FirstName,
		listMember.Email,
		EncodeId(sentEmailId),
		ownerFirstName,
		ownerEmail,
		siteName,
		message,
		pageTitle,
	}
	config := MarkdownTemplateConfig{
		BaseHtmlFile:     mgmtPagesHtmlTemplate,
		HtmlTitle:        pageTitle,
		MarkdownFile:     "/mgmt-pages/" + templateName + ".md",
		PerRequestParams: listMemberParams,
	}
	err := ctx.renderMarkdownToHtmlTemplate(&config)
	
	if err != nil {
		panic(fmt.Sprintf("Failed to render management page: %s, error: %v", config.MarkdownFile, err))
	}
}

func RequestLoginLink(ctx *RequestContext) {
	email := ctx.R.FormValue("email")
	title := ctx.R.FormValue("title")

	if email == "" {
		ctx.SendUserFacingError("ERROR: No email specified for RequestLoginLink", nil)
		return
	}

	listMembers := ctx.SendTemplatedEmail(
		"Login link for " + siteName, 
		emailTemplateLoginLink, 
		email, 
		SendEmailOpts{Login: true, DestinationUrl: ctx.R.Referer(), PageTitle: title})

	ctx.renderMgmtPage("login-link-sent", "Check Your Email", 0, &listMembers[0], "")
}

func Resubscribe(ctx *RequestContext) {
	listMember := ctx.someScribe("resubscribe", "Welcome back :)")

	// Update the unsubscribe status
	if (listMember != nil) {
		listMember.Unsubscribed = nil
		ctx.DB.Update(listMember)
	}
}

func Unsubscribe(ctx *RequestContext) {
	listMember := ctx.someScribe("unsubscribe", "Goodbye, {{.FirstName}}")

	// Update the unsubscribe status
	if (listMember != nil) {
		now := time.Now()
		listMember.Unsubscribed = &now
		ctx.DB.Update(listMember)
		// todo: send an email to unsubscribers to let them know they're off
	}
}

func Join(ctx *RequestContext) {
	firstName := ctx.R.FormValue("first-name")
	email := ctx.R.FormValue("email")

	// todo: validate the input params 

	listMember, _, err := ctx.GetListMemberByEmail(email)

	// If it's an error other than no rows returned, then log it and keep going
	if (err != nil) {
		log.Printf("ERROR selecting member from list. FirstName: %s, Email: %s, err: %v", firstName, email, err)
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
		err = ctx.DB.Update(listMember)
	} else {
		err = ctx.DB.Insert(listMember)

		// Also record this ListMember's location
		_, _, ipAddress := ctx.GetIpInfo()
		ctx.DB.Exec("insert into list_member_locations (select ?,  ip.country_code, ip.country_name, ip.region_name, ip.city_name, ip.time_zone from ip2location ip where ? >= ip_from and ? <= ip_to)", listMember.Id, ipAddress, ipAddress)
	}

	if (err != nil) {
		ctx.SendUserFacingError(fmt.Sprintf("Problem adding member to list. FirstName: %s, Email: %s", firstName, email), err)
	} else {
		ctx.InitMemberCookie(listMember.Id)
		ctx.renderMgmtPage("join", "Welcome, {{.FirstName}}", SentEmailId(0), listMember, "")

		// todo: send a confirmation/double-opt-in email
	}
}

func (ctx RequestContext) GetIpInfo() (string, string, IpAddress) {
	// Try to find the user's IP address in the request
	var rawRemoteAddr string
	realIp := ctx.R.Header.Get("X-Real-IP")
	if len(realIp) == 0 {
		realIp = ctx.R.Header.Get("X-Forwarded-For")
	}
	if len(realIp) > 0 {
		rawRemoteAddr = realIp
	} else {
		rawRemoteAddr = ctx.R.RemoteAddr
	}
	ipString, ipInt := decodeIpAddress(rawRemoteAddr)
	return rawRemoteAddr, ipString, ipInt
}
