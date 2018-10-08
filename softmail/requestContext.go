package softmail

import (
	"github.com/go-pg/pg"
	"net/http"
	"os"
)

var SoftsideDB = pg.Connect(&pg.Options{
	User:     os.Getenv("SOFTSIDE_DB_USER"),
	Database: os.Getenv("SOFTSIDE_DB"),
	Password: os.Getenv("SOFTSIDE_DB_PASSWORD"),
	Addr:     os.Getenv("SOFTSIDE_DB_ADDRESS"),
})

var softsideContentPath = os.Getenv("SOFTSIDE_CONTENT")
var developmentMode = os.Getenv("SOFTSIDE_DEV_MODE") == "true"

var mdTemplateError = "/mgmt-pages/error.md"
var mdTemplateLogin = "/mgmt-pages/login.md"
var mdTemplateMessage = "/mgmt-pages/message.md"
var emailTemplateLoginLink = "/emails/login-link.md"
var blogPageHtmlTemplate = "/html/pages-tmpl.html"
var coursePageHtmlTemplate = "/html/course-tmpl.html"
var homePageHtmlTemplate = "/html/home-page-tmpl.html"
var homePageMdTemplate = "/pages/purposeful-leadership-coaching.md"
var mgmtPagesHtmlTemplate = "/html/mgmt-pages-tmpl.html"

// TODO: Make these configurable
const ownerFirstName = "Vlad"
const ownerLastName = "Giverts"
const ownerEmail = "vlad@softsideoftech.com"
const siteName = "Soft Side of Tech"
const siteDomain = "softsideoftech.com"
const trackingImageUrl = "https://d235962hz41e70.cloudfront.net/bear-100.png"
const trackingPixelUrl = "https://d235962hz41e70.cloudfront.net/transparent-pixel.png"
const FavIconUrl = "https://d235962hz41e70.cloudfront.net/favicon.ico"

type RequestContext struct {
	DB           *pg.DB
	W            http.ResponseWriter
	R            *http.Request
	ContentPath  string
	DevMode      bool
	MemberCookie *MemberCookie
	ListMember   *ListMember
}

func NewRawRequestCtx() *RequestContext {
	return NewRequestCtx(nil, nil, false)
}
func NewRequestCtx(w http.ResponseWriter, r *http.Request, initCtx bool) *RequestContext {
	ctx := &RequestContext{
		DB:          SoftsideDB,
		W:           w,
		R:           r,
		ContentPath: softsideContentPath,
		DevMode:     developmentMode,
	}
	
	//if initCtx {
	//	ctx.InitMemberCookie(0)
	//
	//	// If we have a ListMemberId but the ListMember hasn't been retrieved yet, then retrieve it now.
	//	if ctx.MemberCookie.ListMemberId != 0 && ctx.ListMember == nil {
	//		listMember := ListMember{Id: ctx.MemberCookie.ListMemberId}
	//		err := ctx.DB.Select(listMember)
	//		log.Printf("ERROR retrieving ListMember while initializing RequestContext: %v", err)
	//		ctx.ListMember = &listMember
	//	}
	//}
	
	return ctx
}

func (ctx RequestContext) FileExists(relativePath string) bool {
	fileInfo, _ := os.Stat(ctx.GetFilePath(relativePath))
	return fileInfo != nil
}

func (ctx RequestContext) GetFilePath(relativePath string) string {
	return ctx.ContentPath + relativePath
}