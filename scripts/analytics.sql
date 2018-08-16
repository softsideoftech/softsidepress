-- Email Opens
select  t.id, t.created, count(*) cnt from sent_emails s, email_actions a, email_templates t where a.sent_email_id = s.id and t.id = email_template_id and action = 'opened' group by t.id, t.created   order by cnt desc;

-- Visits & Time per page
select url, count(*), sum(time_on_page) from tracked_urls u, tracking_hits h where u.id = h.tracked_url_id group by url order by count desc;

-- Signups Per Month
select date_part('month', subscribed) as month, count(*) from list_members where subscribed is not null group by date_part('month', subscribed);

-- Unsubscribes per month
select date_part('month', unsubscribed) as month, count(*) from list_members where unsubscribed is not null group by date_part('month', unsubscribed);

-- Recent email opens
with opens as (
select t.id, to_char(min(a.created), 'YYYY-MM-DD HH24:MI') open_date from email_templates t, sent_emails s, email_actions a where s.email_template_id = t.id and a.sent_email_id = s.id and a.action = 'opened' group by t.id order by t.created desc),
rest as (select t.id, t.subject, t.created from email_templates t, sent_emails s, email_actions a where s.email_template_id = t.id and a.sent_email_id = s.id group by t.id, t.subject, t.created)
select subject,to_char(created, 'YYYY-MM-DD HH24:MI') sent, open_date from opens right join rest on opens.id = rest.id order by created desc limit 50;


with asd as 
(select t.id, min(a.created) open_date from email_templates t, sent_emails s, email_actions a where s.email_template_id = t.id and a.sent_email_id = s.id and a.action = 'opened' group by t.id order by t.created desc limit 50)
select * from asd;