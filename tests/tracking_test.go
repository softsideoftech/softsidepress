package softside

import (
	"softside/softmail"
	"fmt"
	"testing"
)

func testIdEncoding() {
	//nums := []uint{0, 1, 2, 3, 4, 32432423142314321}
	//sum := 0
	//for _, num := range nums {
	//
	//}
}

func TestDecodeSentMailLinkFromUri(t *testing.T) {
	var rawSentMailId = "sentmailid+_"
	var targetLink = "https://www.foobar.com/my-fancy-shmancy-article"
	actualId, _ := softmail.DecodeId(rawSentMailId)
	decodedID, decodedTargetLink := softmail.DecodeSentMailIdFromUri(targetLink + "-" + rawSentMailId)
	if (decodedID != actualId) {
		t.Error(fmt.Sprintf("decodedID != sentMailId --- %d != %d", decodedID, actualId))
	}
	if (*decodedTargetLink != targetLink) {
		t.Error(fmt.Sprintf("decodedTargetLink != targetLink --- %v != %v", decodedTargetLink, targetLink))
	}
}


func TestDecodeTrackingPixel(t *testing.T) {
	var rawSentMailId = "sentmailid+_"
	actualId, _ := softmail.DecodeId(rawSentMailId)
	decodedID, decodedTargetLink := softmail.DecodeSentMailIdFromUri("https://www.foobar.com/bear/" + rawSentMailId + ".png")
	if (decodedID != actualId) {
		t.Error(fmt.Sprintf("decodedID != sentMailId --- %d != %d", decodedID, actualId))
	}
	if (decodedTargetLink != nil) {
		t.Error(fmt.Sprintf("decodedTargetLink not nil as expected: %v", decodedTargetLink))
	}
}

