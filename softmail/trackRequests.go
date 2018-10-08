package softmail

import (
	"log"
	"net/http"
	"regexp"
	"encoding/base64"
	"fmt"
	"time"
	"crypto/md5"
	"hash/fnv"
	"strings"
	"math/rand"
	"strconv"
)

var extractSentEmailIdFromImgPixel = regexp.MustCompile(".*/(.+?).png")
var extractSentEmailIdFromUrlEndDash = regexp.MustCompile("(.*)-(.+?$)")
var extractIpAddress = regexp.MustCompile("(.*?):")

const encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+_"

var URLEncoding = base64.NewEncoding(encodeURL).WithPadding(base64.NoPadding)

const cookieName = "sftml"


type TrackingRequestParams struct {
	TrackingId TrackingHitId
}

// TODO: Move to util?
func EncodeId(sentEmailId uint32) string {
	var buf [4]byte
	buf[0] = byte(sentEmailId >> 24)
	buf[1] = byte(sentEmailId >> 16)
	buf[2] = byte(sentEmailId >> 8)
	buf[3] = byte(sentEmailId)
	str := URLEncoding.EncodeToString(buf[:])

	// Strip out the leading A's so we have nice short ids in the url
	str = strings.TrimLeft(str, "A")
	return str
}

func DecodeId(idString string) (uint32, error) {

	// Add back the leading A's that were stripped out in EncodeId().
	if len(idString) < 6 {
		idString = strings.Repeat("A", 6-len(idString)) + idString
	}

	buf, err := URLEncoding.DecodeString(idString)
	if err != nil {
		return 0, err
	}
	decodedId := uint32(0)
	decodedId |= uint32(buf[0]) << 24
	decodedId |= uint32(buf[1]) << 16
	decodedId |= uint32(buf[2]) << 8
	decodedId |= uint32(buf[3])
	return decodedId, nil
}

func UrlToId(url string) TrackedUrlId {
	hash := md5.New()
	hash.Write([]byte(url))
	md5Sum := hash.Sum(nil)
	hash64 := fnv.New64()
	hash64.Write(md5Sum)
	return int64(hash64.Sum64()) // make it signed to conform with the Postgres "bigint" type
}

func DecodeMemberCookieId(idString string) (MemberCookieId, error) {
	decodedId, err := DecodeId(idString)
	return MemberCookieId(decodedId), err
}

func GenerateTrackingLink(ctx *RequestContext) {
	targetUrl := ctx.R.URL.Query().Get("target")

	// Keep trying until we create a new short url
	url := ""
	for len(url) == 0 {
		curUrl, err := ctx.TryToCreateShortTrackedUrl(targetUrl, 0, false)

		if err != nil {
			panic(fmt.Errorf("Failed to generate tracking url for link: %s, err: $v", targetUrl, err))
			http.Error(ctx.W, "Something went wrong generating the link.", http.StatusInternalServerError)
			return
		}

		url = curUrl
	}

	fmt.Fprintf(ctx.W, "<a href='%s'>%s</a>", url, url)
}

// TODO: use this method to replace external links in emails
func (ctx *RequestContext) TryToCreateShortTrackedUrl(targetUrl string, sentEmailId SentEmailId, login bool) (string, error) {
	// Randomly generate a url
	url := "/" + EncodeId(rand.Uint32())
	trackedUrl := &TrackedUrl{Id: UrlToId(url), SentEmailId: sentEmailId, Login: login}

	err := ctx.DB.Select(trackedUrl)
	if err != nil {
		// Only continue if we didn't collide with an existing url.
		if IsPgSelectEmpty(err) {
			trackedUrl.Url = url
			trackedUrl.TargetUrl = targetUrl
			err := ctx.DB.Insert(trackedUrl)
			if err != nil {
				panic(fmt.Errorf("failed to insert TrackedUrl: %s, err: %v", trackedUrl.Url, err))
				return "", err
			}

			return url, nil
		} else {
			return "", err
		}
	} else {
		// If we got here, we must have collided with another url, so try again.
		return ctx.TryToCreateShortTrackedUrl(targetUrl, sentEmailId, login)
	}
	return "", nil
}

