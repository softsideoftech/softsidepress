-- Email Opens
select  t.id, t.created, count(*) cnt from sent_emails s, email_actions a, email_templates t where a.sent_email_id = s.id and t.id = email_template_id and action = 'opened' group by t.id, t.created   order by cnt desc;

-- Visits & Time per page
select url, count(*), sum(time_on_page) from tracked_urls u, tracking_hits h where u.id = h.tracked_url_id and time_on_page > 15 group by url order by count desc;

-- Time per word on page (for people who stayed for more than 5 min)
select w.url, count(*) as reads, sum(time_on_page) as seconds,  round((1000.0 * sum(time_on_page)/count(*))/wordcount) as time_per_visit_per_word
from tracked_urls u, tracking_hits h,
(values 
  ('/motivation-at-clara', 3331),
  ('/the-safety-bubble', 1880),
  ('/startup-humiliation', 3401),
  ('/personality-parts', 1104),
  ('/purposeful-leadership-coaching', 636)
) as w (url, wordcount) 
where u.id = h.tracked_url_id and u.url = w.url and time_on_page > 300 group by w.url, wordcount order by count(*) desc;


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

-- Locations of all list members
with user_ips as (select list_member_id, ip_address from tracking_hits where list_member_id is not null),
    user_ips2 as (select list_member_id, mode() within group (order by ip_address)
                  from user_ips
                  group by list_member_id)
select list_member_id, ip.country_code, ip.country_name, ip.region_name, ip.city_name, ip.time_zone from user_ips2 u, ip2location ip where u.mode >= ip.ip_from and u.mode <= ip_to;