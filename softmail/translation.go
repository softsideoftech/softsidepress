package softmail

import (
	"regexp"
	"strings"
	"golang.org/x/text/language"
	"golang.org/x/net/context"
	"cloud.google.com/go/translate"
	"log"
	"math/rand"
)

var matchLeadingWhitespace = regexp.MustCompile(">(\\s+)")
var matchWhitespace = regexp.MustCompile("\\s+")
var matchSingleWhitespace = regexp.MustCompile(">\\s<")
var extractText = regexp.MustCompile(">([^<>]+)")

func ParseTextFromHtml(htmlStr string) []string {

	cleanedHtml := matchWhitespace.ReplaceAllString(htmlStr, " ")
	cleanedHtml = matchSingleWhitespace.ReplaceAllString(cleanedHtml, "><")
	matches := extractText.FindAllStringSubmatch(cleanedHtml, -1)

	toTranslate := make([]string, 0, len(matches)/2)
	for _, matchArray := range matches {
		toTranslate = append(toTranslate, strings.Trim(matchArray[1], " "))
	}

	return toTranslate
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func shuffle(src []string) []string {
	dest := make([]string, len(src))
	perm := rand.Perm(len(src))
	for i, v := range perm {
		dest[v] = src[i]
	}
	return dest
}

func TranslateText(sourceText []string) (map[string]string, error) {

	// TODO: Move this to a global context somewhere?
	// Initialize the translation api client
	ctx := context.Background()
	client, err := translate.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	
	
	// Call the DetectLanguage api to figure out what we're dealing with
	sampleSize := 15
	sliceLen := min(sampleSize, len(sourceText))
	sampleText := shuffle(sourceText)
	sourceTextSample := sampleText[0:sliceLen]
	detections, err := client.DetectLanguage(ctx, sourceTextSample)
	if err != nil {
		return nil, err
	}

	// Find the most common language in our sourceTextSample according to the api
	detectionMap := make(map[language.Tag]int)
	for _, detectionArr := range detections {
		for _, detection := range detectionArr {
			if detection.IsReliable || detection.Confidence >= 0.9 {
				detectionMap[detection.Language] = detectionMap[detection.Language] + 1 
			}
		}
	}
	
	// Set the source language
	maxSourceLanguage := language.English
	maxLangaugeCount := 0
	for sourceLangCandidate, sourceLangCount := range detectionMap {
		// Ignore Und (undetermined)
		if sourceLangCount > maxLangaugeCount && sourceLangCandidate != language.Und {
			maxLangaugeCount = sourceLangCount
			maxSourceLanguage = sourceLangCandidate
			log.Printf("sourceLangCandidate: %s, sourceLangCount: %d\n", sourceLangCandidate, sourceLangCount)
		}
	} 
	
	// Don't bother translating if the source language is English
	if maxSourceLanguage == language.English {
		return nil, nil
	}
	
	// Actually do the translation
	translations, err := client.Translate(ctx,
		sourceText, language.English,
		&translate.Options{
			Source: maxSourceLanguage,
			Format: translate.Text,
		})
	if err != nil {
		return nil, err
	}
	
	// Create a map from the source text to the translated text
	translatedText := make(map[string]string)
	for i, translation := range translations {
		translatedText[sourceText[i]] = translation.Text
	}
	
	return translatedText, nil
}

func ReplaceHtmlWithTranslation(htmlStr string, translationMap map[string]string) string {
	cleanedHtml := matchWhitespace.ReplaceAllString(htmlStr, " ")
	cleanedHtml = matchSingleWhitespace.ReplaceAllString(cleanedHtml, "><")

	return extractText.ReplaceAllStringFunc(cleanedHtml, func(str string) string {
		var prefix = ""
		var suffix = ""
		if strings.HasPrefix(str, "> ") {
			prefix = "&nbsp;"
		}
		if strings.HasSuffix(str, " ") {
			suffix = "&nbsp;"
		}
		str = strings.Trim(str, "> ")
		return ">" + prefix + translationMap[str] + suffix
	})
}

func TranslateHtml(htmlString string) (string, error) {
	sourceText := ParseTextFromHtml(htmlString)
	translations, err := TranslateText(sourceText)

	if err != nil {
		return "", err
	}
	if translations == nil || len(translations) == 0 {
		return "", nil
	}
	
	return ReplaceHtmlWithTranslation(htmlString, translations), nil
}
