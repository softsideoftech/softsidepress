-- Email Opens
select  t.id, t.created, count(*) cnt from sent_emails s, email_actions a, email_templates t where a.sent_email_id = s.id and t.id = email_template_id and action = 'opened' group by t.id, t.created   order by cnt desc;

-- Visits & Time per page
select url, count(*), sum(time_on_page) from tracked_urls u, tracking_hits h where u.id = h.tracked_url_id group by url order by count desc;

-- Signups Per Month
select date_part('month', subscribed) as month, count(*) from list_members where subscribed is not null group by date_part('month', subscribed);

-- Unsubscribes per month
select date_part('month', unsubscribed) as month, count(*) from list_members where unsubscribed is not null group by date_part('month', unsubscribed);