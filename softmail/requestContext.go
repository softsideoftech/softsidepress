package softmail

import (
	"github.com/go-pg/pg"
	"net/http"
	"os"
)

var SoftsideDB = pg.Connect(&pg.Options{
	User: os.Getenv("SOFTSIDE_DB_USER"),
	Database: os.Getenv("SOFTSIDE_DB"),
	Password: os.Getenv("SOFTSIDE_DB_PASSWORD"),
	Addr: os.Getenv("SOFTSIDE_DB_ADDRESS"),
	})

var SoftsideContentPath = os.Getenv("SOFTSIDE_CONTENT")
var DevelopmentMode = os.Getenv("SOFTSIDE_DEV_MODE") == "true"

type RequestContext struct {
	db *pg.DB
	w http.ResponseWriter
	r *http.Request
}

func NewRequestCtx(w http.ResponseWriter, r *http.Request) *RequestContext {
	return &RequestContext{db: SoftsideDB, w: w, r: r}
}
