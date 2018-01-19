package main

import (
	"softside/softmail"
	"fmt"
)

func main() {
	testDecodeSentMailLinkFromUri()
	testDecodeTrackingPixel()
}

func testIdEncoding() {
	//nums := []uint{0, 1, 2, 3, 4, 32432423142314321}
	//sum := 0
	//for _, num := range nums {
	//
	//}
}

func testDecodeSentMailLinkFromUri() {
	var rawSentMailId = "sentmailid+_"
	var targetLink = "https://www.foobar.com/my-fancy-shmancy-article"
	actualId, _ := softmail.DecodeId(rawSentMailId)
	decodedID, decodedTargetLink := softmail.DecodeSentMailIdFromUri(targetLink + "-" + rawSentMailId)
	if (decodedID != actualId) {
		panic(fmt.Sprintf("decodedID != sentMailId --- %d != %d", decodedID, actualId))
	}
	if (*decodedTargetLink != targetLink) {
		panic(fmt.Sprintf("decodedTargetLink != targetLink --- %s != %s", decodedTargetLink, targetLink))
	}
}


func testDecodeTrackingPixel() {
	var rawSentMailId = "sentmailid+_"
	actualId, _ := softmail.DecodeId(rawSentMailId)
	decodedID, decodedTargetLink := softmail.DecodeSentMailIdFromUri("https://www.foobar.com/bear/" + rawSentMailId + ".png")
	if (decodedID != actualId) {
		panic(fmt.Sprintf("decodedID != sentMailId --- %d != %d", decodedID, actualId))
	}
	if (decodedTargetLink != nil) {
		panic(fmt.Sprintf("decodedTargetLink not nil as expected: ", decodedTargetLink))
	}
}