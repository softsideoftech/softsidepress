package softside

import (
	"math/rand"
	"softside/softmail"
	"fmt"
	"testing"
)

func TestDecodeSentMailLinkFromUri(t *testing.T) {
	var rawSentMailId = "sentmailid+_"
	var targetLink = "https://www.foobar.com/my-fancy-shmancy-article"
	actualId, _ := softmail.DecodeSentEmailId(rawSentMailId)
	decodedID, decodedTargetLink := softmail.DecodeSentMailIdFromUri(targetLink + "-" + rawSentMailId)
	if (decodedID != actualId) {
		t.Error(fmt.Sprintf("decodedID != sentMailId --- %d != %d", decodedID, actualId))
	}
	if (*decodedTargetLink != targetLink) {
		t.Error(fmt.Sprintf("decodedTargetLink != targetLink --- %v != %v", decodedTargetLink, targetLink))
	}
}
func TestEncodeCookieId(t *testing.T) {
	cookieId := softmail.MemberCookieId(rand.Uint64())
	idString := softmail.EncodeCookieId(cookieId)
	decodedCookieId, err := softmail.DecodeCookieId(idString)
	
	if err != nil {
		t.Error(err)
	}
	if len(idString) < 1 {
		t.Errorf("Bad id string: %s", idString)
	}
	
	if cookieId != decodedCookieId {
		t.Errorf("Didn't encode/decode cookie correctly. Original: %d, Decoded: %d", cookieId, decodedCookieId)
	}
}


func TestDecodeTrackingPixel(t *testing.T) {
	var rawSentMailId = "sentmailid+_"
	actualId, _ := softmail.DecodeSentEmailId(rawSentMailId)
	decodedID, decodedTargetLink := softmail.DecodeSentMailIdFromUri("https://www.foobar.com/bear/" + rawSentMailId + ".png")
	if (decodedID != actualId) {
		t.Error(fmt.Sprintf("decodedID != sentMailId --- %d != %d", decodedID, actualId))
	}
	if (decodedTargetLink != nil) {
		t.Error(fmt.Sprintf("decodedTargetLink not nil as expected: %v", decodedTargetLink))
	}
}

