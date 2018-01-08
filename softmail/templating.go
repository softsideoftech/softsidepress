package softmail

import (
	txtTemplate "text/template"
	htmlTemplate "html/template"
	"io"
	"bytes"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"sync"
	"regexp"
)

type HtmlPageParams struct {
	Url        string
	Title      string
	Summary    string
	Css        string
	JavaScript string
	Body       string
}

type BodyParams interface {
}

type MarkdownTemplateConfig struct {
	Writer       io.Writer
	BaseHtmlFile string
	Url          string
	Title        string
	Summary      string
	MarkdownFile string
	BodyParams   BodyParams
}

// Load the css file
const cssFile = "src/softside/html/style.css"    // TODO: make this a relative path
const jsFile = "src/softside/html/javascript.js" // TODO: make this a relative path
var extractTitle = regexp.MustCompile("^# (.+)")
var templateCache = sync.Map{}

func renderMarkdownToHtmlTemplate(c MarkdownTemplateConfig) error {

	templateName := c.BaseHtmlFile + c.MarkdownFile

	// Get the template from the cache to avoid constantly reading and parsing files from disk
	fullPageTemplate, cacheLoaded := templateCache.Load(templateName)

	// todo: set to FALSE to turn off caching for development purposes
	//cacheLoaded = false

	if !cacheLoaded {
		// Load the markdown template file
		markdownTemplateBytes, err := ioutil.ReadFile(c.MarkdownFile)
		if err != nil {
			return err
		}

		// If a title wasn't provided, get it from the H1 from the markdown file.
		if (len(c.Title) == 0) {
			submatch := extractTitle.FindStringSubmatch(string(markdownTemplateBytes))
			if (submatch != nil) {
				c.Title = submatch[1]
			}
		}

		// Render the markdown as HTML
		bodyHtml := string(blackfriday.Run(markdownTemplateBytes))

		// Merge the markdown html into the base
		baseHtmlTemplate, err := txtTemplate.ParseFiles(c.BaseHtmlFile)
		if err != nil {
			return err
		}

		// Load the css file
		cssFileBytes, err := ioutil.ReadFile(cssFile)
		if err != nil {
			return err
		}

		// Load the js file
		jsFileBytes, err := ioutil.ReadFile(jsFile)
		if err != nil {
			return err
		}

		// Generate the html template by templatizing the page components and body html arguments into it
		buffer := &bytes.Buffer{}
		err = baseHtmlTemplate.Execute(buffer, HtmlPageParams{
			Url:        c.Url,
			Title:      c.Title,
			Summary:    c.Summary,
			Css:        string(cssFileBytes),
			JavaScript: string(jsFileBytes),
			Body:       bodyHtml,
		})
		if (err != nil) {
			return err
		}

		// Parse the rendered HTML as a new template
		// Using an html template instead of text to safely escape user-passed arguments
		fullPageTemplate, err = htmlTemplate.New(templateName).Parse(buffer.String())
		if err != nil {
			return err
		}
		templateCache.Store(templateName, fullPageTemplate)
	}

	template := fullPageTemplate.(*htmlTemplate.Template)
	template.Execute(c.Writer, c.BodyParams)

	return nil
}
