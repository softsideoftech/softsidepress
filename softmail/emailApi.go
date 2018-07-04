package softmail

import "net/http"

func GetEmailInfo(w http.ResponseWriter, r *http.Request) {
	//recipient := r.URL.Query().Get("recipient")
	////subject := r.URL.Query().Get("subject")
	//
	//ctx := NewRequestCtx(w, r)
	//
	//"select t.subject, s.id as sent_email_id, s.created as email_sent, a.action, a.created as action_time from list_members l, email_templates t, email_actions a, sent_emails s where t.id = s.email_template_id and s.list_member_id = l.id and a.sent_email_id = s.id order by s.id desc limit 20;"

}