func TrackTimeOnPage(ctx *RequestContext) {
	trackingHitIdStr := ctx.R.URL.Query()["id"][0]
	parsedId, err := strconv.ParseInt(trackingHitIdStr, 10, 32)
	if err != nil {
		fmt.Printf("Counld't parse tracking id: %s, err: %v", trackingHitIdStr, err)
	}
	var trackingHit TrackingHit
	_, err = ctx.DB.Model(&trackingHit).Set("time_on_page = time_on_page + 5").Where("id = ?", TrackingHitId(parsedId)).Update()
	if err != nil {
		fmt.Printf("Counldn't update tracking hit with id: %d, err: %v", parsedId, err)
	}
}

func HandleNormalRequest(ctx *RequestContext) {
	rawRemoteAddr, ipString, ipInt := getIpInfo(ctx.R)

	var trackingHitId TrackingHitId
	trackedUrl := &TrackedUrl{Id: 0} // Default linkId for tracking requests

	// Try to match a page assuming there is no tracking code in the url.
	pageTemplateCfg := ctx.matchWebPage(trackingHitId, false)

	var sentEmailId SentEmailId
	var emailTargetUrl *string

	// If we couldn't find a page with the bare url, that means the url may contain a tracking id.
	urlPath := ctx.R.URL.Path
	if pageTemplateCfg == nil {
		sentEmailId, emailTargetUrl = DecodeSentMailIdFromUri(urlPath)
	}

	var sentEmail *SentEmail = nil

	// Don't don't bother with cookies for local requests (healthchecks, etc)
	if ipString != "127.0.0.1" {

		// Try to obtain the ListMemberId using the encoded SendEmailId in the url path if it exists.
		var err error
		sentEmail, err = ctx.GetSentEmail(sentEmailId)
		// ignore any error and keep going

		// Get the cookie from the request so we could look it up in the
		// database and create a new record if it doesn't already exist
		ctx.InitMemberCookie(sentEmail)

		// Obtain or save the url if it's not an email tracking pixel
		if !strings.HasSuffix(urlPath, ".png") {
			trackedUrl, err = ctx.obtainOrCreateTrackedUrl(urlPath)
		}

		// Only track if we didn't have any errors (most likely db)
		if err == nil {

			// Track the hit
			trackingHit := TrackingHit{
				TrackedUrlId:    trackedUrl.Id,
				MemberCookieId:  ctx.MemberCookie.Id,
				ReferrerUrl:     ctx.R.Referer(),
				IpAddress:       ipInt,
				IpAddressString: ipString,
				// Use the cookie as the authority on the ListMemberId because it would
				// have either been initialized correctly or retrieved from the db
				ListMemberId: ctx.MemberCookie.ListMemberId,
			}
			err = ctx.DB.Insert(&trackingHit)
			if err != nil {
				err = fmt.Errorf(
					"Problem inserting TrackingHit record. ListMemberId: : %d, Remote IP Address: %s, ReferrerURL: %s DB error: %v\n",
					ctx.MemberCookie.ListMemberId, rawRemoteAddr, ctx.R.Referer(), err)
			} else {
				trackingHitId = trackingHit.Id
			}
		}

		// Log any error, but keep trying to serve the page
		if err != nil {
			fmt.Println(err)
		}
	}

	if ctx.serveRedirect(trackedUrl, sentEmail, emailTargetUrl, urlPath) {
		return;
	} else if pageTemplateCfg == nil {
		// We're not doing a redirect, so it must be a web page. 
		// Default to the home page if no page is matched.
		pageTemplateCfg = ctx.matchWebPage(trackingHitId, true)
	}

	// Try to render the page
	err := ctx.renderMarkdownToHtmlTemplate(pageTemplateCfg)

	if err != nil {
		ctx.SendUserFacingError("", err)
	}
}

