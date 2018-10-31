package softmail

import (
	"bytes"
	"gopkg.in/russross/blackfriday.v2"
	htmlTemplate "html/template"
	"io/ioutil"
	"regexp"
	"sync"
	txtTemplate "text/template"
)

type HtmlPageParams struct {
	Url        string
	Title      string
	Summary    string
	Css        string
	JavaScript string
	Body       string
	Request    PerRequestParams
}

type PerRequestParams interface {
}

type MarkdownTemplateConfig struct {
	BaseHtmlFile     string
	Url              string
	HtmlTitle        string
	HtmlSummary      string
	MarkdownFile     string
	PerRequestParams PerRequestParams
}

type CommonMdTemplateParams struct {
	FirstName  string
	Email      string
	EncodedId  string
	OwnerName  string
	OwnerEmail string
	SiteName   string
	Message    string
}

var cssFile = "/html/style.css"
var jsFile = "/html/javascript.js"
var extractTitle = regexp.MustCompile("^# (.+)")
var templateCache = sync.Map{}

func (ctx *RequestContext) renderMarkdownToHtmlTemplate(c *MarkdownTemplateConfig) error {
	// Use the combo of the html file and md files to uniquely identify the template.
	templateName := c.BaseHtmlFile + c.MarkdownFile

	// Get the template from the cache to avoid constantly reading and parsing files from disk
	fullPageTemplate, cacheLoaded := templateCache.Load(templateName)

	if !cacheLoaded || ctx.DevMode {
		// Load the markdown template file
		markdownTemplateBytes, err := ioutil.ReadFile(ctx.GetFilePath(c.MarkdownFile))
		if err != nil {
			return err
		}

		// If a title wasn't provided, get it from the H1 from the markdown file.
		if len(c.HtmlTitle) == 0 {
			submatch := extractTitle.FindStringSubmatch(string(markdownTemplateBytes))
			if submatch != nil {
				c.HtmlTitle = submatch[1]
			}
		}

		// Render the markdown as HTML
		bodyHtml := string(blackfriday.Run(markdownTemplateBytes))

		// Merge the markdown html into the base
		baseHtmlTemplate, err := txtTemplate.ParseFiles(ctx.GetFilePath(c.BaseHtmlFile))
		if err != nil {
			return err
		}

		// Load the css file
		cssFileBytes, err := ioutil.ReadFile(ctx.GetFilePath(cssFile))
		if err != nil {
			return err
		}

		// Load the js file
		jsFileBytes, err := ioutil.ReadFile(ctx.GetFilePath(jsFile))
		if err != nil {
			return err
		}

		// Generate the html template by templatizing the page components and body html arguments into it
		buffer := &bytes.Buffer{}
		params := HtmlPageParams{
			Url:        c.Url,
			Title:      c.HtmlTitle,
			Summary:    c.HtmlSummary,
			Css:        string(cssFileBytes),
			JavaScript: string(jsFileBytes),
			Body:       bodyHtml,
			Request:    c.PerRequestParams,
		}
		err = baseHtmlTemplate.Execute(buffer, params)
		if err != nil {
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
	return template.Execute(ctx.W, c.PerRequestParams)
}

func MdMessageParams(message string) *CommonMdTemplateParams {
	return &CommonMdTemplateParams{
		OwnerName:  ownerFirstName,
		OwnerEmail: ownerEmail,
		SiteName:   siteName,
		Message:    message,
	}
}
