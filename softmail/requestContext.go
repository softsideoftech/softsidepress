package softmail

import (
	"github.com/go-pg/pg"
	"net/http"
)

var SoftsideDB = pg.Connect(&pg.Options{User: "softside",})

type RequestContext struct {
	db *pg.DB
	w http.ResponseWriter
	r *http.Request
}

func NewRequestCtx(w http.ResponseWriter, r *http.Request) *RequestContext {
	return &RequestContext{db: SoftsideDB, w: w, r: r}
}
