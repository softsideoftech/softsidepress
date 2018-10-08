package softmail

import (
	"fmt"
	"strings"
)

func (ctx RequestContext) GetListMemberByEmail(email string) (*ListMember, bool, error) {
	listMember := &ListMember{}
	err := ctx.DB.Model(listMember).Column("list_member.*").Where("list_member.email = ?", email).Select()
	if IsPgSelectEmpty(err) {
		return listMember, false, nil
	} else {
		return listMember, listMember.Id > 0, err
	}
}

func IsPgSelectEmpty(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no rows in result set")
}

func (ctx *RequestContext) GetSentEmail(sentEmailId SentEmailId) (*SentEmail, error) {
	if sentEmailId == 0 {
		return nil, nil
	}
	// Get the sent email from the db so we could find the list_member_id
	sentEmail := SentEmail{Id: sentEmailId}
	err := ctx.DB.Select(&sentEmail)
	if err != nil {
		return nil, fmt.Errorf("Problem retrieving SentEmail record with id: : %d, DB error: %s\v", sentEmailId, err)
	}
	return &sentEmail, nil
}
