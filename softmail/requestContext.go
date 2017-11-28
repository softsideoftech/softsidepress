package softmail

import "github.com/go-pg/pg"

type RequestContext struct {
	db *pg.DB
}