func (ctx *RequestContext) serveRedirect(trackedUrl *TrackedUrl, sentEmail *SentEmail, emailTargetUrl *string, urlPath string) bool {

	// If the request is a redirect-tracking link, then redirect the request now.
	// It's possible that trackedUrl will be nil if we had a error (db most likely)
	if trackedUrl != nil && len(trackedUrl.TargetUrl) > 0 {
		http.Redirect(ctx.W, ctx.R, trackedUrl.TargetUrl, http.StatusTemporaryRedirect)
		return true
	}

	// If this was an email link to an internal page, then redirect to it now
	if sentEmail != nil && sentEmail.ListMemberId != 0 && emailTargetUrl != nil {
		ctx.DB.Insert(&EmailAction{SentEmailId: sentEmail.Id, Action: "clicked", Metadata: *emailTargetUrl})
		http.Redirect(ctx.W, ctx.R, *emailTargetUrl, http.StatusTemporaryRedirect)
		return true
	}

	// For now, assume all png requests are email tracking so serve up the tracking image
	if strings.HasSuffix(urlPath, ".png") {
		ctx.DB.Insert(&EmailAction{SentEmailId: sentEmail.Id, Action: "opened"})
		var trackingUrl string
		if strings.Contains(urlPath, trackingPixelPath) {
			trackingUrl = trackingPixelUrl
		} else {
			trackingUrl = trackingImageUrl
		}
		http.Redirect(ctx.W, ctx.R, trackingUrl, http.StatusTemporaryRedirect)
		return true
	}

	return false
}

func (ctx *RequestContext) matchWebPage(trackingHitId TrackingHitId, defaultToHome bool) *MarkdownTemplateConfig {

	// Build the escapedUrl for pages to potentially use
	var escapedUrl = fmt.Sprintf("https://%s%s", siteDomain, ctx.R.URL.EscapedPath())

	var cfg MarkdownTemplateConfig

	// Use the URL path as the summary
	urlPath := ctx.R.URL.Path
	pathDirName := strings.Trim(urlPath, "/")
	cfg.HtmlSummary = strings.Join(strings.Split(pathDirName, "-"), " ")
	trackingParams := TrackingRequestParams{TrackingId: trackingHitId}
	cfg.PerRequestParams = trackingParams
	cfg.Url = escapedUrl

	// Look for the possible page types
	if ctx.FileExists(ctx.getBlogPagePath(urlPath)) {
		cfg.MarkdownFile = ctx.getBlogPagePath(urlPath)
		cfg.BaseHtmlFile = blogPageHtmlTemplate

	} else if ctx.FileExists(ctx.getCourseDescriptionPath(urlPath)) {
		cfg.MarkdownFile = ctx.getCourseDescriptionPath(urlPath)
		cfg.BaseHtmlFile = coursePageHtmlTemplate
		courseParams, err := ctx.GetCoursePageParams(pathDirName, trackingParams)

		// If we get an error here, it means the user must log in to view this page.
		if err != nil {
			switch err.(type) {
			case NotLoggedInError:
				// Render the login page 
				cfg.BaseHtmlFile = mgmtPagesHtmlTemplate
				cfg.MarkdownFile = mdTemplateLogin
				cfg.HtmlTitle = "Please Login First"
				cfg.PerRequestParams = MdMessageParams(courseParams.Name)
			case NoSuchCourseError:
				// Let the user know this course doesn't exist  
				cfg.BaseHtmlFile = mgmtPagesHtmlTemplate
				cfg.MarkdownFile = mdTemplateMessage
				cfg.HtmlTitle = "Course Not Found"
				cfg.PerRequestParams = MdMessageParams(fmt.Sprintf("I'm sorry. I couldn't find a course with the name '%s'.", pathDirName))
			case NotRegisteredForCourseError:
				
			case CourseNotStartedError:
				// Let the user know this course doesn't exist 
				cfg.BaseHtmlFile = mgmtPagesHtmlTemplate
				cfg.MarkdownFile = mdTemplateMessage
				cfg.HtmlTitle = "Course Hasn't Started Yet"
				startDateStr := err.(CourseNotStartedError).StartDate.Format("Jan 2, 2006")
				cfg.PerRequestParams = MdMessageParams(fmt.Sprintf("I'm sorry. It looks like this course hasn't started yet. Check back in on the start date: %s.", startDateStr))
			}
		} else {
			cfg.PerRequestParams = courseParams
		}

	} else if defaultToHome {
		// Always default to the home page
		cfg.MarkdownFile = homePageMdTemplate
		cfg.BaseHtmlFile = homePageHtmlTemplate
		cfg.HtmlTitle = siteName
	} else {
		// If we didn't match one of the pages and we're not defaulting to the home page, then return nil.
		return nil
	}

	return &cfg
}

