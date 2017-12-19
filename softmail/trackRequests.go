package softmail

import (
	"net/http"
	"regexp"
	"encoding/base64"
	"fmt"
	"github.com/go-pg/pg"
	"time"
	"crypto/md5"
	"hash/fnv"
	"strings"
	"math/rand"
	"os"
)

var extractSentEmailIdFromImgPixel = regexp.MustCompile("/(.*?)\\.jpg")
var extractIpAddress = regexp.MustCompile("(.*?):")

const encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+_"

var URLEncoding = base64.NewEncoding(encodeURL).WithPadding(base64.NoPadding)

const cookieName = "sftml"

// TODO: make this configurable
const siteDomain = "softsideoftech.com"

// TODO: Move to util
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
	// TODO: Replace the naive DB connection with connection pooling and a config driven connection string
	ctx := &RequestContext{
		db: pg.Connect(&pg.Options{
			User: "vlad",
		}),
	}

	// Keep trying until we create a new short url
	url := ""
	for len(url) == 0 {
		curUrl, err := ctx.tryToCreateShortTrackedUrl(targetUrl)

		if err != nil {
			panic(fmt.Errorf("Failed to generate tracking url for link: %s, err: $v", targetUrl, err))
			http.Error(w, "Something went wrong generating the link.", http.StatusInternalServerError)
			return
		}

		url = curUrl
	}

	fmt.Fprintf(w, "<a href='%s'>%s</a>", url, url)
}
func (ctx *RequestContext) tryToCreateShortTrackedUrl(targetUrl string) (string, error) {
	// Randomly generate a url
	url := "/" + EncodeId(rand.Uint32())
	trackedUrl := &TrackedUrl{Id: UrlToId(url)}

	// Check if we accidentally generated an existing url
	err := ctx.db.Select(trackedUrl)
	// todo: refactor checking for no results in select
	if err != nil {
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
	}
	return "", nil
}

func TrackRequest(w http.ResponseWriter, r *http.Request) {

	// TODO: Replace the naive DB connection with connection pooling and a config driven connection string
	ctx := &RequestContext{
		db: pg.Connect(&pg.Options{
			User: "vlad",
		}),
	}

	// Try to obtain the ListMemberId using the encoded SendEmailId in the url path if it exists.
	sentEmailId := decodeSendMailIdFromUri(r.URL.Path)
	listMemberId, err := ctx.getListMemberIdFromSentEmail(sentEmailId)
	// ignore any error and keep going

	// Get the cookie from the request so we could look it up in the
	// database and create a new record if it doesn't already exist
	httpCookie, err := r.Cookie(cookieName)
	memberCookie := ctx.obtainOrCreateMemberCookie(listMemberId, httpCookie)

	// Set the cookie on the HTTP request. It might already exist, so we'll simply be refreshing the MaxAge.
	http.SetCookie(w, &http.Cookie{
		Domain: siteDomain,
		MaxAge: 60 * 60 * 24 * 365, // 1 year
		Name:   cookieName,
		Value:  EncodeId(uint32(memberCookie.Id)),
	})

	// Obtain or save the url if it's not an email tracking pixel
	trackedUrl := ctx.obtainOrCreateTrackedUrl(sentEmailId, r.URL.Path)

	// Track the hit
	trackingHit := TrackingHit{
		TrackedUrlId:   trackedUrl.Id,
		MemberCookieId: memberCookie.Id,
		ReferrerUrl:    r.Referer(),
		IpAddress:      decodeIpAddress(r.RemoteAddr),

		// Use the cookie as the authority on the ListMemberId because it would
		// have either been initialized correctly or retrieved from the db
		ListMemberId: memberCookie.ListMemberId,
	}
	err = ctx.db.Insert(&trackingHit)
	if err != nil {
		panic(fmt.Errorf("Problem inserting TrackingHit record. ListMemberId: : %d, Remote IP Address: %s, ReferrerURL: %s DB error: %v\n", memberCookie.ListMemberId, r.RemoteAddr, r.Referer(), err))
	}

	// If the request is a redirect-tracking link, then redirect the request now
	if len(trackedUrl.TargetUrl) > 0 {
		http.Redirect(w, r, trackedUrl.TargetUrl, http.StatusTemporaryRedirect)

	} else {
		// Otherwise this might be a website page, so look for that.
		templateFile := "src/softside/pages" + r.URL.Path + ".md"
		fileInfo, err := os.Stat(templateFile)

		// Check if we should load a regular page or the home page
		if fileInfo != nil && !strings.Contains(templateFile, "index.html") {
			words := strings.Split(fileInfo.Name(), "-")
			title := strings.Title(strings.Join(words, " "))
			err = renderMarkdownToHtmlTemplate(w, pagesHtmlTemplate, title, templateFile, nil)
		} else {
			err = renderMarkdownToHtmlTemplate(w, homePageHtmlTemplate, "Soft Side of Tech", homePageMdTemplate, nil)
		}
		if err != nil {
			sendUserFacingError("", err, w)
		}

	}
}

