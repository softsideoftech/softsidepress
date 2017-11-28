package softmail

import (
	"net/http"
	"regexp"
	"encoding/base64"
	"fmt"
	"github.com/go-pg/pg"
	"time"
)

var extractSentEmailIdFromImgPixel = regexp.MustCompile("/(.*?)\\.jpg")
var extractIpAddress = regexp.MustCompile("(.*?):")

const encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+_"

var URLEncoding = base64.NewEncoding(encodeURL).WithPadding(base64.NoPadding)

const cookieName = "sftml"

// TODO: make this configurable
const trackingDomain = "softsideoftech.com"

func EncodeId(sentEmailId uint32) string {
	var buf [4]byte
	buf[0] = byte(sentEmailId >> 24)
	buf[1] = byte(sentEmailId >> 16)
	buf[2] = byte(sentEmailId >> 8)
	buf[3] = byte(sentEmailId)
	return URLEncoding.EncodeToString(buf[:])
}

func DecodeId(encodedSentEmailId string) (uint32, error) {
	buf, err := URLEncoding.DecodeString(encodedSentEmailId)
	if (err != nil) {
		return 0, err
	}
	sentEmailId := uint32(0)
	sentEmailId |= uint32(buf[0]) << 24
	sentEmailId |= uint32(buf[1]) << 16
	sentEmailId |= uint32(buf[2]) << 8
	sentEmailId |= uint32(buf[3])
	return sentEmailId, nil
}

func HandleEmailOpen(w http.ResponseWriter, r *http.Request) {

	// TODO: Replace the naive DB connection with connection pooling and a config driven connection string
	ctx := &RequestContext{
		db: pg.Connect(&pg.Options{
			User: "vlad",
		}),
	}

	// Obtain the SentEmailId from the url path
	sentEmailId := decodeSendMailIdFromUri(r.URL.Path)

	// Get the encoded cookie id from http request
	httpCookie, err := r.Cookie(cookieName)

	// Look up the cookie from the database and create a new record if it doesn't already exist
	memberCookie := ctx.obtainOrCreateMemberCookie(sentEmailId, httpCookie)

	// Set the cookie on the HTTP request. It might already exist, so we'll simply be refreshing the MaxAge.
	http.SetCookie(w, &http.Cookie{
		Domain: trackingDomain,
		MaxAge: 60 * 60 * 24 * 365, // 1 year
		Name:   cookieName,
		Value:  EncodeId(memberCookie.Id),
	})

	// Track the hit
	trackingHit := TrackingHit{
		LinkId:         1, // Default linkId for email tracking pixel
		MemberCookieId: memberCookie.Id,
		ReferrerUrl:    r.Referer(),
		IpAddress:      decodeIpAddress(r.RemoteAddr),

		// Use the cookie as the authority on the ListMemberId because it would
		// have either been initialized correctly or retrieved from the db
		ListMemberId: memberCookie.ListMemberId,
	}
	err = ctx.db.Insert(&trackingHit)
	if (err != nil) {
		fmt.Printf("Problem inserting TrackingHit record. ListMemberId: : %d, Remote IP Address: %s, ReferrerURL: %s DB message: %s\n", memberCookie.ListMemberId, r.RemoteAddr, r.Referer(), err)
	}
}
func decodeSendMailIdFromUri(path string) uint32 {
	submatch := extractSentEmailIdFromImgPixel.FindStringSubmatch(path)
	if (submatch == nil) {
		return 0
	}
	sentEmailId, err := DecodeId(submatch[1])
	if (err != nil) {
		fmt.Printf("Problem parsing SentEmailId from url: %s, error message: %s", path, err)
		// Keep going anyway so we could set the cookie and retroactive track this user later if we obtain a ListMemberId
	}
	return sentEmailId
}

func decodeIpAddress(remoteAddr string) string {
	submatch := extractIpAddress.FindStringSubmatch(remoteAddr)
	if (submatch == nil) {
		return ""
	} else {
		return submatch[1]
	}
}


func (ctx RequestContext) obtainOrCreateMemberCookie(sentEmailId uint32, httpCookie *http.Cookie) *MemberCookie {
	// Get the sent email from the db so we could find the list_member_id
	var listMemberId uint32 = 0
	sentEmail := SentEmail{Id: sentEmailId}
	err := ctx.db.Select(&sentEmail)
	if (err != nil) {
		fmt.Printf("Problem retrieving SentEmail record with id: : %d, DB message: %s\n", sentEmailId, err)
	} else {
		listMemberId = sentEmail.ListMemberId
	}

	// Retrieve the MemberCookie if it exists
	var memberCookie *MemberCookie = nil
	if (httpCookie != nil) {
		encodedCookieId := httpCookie.Value
		memberCookieId, err := DecodeId(encodedCookieId)
		if (err != nil) {
			fmt.Printf("Problem parsing MemberCookieId from httpCookie: %s, error message: %s", encodedCookieId, err)
		} else {
			memberCookie = &MemberCookie{Id: memberCookieId}
			err := ctx.db.Select(memberCookie)
			if (err != nil) {
				fmt.Printf("Problem retrieving MemberCookie record with id: : %d, DB message: %s\n", memberCookieId, err)
			}
		}
	}
	if (memberCookie != nil) {
		if (listMemberId == 0) {
			// If we didn't find a ListMemberId in the url and we have a httpCookie, then initialize the id from the httpCookie
			listMemberId = memberCookie.ListMemberId

		} else if (memberCookie.ListMemberId == 0 || memberCookie.ListMemberId != listMemberId) {
			// If we already had a httpCookie that didn't have a list member id or the member id has changed, then update the httpCookie
			memberCookie.ListMemberId = listMemberId
			memberCookie.Updated = time.Now()
			err := ctx.db.Update(memberCookie)
			if (err != nil) {
				fmt.Printf("Problem updating MemberCookie record with id: %d, ListMemberId: %d, DB message: %s\n", memberCookie.Id, listMemberId, err)
			}
		}
	} else {
		// Sicne we don't have a MemberCookie yet, then create one
		memberCookie = &MemberCookie{ListMemberId: listMemberId}
		err := ctx.db.Insert(memberCookie)
		if (err != nil) {
			fmt.Printf("Problem inserting MemberCookie record for ListMemberId: : %d, DB message: %s\n", listMemberId, err)
		}
	}
	return memberCookie
}