func (ctx *RequestContext) getBlogPagePath(urlPath string) string {
	return "/pages" + urlPath + ".md"
}

func (ctx *RequestContext) getCourseDescriptionPath(urlPath string) string {
	return "/courses" + urlPath + "/content.md"
}

func getIpInfo(r *http.Request) (string, string, IpAddress) {
	// Try to find the user's IP address in the request
	var rawRemoteAddr string
	realIp := r.Header.Get("X-Real-IP")
	if len(realIp) == 0 {
		realIp = r.Header.Get("X-Forwarded-For")
	}
	if len(realIp) > 0 {
		rawRemoteAddr = realIp
	} else {
		rawRemoteAddr = r.RemoteAddr
	}
	ipString, ipInt := decodeIpAddress(rawRemoteAddr)
	return rawRemoteAddr, ipString, ipInt
}

func (ctx *RequestContext) obtainOrCreateTrackedUrl(urlPath string) (*TrackedUrl, error) {
	trackedUrl := &TrackedUrl{Id: UrlToId(urlPath)}
	err := ctx.DB.Select(trackedUrl)
	if err != nil {
		// todo: refactor checking for no results in select
		if IsPgSelectEmpty(err) {
			trackedUrl.Url = urlPath
			err = ctx.DB.Insert(trackedUrl)
			if err != nil {
				return nil, fmt.Errorf("failed to insert tracked url: %s, err: %v", urlPath, err)
			}
		} else {
			return nil, fmt.Errorf("failed to select from db TrackedUrlIUd: %d, err: %v", trackedUrl.Id, err)
		}
	}
	return trackedUrl, nil
}

func DecodeSentMailIdFromUri(path string) (SentEmailId, *string) {
	var targetLink *string
	submatch := extractSentEmailIdFromImgPixel.FindStringSubmatch(path)
	var sentEmailId = SentEmailId(0)
	var err error

	// Try to get the sentEmailId assuming this is a tracking pixel
	if submatch != nil {
		sentEmailId, err = DecodeId(submatch[1])
	} else {
		// Otherwise, assume this is an email link to a site page and try to get the id from there
		submatch = extractSentEmailIdFromUrlEndDash.FindStringSubmatch(path)
		if submatch == nil {
			return 0, nil
		}
		targetLink = &(submatch[1])
		sentEmailId, err = DecodeId(submatch[2])
	}
	if err != nil {
		log.Printf("Problem parsing SentEmailId from url: %s, error message: %v\n", path, err)
		// Keep going anyway so we could set the cookie and retroactive track this user later if we obtain a ListMemberId
	}
	return sentEmailId, targetLink
}

func decodeIpAddress(remoteAddr string) (string, IpAddress) {

	var ipAddressString string
	if strings.Index(remoteAddr, ":") == -1 {
		ipAddressString = remoteAddr
	} else if strings.HasPrefix(remoteAddr, "[::1]") {
		// The decoder only understands IP V4.
		ipAddressString = "127.0.0.1"
	} else {
		submatch := extractIpAddress.FindStringSubmatch(remoteAddr)
		if submatch == nil {
			return "", 0
		}
		ipAddressString = submatch[1]
	}

	// todo: make the IP decoding logic understand IPV6
	parts := strings.Split(ipAddressString, ".")
	firstOctet, err1 := strconv.ParseInt(parts[0], 10, 64)
	secondOctet, err2 := strconv.ParseInt(parts[1], 10, 64)
	thirdOctet, err3 := strconv.ParseInt(parts[2], 10, 64)
	fourthOctet, err4 := strconv.ParseInt(parts[3], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		fmt.Printf("Problem parsing IP Address: %s, error: %v, %v, %v, %v", remoteAddr, err1, err2, err3, err4)
		return "", 0
	}
	return ipAddressString, (firstOctet * 16777216) + (secondOctet * 65536) + (thirdOctet * 256) + (fourthOctet)
}