func (ctx *RequestContext) obtainOrCreateTrackedUrl(sentEmailId SentEmailId, urlPath string) *TrackedUrl {
	trackedUrl := &TrackedUrl{Id: 0}
	// Default linkId for email tracking pixel
	if sentEmailId == 0 {
		trackedUrl.Id = UrlToId(urlPath)
		err := ctx.db.Select(trackedUrl)
		if err != nil {
			// todo: refactor checking for no results in select
			if strings.Contains(err.Error(), "no rows in result set") {
				trackedUrl.Url = urlPath
				err = ctx.db.Insert(trackedUrl)
				if err != nil {
					panic(fmt.Errorf("failed to insert tracked url: %s, err: %v", urlPath, err))
				}
			} else {
				panic(fmt.Errorf("failed to select from db TrackedUrlIUd: %d, err: %v", trackedUrl.Id, err))
			}
		}
	}
	return trackedUrl
}

func decodeSendMailIdFromUri(path string) SentEmailId {
	submatch := extractSentEmailIdFromImgPixel.FindStringSubmatch(path)
	if submatch == nil {
		return 0
	}
	sentEmailId, err := DecodeId(submatch[1])
	if err != nil {
		fmt.Printf("Problem parsing SentEmailId from url: %s, error message: %s", path, err)
		// Keep going anyway so we could set the cookie and retroactive track this user later if we obtain a ListMemberId
	}
	return sentEmailId
}

func decodeIpAddress(remoteAddr string) string {
	submatch := extractIpAddress.FindStringSubmatch(remoteAddr)
	if submatch == nil {
		return ""
	} else {
		return submatch[1]
	}
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

func (ctx *RequestContext) obtainOrCreateMemberCookie(listMemberId ListMemberId, httpCookie *http.Cookie) *MemberCookie {
	// Retrieve the MemberCookie if it exists
	var memberCookie *MemberCookie = nil
	if httpCookie != nil {
		encodedCookieId := httpCookie.Value
		memberCookieId, err := DecodeMemberCookieId(encodedCookieId)
		if err != nil {
			fmt.Printf("Problem parsing MemberCookieId from httpCookie: %s, error message: %s", encodedCookieId, err)
		} else {
			memberCookie = &MemberCookie{Id: memberCookieId}
			err := ctx.db.Select(memberCookie)
			if err != nil {
				fmt.Printf("Problem retrieving MemberCookie record with id: : %d, DB message: %s\n", memberCookieId, err)
			}
		}
	}
	if memberCookie != nil {
		if listMemberId == 0 {
			// If we didn't find a ListMemberId in the url and we have a httpCookie, then initialize the id from the httpCookie
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
	} else {
		// Sicne we don't have a MemberCookie yet we should create one
		memberCookie = &MemberCookie{ListMemberId: listMemberId}
		err := ctx.db.Insert(memberCookie)
		if err != nil {
			fmt.Printf("Problem inserting MemberCookie record for ListMemberId: : %d, DB message: %s\n", listMemberId, err)
		}
	}
	return memberCookie
}
