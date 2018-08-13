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
	if len(sourceText) < 50 {
		t.Error("Failed to parse all the text from TranslationSample HTML")
	}
	

	translations, err := softmail.TranslateText(sourceText[0:10])
	
	if err != nil {
		t.Error(fmt.Sprintf("Problem translating text: %v", err))
	}
	
	log.Printf("translations: \n%v\n\n", translations)

	translation := softmail.ReplaceHtmlWithTranslation(sampleTranslationHtml.TranslationSample, translations)
	println(translation)
}
