package softmail

import (
	txtTemplate "text/template"
	htmlTemplate "html/template"
	"io"
	"bytes"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"sync"
)

type HtmlPageParams struct {
	Url   string
	Title string
	Css   string
	Body  string
}

type BodyParams interface {
}

// Load the css file
const cssFile = "src/softside/style.css" // TODO: make this a relative path
var templateCache = sync.Map{}

func renderMarkdownToHtmlTemplate(writer io.Writer, baseHtmlFile string, url string, title string, markdownFile string, bodyParams BodyParams) error {
	templateName := baseHtmlFile + markdownFile

	// Get the template from the cache to avoid constantly reading and parsing files from disk
	fullPageTemplate, cacheLoaded := templateCache.Load(templateName)

	// todo: set to FALSE to turn off caching for development purposes
	cacheLoaded = false

	if !cacheLoaded {
		// Load the markdown template file
		markdownTemplateBytes, err := ioutil.ReadFile(markdownFile)
		if err != nil {
			return err
		}

		// Render the markdown as HTML
		bodyHtml := string(blackfriday.Run(markdownTemplateBytes))

		// Merge the markdown html into the base
		baseHtmlTemplate, err := txtTemplate.ParseFiles(baseHtmlFile)
		if err != nil {
			return err
		}
		buffer := &bytes.Buffer{}

		// Load the css file
		cssFileBytes, err := ioutil.ReadFile(cssFile)
		if err != nil {
			return err
		}
		var cssFileString = string(cssFileBytes)

		err = baseHtmlTemplate.Execute(buffer, HtmlPageParams{Url: url, Title: title, Css: cssFileString, Body: bodyHtml})
		if (err != nil) {
			return err
		}

		// Render the parameters into the full template.
		// Using an html template instead of text to safely escape user-passed params
		fullPageTemplate, err = htmlTemplate.New(templateName).Parse(buffer.String())
		if err != nil {
			return err
		}
		templateCache.Store(templateName, fullPageTemplate)
	}

	template := fullPageTemplate.(*htmlTemplate.Template)
	template.Execute(writer, bodyParams)

	return nil
}
