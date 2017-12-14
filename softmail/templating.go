package softmail

import (
	"text/template"
	"io"
	"bytes"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
)

type HtmlPageParams struct {
	Title string
	Css string
	Body string
}

type BodyParams interface {

}

// Load the css file
const cssFile = "/Users/vlad/go/src/softside/style.css" // TODO: make this a relative path
var cssFileBytes, err = ioutil.ReadFile(cssFile)
var cssFileString = string(cssFileBytes)


// TODO: figure out how to load and cache all the template files
func renderMarkdownToHtmlTemplate(writer io.Writer, templateFile string, title string, bodyFile string, bodyParams BodyParams) error {

	// Load the markdown template file
	markdownTemplateBytes, err := ioutil.ReadFile(bodyFile)
	if err != nil {
		return err
	}

	// Render the markdown as HTML
	bodyHtml := string(blackfriday.Run(markdownTemplateBytes))

	// Templatize the markdown body
	bodyMarkdownTemplate, err := template.New(bodyFile).Parse(bodyHtml)
	if err != nil {
		return err
	}
	buffer := &bytes.Buffer{}
	bodyMarkdownTemplate.Execute(buffer, bodyParams)

	// Write out the html
	htmlTemplate, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}
	htmlTemplate.Execute(writer, HtmlPageParams{Title: title, Css: cssFileString, Body: buffer.String()})

	return nil
}