func (ctx *RequestContext) InitMemberCookie(sentEmail *SentEmail) {
	httpCookie, err := ctx.R.Cookie(cookieName)

	// See if we have a ListMemberId. If not, default to 0
	var listMemberId ListMemberId = 0
	if sentEmail != nil {
		listMemberId = sentEmail.ListMemberId
	}
	
	// Create a MemberCookie in the db to link the tracking back to the ListMember and set the browser cookie in case it wasn't already set
	ctx.MemberCookie = ctx.ObtainOrCreateMemberCookie(listMemberId, httpCookie)

	if err == nil {
		// Set the cookie on the HTTP request. It might already exist, so we'll simply be refreshing the MaxAge.
		http.SetCookie(ctx.W, &http.Cookie{
			Domain: siteDomain,
			MaxAge: 60 * 60 * 24 * 365, // 1 year
			Name:   cookieName,
			Value:  EncodeId(uint32(ctx.MemberCookie.Id)),
		})
	} else {
		fmt.Printf("Problem obtaining or creating MemberCookie for listMemberId: %d, err: %v", listMemberId, err)
	}
}

/*
Here's how we handle cookies:

If there's no httpCookie or dbCookie, then we create and return an httpCookie.
If there's an httpCookie but no dbCookie and a listMemberId is present, then we create the dbCookie so we could relate past activity to a listMember

Prefer the passed in listMemberId over the one in the dbCookie
   This way if someone get's a forwarded email from their friend and get assigned their friend's listMemberId when they click a link,
   it will get overriden either when they sign up or click a link in an email meant for them at a later time

If a listMemberId is present and either the httpCookie or dbCookie exist but don't have it set, then update them with the listMemberId.

*/
func (ctx *RequestContext) ObtainOrCreateMemberCookie(listMemberId ListMemberId, httpCookie *http.Cookie) *MemberCookie {
	if httpCookie == nil {
		// If there's no httpCookie, then randomly generate an id and return a new one. We're taking a small chance of a collision, but that's ok.
		randomCookieId := MemberCookieId(rand.Uint64())
		memberCookie := &MemberCookie{Id: randomCookieId, ListMemberId: listMemberId}

		// Only save the cookie in the db if a listMemberId is present
		if (listMemberId != 0) {
			ctx.DB.Insert(memberCookie)
		}
		return memberCookie
	} else {
		// Since we have an httpCookie, try to retrieve the dbCookie if it exists
		encodedCookieId := httpCookie.Value
		memberCookieId, err := DecodeMemberCookieId(encodedCookieId)
		memberCookie := &MemberCookie{Id: memberCookieId}
		if err != nil {
			fmt.Printf("Problem parsing MemberCookieId from httpCookie: %s, error message: %s", encodedCookieId, err)
		} else {
			err := ctx.DB.Select(memberCookie)
			if err != nil {
				if IsPgSelectEmpty(err) {
					// Since we didn't find a cookie in the db, save it if a listMemberId is present
					if (listMemberId != 0) {
						memberCookie.ListMemberId = listMemberId
						err = ctx.DB.Insert(memberCookie)
						if err != nil {
							fmt.Printf("Problem inserting MemberCookie record with id: : %d, DB message: %s\n", memberCookieId, err)
						}
					}
				} else {
					fmt.Printf("Problem retrieving MemberCookie record with id: : %d, DB message: %s\n", memberCookieId, err)
				}
			} else {
				// If we got here, it means we have an httpCookie and we retrieved the corresponding dbCookie
				if listMemberId == 0 {
					// If we didn't find a ListMemberId in the url and we have a httpCookie, then initialize the id from the dbCookie
					listMemberId = memberCookie.ListMemberId

				} else if memberCookie.ListMemberId == 0 || memberCookie.ListMemberId != listMemberId {
					// If we already had a httpCookie that didn't have a list member id or the member id has changed, then update the httpCookie
					memberCookie.ListMemberId = listMemberId
					memberCookie.Updated = time.Now()
					err := ctx.DB.Update(memberCookie)
					if err != nil {
						fmt.Printf("Problem updating MemberCookie record with id: %d, ListMemberId: %d, DB message: %s\n", memberCookie.Id, listMemberId, err)
					}
				}
			}
		}
		return memberCookie
	}
}
