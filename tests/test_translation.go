package softside

import (
	"testing"
	"softside/softmail"
	"softside/tests/sampleTranslationHtml"
	"fmt"
	"log"
)

// This is meant to be run manually because the environment needs to have the translation service credentials.
func TestTranslationParsing(t *testing.T) {
	sourceText := softmail.ParseTextFromHtml(sampleTranslationHtml.TranslationSample)
	translate(sourceText, t)
	TestNonTranslation(t)
}

func TestNonTranslation(t *testing.T) {
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
