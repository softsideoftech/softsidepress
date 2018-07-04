package softmail

import "strings"

func GetListMemberByEmail(email string) (*ListMember, bool, error) {
	listMember := &ListMember{}
	err := SoftsideDB.Model(listMember).Column("list_member.*").Where("list_member.email = ?", email).Select()
	if IsPgSelectEmpty(err) {
		return listMember, false, nil
	} else {
		return listMember, listMember.Id > 0, err
	}
}

func IsPgSelectEmpty(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no rows in result set")
}
