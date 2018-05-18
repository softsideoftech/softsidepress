package softmail

import (
	"net/http"
	"regexp"
	"encoding/base64"
	"fmt"
	"time"
	"crypto/md5"
	"hash/fnv"
	"strings"
	"math/rand"
	"os"
	"strconv"
)

var extractSentEmailIdFromImgPixel = regexp.MustCompile(".*/(.+?).png")
var extractSentEmailIdFromUrlEndDash = regexp.MustCompile("(.*)-(.+?$)")
var extractIpAddress = regexp.MustCompile("(.*?):")

const encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+_"

var URLEncoding = base64.NewEncoding(encodeURL).WithPadding(base64.NoPadding)

const cookieName = "sftml"

// TODO: make these configurable
const siteDomain = "softsideoftech.com"
const trackingImageUrl = "https://d235962hz41e70.cloudfront.net/bear-100.png"
const FavIconUrl = "https://d235962hz41e70.cloudfront.net/favicon.ico"

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

func GenerateTrackingLink(w http.ResponseWriter, r *http.Request) {
	targetUrl := r.URL.Query().Get("target")

	ctx := NewRequestCtx(w, r)

	// Keep trying until we create a new short url
	url := ""
	for len(url) == 0 {
		curUrl, err := ctx.TryToCreateShortTrackedUrl(targetUrl, 0)

		if err != nil {
			panic(fmt.Errorf("Failed to generate tracking url for link: %s, err: $v", targetUrl, err))
			http.Error(w, "Something went wrong generating the link.", http.StatusInternalServerError)
			return
		}

		url = curUrl
	}

	fmt.Fprintf(w, "<a href='%s'>%s</a>", url, url)
}

