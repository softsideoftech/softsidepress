package softside

import (
	"softside/softmail"
	"fmt"
	"testing"
	"database/sql"
	"log"
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

func TestDb(t *testing.T) {
	connStr := "user=softside dbname=softside sslmode=disable password=local_dev_passwordasdf"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT * FROM list_members limit 60")

	if err != nil {
		t.Errorf("Error: %v", err)
	}
	println("HELLO WORLD!!!")


	cols, err := rows.Columns()
	if err != nil {
		fmt.Println("Failed to get columns", err)
		return
	}

	// Result is your slice string.
	rawResult := make([][]byte, len(cols))
	result := make([]string, len(cols))

	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			fmt.Println("Failed to scan row", err)
			return
		}

		for i, raw := range rawResult {
			if raw == nil {
				result[i] = "\\N"
			} else {
				result[i] = string(raw)
			}
		}

		fmt.Printf("%#v\n", result)
	}
}