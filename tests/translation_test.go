package softside

import (
	"testing"
	"softside/softmail"
	"softside/tests/sampleTranslationHtml"
	"fmt"
	"log"
	"softside/forwardEmail"
	"strings"
	"softside/tests/sampleEmails"
)

func TestTranslationParsing(t *testing.T) {
	sourceText := softmail.ParseTextFromHtml(sampleTranslationHtml.TranslationSample)
	print(sourceText)

	// This is meant to be run manually because the environment needs to have the translation service credentials.
	//doTranslation(sourceText, t)
}

func doTranslation(sourceText []string, t *testing.T) {
	translate(sourceText, t)
	testNonTranslation(t)
	user := forwardEmail.AnnonymousUser{}
	user.Send("vgiverts@gmail.com", []string{"test@mail.softsideoftech.com"}, strings.NewReader(sampleEmails.EmailSample))
}

func testNonTranslation(t *testing.T) {
	translationMap, err := softmail.TranslateText([]string{"the quick brown fox jumped over the lazy dog"})
	if err != nil {
		t.Error(fmt.Sprintf("Problem translating text: %v", err))
	}
	log.Printf("translationMap: \n%v\n\n", translationMap)
	translation := softmail.ReplaceHtmlWithTranslation(sampleTranslationHtml.TranslationSample, translationMap)
	println(translation)
}

func translate(sourceText []string, t *testing.T) {
	if len(sourceText) < 50 {
		t.Error("Failed to parse all the text from TranslationSample HTML")
	}
	translationMap, err := softmail.TranslateText(sourceText[0:10])

	if err != nil {
		t.Error(fmt.Sprintf("Problem translating text: %v", err))
	}
	log.Printf("translationMap: \n%v\n\n", translationMap)
	translation := softmail.ReplaceHtmlWithTranslation(sampleTranslationHtml.TranslationSample, translationMap)
	println(translation)
}