// TODO: use this method to replace external links in emails
func (ctx *RequestContext) TryToCreateShortTrackedUrl(targetUrl string, sentEmailId SentEmailId) (string, error) {
	// Randomly generate a url
	url := "/" + EncodeId(rand.Uint32())
	trackedUrl := &TrackedUrl{Id: UrlToId(url), SentEmailId: sentEmailId}

	err := ctx.db.Select(trackedUrl)
	// todo: refactor checking for no results in select
	if err != nil {
		// Only continue if we didn't collide with an existing url.
		if strings.Contains(err.Error(), "no rows in result set") {
			trackedUrl.Url = url
			trackedUrl.TargetUrl = targetUrl
			err := ctx.db.Insert(trackedUrl)
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
		return ctx.TryToCreateShortTrackedUrl(targetUrl, sentEmailId)
	}
	return "", nil
}

func TrackTimeOnPage(w http.ResponseWriter, r *http.Request) {
	trackingHitIdStr := r.URL.Query()["id"][0]

	ctx := NewRequestCtx(w, r)

	parsedId, err := strconv.ParseInt(trackingHitIdStr, 10, 32)
	if err != nil {
		fmt.Printf("Counld't parse tracking id: %s, err: %v", trackingHitIdStr, err)
	}
	var trackingHit TrackingHit
	_, err = ctx.db.Model(&trackingHit).Set("time_on_page = time_on_page + 5").Where("id = ?", TrackingHitId(parsedId)).Update()
	if err != nil {
		fmt.Printf("Counldn't update tracking hit with id: %d, err: %v", parsedId, err)
	}
}

func HandleNormalRequest(w http.ResponseWriter, r *http.Request) {

	ctx := NewRequestCtx(w, r)

	rawRemoteAddr, ipString, ipInt := getIpInfo(r)

	var trackingHitId TrackingHitId
	trackedUrl := &TrackedUrl{Id: 0} // Default linkId for tracking requests

	urlPath := r.URL.Path
	sentEmailId, emailTargetUrl := DecodeSentMailIdFromUri(urlPath)
	var sentEmailListMemberId  = ListMemberId(0)

	// Don't don't bother with cookies for local requests (healthchecks, etc)
	if ipString != "127.0.0.1" {


		// Try to obtain the ListMemberId using the encoded SendEmailId in the url path if it exists.
		var err error
		sentEmailListMemberId, err = ctx.getListMemberIdFromSentEmail(sentEmailId)
		// ignore any error and keep going

		// Get the cookie from the request so we could look it up in the
		// database and create a new record if it doesn't already exist
		httpCookie, err := r.Cookie(cookieName)
		memberCookie := ctx.ObtainOrCreateMemberCookie(sentEmailListMemberId, httpCookie)

		// Set the cookie on the HTTP request. It might already exist, so we'll simply be refreshing the MaxAge.
		SetHttpCookie(w, memberCookie)

		// Obtain or save the url if it's not an email tracking pixel
		if !strings.HasSuffix(urlPath, ".png") {
			trackedUrl, err = ctx.obtainOrCreateTrackedUrl(urlPath)
		}

		// Only track if we didn't have any errors (most likely db)
		if err == nil {

			// Track the hit
			trackingHit := TrackingHit{
				TrackedUrlId:    trackedUrl.Id,
				MemberCookieId:  memberCookie.Id,
				ReferrerUrl:     r.Referer(),
				IpAddress:       ipInt,
				IpAddressString: ipString,
				// Use the cookie as the authority on the ListMemberId because it would
				// have either been initialized correctly or retrieved from the db
				ListMemberId: memberCookie.ListMemberId,
			}
			err = ctx.db.Insert(&trackingHit)
			if err != nil {
				err = fmt.Errorf(
					"Problem inserting TrackingHit record. ListMemberId: : %d, Remote IP Address: %s, ReferrerURL: %s DB error: %v\n",
					memberCookie.ListMemberId, rawRemoteAddr, r.Referer(), err)
			} else {
				trackingHitId = trackingHit.Id
			}
		}

		// Log any error, but keep trying to serve the page
		if err != nil {
			fmt.Println(err)
		}
	}

	// If the request is a redirect-tracking link, then redirect the request now.
	// It's possible that trackedUrl will be nil if we had a error (db most likely)
	if trackedUrl != nil && len(trackedUrl.TargetUrl) > 0 {
		http.Redirect(w, r, trackedUrl.TargetUrl, http.StatusTemporaryRedirect)
		return
	}

	// If this was an email link to an internal page, then redirect to it now
	if sentEmailListMemberId != 0 && emailTargetUrl != nil {
		ctx.db.Insert(&EmailAction{SentEmailId: sentEmailId, Action: "clicked", Metadata: *emailTargetUrl})
		http.Redirect(w, r, *emailTargetUrl, http.StatusTemporaryRedirect)
		return
	}

	// For now, assume all png requests are email tracking so serve up the tracking image
	if strings.HasSuffix(urlPath, ".png") {
		ctx.db.Insert(&EmailAction{SentEmailId: sentEmailId, Action: "opened"})
		http.Redirect(w, r, trackingImageUrl, http.StatusTemporaryRedirect)
		return
	}

	// Otherwise this might be a markdown page, so let's look for that.
	templateFile := SoftsideContentPath + "/pages" + urlPath + ".md"
	fileInfo, err := os.Stat(templateFile)

	// Build the escapedUrl for pages to potentially use
	var escapedUrl = fmt.Sprintf("https://%s%s", siteDomain, r.URL.EscapedPath())

	// Check if we should load a regular page or the home page
	if fileInfo != nil && !strings.Contains(templateFile, "index.html") {
		words := strings.Split(strings.Trim(urlPath, "/"), "-")
		summaryPhrase := strings.Join(words, " ")
		err = renderMarkdownToHtmlTemplate(MarkdownTemplateConfig{
			Writer:           w,
			BaseHtmlFile:     pagesHtmlTemplate,
			Url:              escapedUrl,
			Summary:          summaryPhrase,
			MarkdownFile:     templateFile,
			PerRequestParams: TrackingRequestParams{TrackingId: trackingHitId},
		})
	} else {
		// Didn't find a regular page so load the home page
		err = renderMarkdownToHtmlTemplate(
			MarkdownTemplateConfig{
				Writer:           w,
				BaseHtmlFile:     homePageHtmlTemplate,
				Url:              escapedUrl,
				Title:            "Soft Side of Tech",
				MarkdownFile:     homePageMdTemplate,
				PerRequestParams: TrackingRequestParams{TrackingId: trackingHitId},
			})
	}
	if err != nil {
		sendUserFacingError("", err, w)
	}
}
func SetHttpCookie(w http.ResponseWriter, memberCookie *MemberCookie) {
	http.SetCookie(w, &http.Cookie{
		Domain: siteDomain,
		MaxAge: 60 * 60 * 24 * 365, // 1 year
		Name:   cookieName,
		Value:  EncodeId(uint32(memberCookie.Id)),
	})
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
	err := ctx.db.Select(trackedUrl)
	if err != nil {
		// todo: refactor checking for no results in select
		if strings.Contains(err.Error(), "no rows in result set") {
			trackedUrl.Url = urlPath
			err = ctx.db.Insert(trackedUrl)
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
		fmt.Printf("Problem parsing SentEmailId from url: %s, error message: %v", path, err)
		// Keep going anyway so we could set the cookie and retroactive track this user later if we obtain a ListMemberId
	}
	return sentEmailId, targetLink
}

func decodeIpAddress(remoteAddr string) (string, IpAddress) {
	var ipAddressString string
	if strings.Index(remoteAddr, ":") == -1 {
		ipAddressString = remoteAddr
	} else {
		submatch := extractIpAddress.FindStringSubmatch(remoteAddr)
		if submatch == nil {
			return "", 0
		}
		ipAddressString = submatch[1]
	}

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

func (ctx *RequestContext) getListMemberIdFromSentEmail(sentEmailId SentEmailId) (ListMemberId, error) {
	if sentEmailId == 0 {
		return 0, nil
	}
	// Get the sent email from the db so we could find the list_member_id
	sentEmail := SentEmail{Id: sentEmailId}
	err := ctx.db.Select(&sentEmail)
	if err != nil {
		return 0, fmt.Errorf("Problem retrieving SentEmail record with id: : %d, DB error: %s\v", sentEmailId, err)
	}
	return sentEmail.ListMemberId, nil
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
			ctx.db.Insert(memberCookie)
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
			err := ctx.db.Select(memberCookie)
			if err != nil {
				if strings.Contains(err.Error(), "no rows in result set") {
					// Since we didn't find a cookie in the db, save it if a listMemberId is present
					if (listMemberId != 0) {
						memberCookie.ListMemberId = listMemberId
						err = ctx.db.Insert(memberCookie)
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
					err := ctx.db.Update(memberCookie)
					if err != nil {
						fmt.Printf("Problem updating MemberCookie record with id: %d, ListMemberId: %d, DB message: %s\n", memberCookie.Id, listMemberId, err)
					}
				}
			}
		}
		return memberCookie
	}
}
