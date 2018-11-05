
SELECT * FROM list_members;

SELECT * FROM member_groups;

update list_members set unsubscribed = now() where email = 'josh@giverts.com';


INSERT INTO tracked_urls (id, url) VALUES (0, '');
SELECT *
FROM tracked_urls;

INSERT INTO tracked_urls (id, url, target_url) VALUES (-2935145334901629174, '/asdf2', '"http://fakeurl.com/"');

SELECT *
FROM tracking_hits;

truncate tracking_hits;

INSERT INTO list_members (first_name, last_name, email) VALUES ('Mark', 'Johnson', 'vlad+3@cloudmars.com');

INSERT INTO email_actions_enum (action)
VALUES ('sent'), ('delivered'), ('opened'), ('clicked'), ('hard_bounce'), ('soft_bounce'), ('complaint');

SELECT *
FROM email_actions_enum;

SELECT *
FROM member_cookies order by created desc;
SELECT * FROM tracking_hits;
SELECT * FROM tracking_hits where time_on_page > 0;

INSERT INTO email_templates (subject, body) VALUES ('foo', 'bar'), ('asdf', 'qwer');

SELECT *
FROM sent_emails;

select * from email_actions;

delete from sent_emails where third_party_id is null;


-- new schema work
alter table list_members drop personal_role;
CREATE TABLE member_roles (
  id VARCHAR(128) PRIMARY KEY
);
INSERT INTO member_roles (id)
VALUES ('founder'), ('executive'), ('manager'), ('engineer'), ('ic'), ('investor');
alter table list_members ADD COLUMN member_role VARCHAR(128) REFERENCES member_roles (id);
CREATE TABLE member_groups (
  name VARCHAR(128) NOT NULL,
  list_member_id INT REFERENCES list_members (id) NOT NULL
);


alter table member_groups RENAME COLUMN id to name;
INSERT INTO member_roles (id)  values ('investor');
insert into member_groups VALUES ('test_delivery', 2),('test_delivery', 3),('test_delivery', 8),('test_bounce', 4),('test_bounce', 5),('test_bounce', 6),('test_bounce', 7);



select * from list_members l, member_groups g where l.id = g.list_member_id and g.name = 'test_delivery';


-- Import List Members
create temporary table csv_import (first_name text, last_name text, member_role text, company text, position text, email text);
\copy csv_import from '/Users/vlad/Documents/vlad-lkd.csv' WITH CSV HEADER DELIMITER AS ',';
insert into list_members (first_name, member_role, company, position, email) select first_name, member_role, company, position, email from csv_import;
drop table csv_import;


select t.subject, s.id as sent_email_id, s.created as email_sent, a.action, a.created as action_time from list_members l, email_templates t, email_actions a, sent_emails s where t.id = s.email_template_id and s.list_member_id = l.id and a.sent_email_id = s.id order by s.id desc limit 20;

select * from list_members where id = 2 and created < now();

select * from sent_emails where id = 13;


select * from course_cohorts;

select * from course_cohorts c where c.start_date <= now() and end_date > now();

select id, first_name, last_name, email from list_members l where email in ('akilburn924@gmail.com', 'gregsilin@gmail.com', 'myblake@gmail.com', 'julie.michelle.smith@gmail.com', 'michael.dore@gmail.com', 'dustin@dustinbuss.com', 'ferhat.hatay@gmail.com', 'alexcloudcto@gmail.com', 'kringotime@me.com', 'cshenoy@gmail.com', 'benvnguyen@gmail.com', 'john.celenza@gmail.com', 'brendan.hayes@gmail.com', 'shane.kelly@gmail.com', 'evan.hourigan@gmail.com', 'endre.soos@gmail.com', 'armen.abrahamian@gmail.com');

select *, l.*  from member_groups g, list_members l where g.list_member_id = l.id and g.name = 'inner-leadership-2018-nov';


select * from member_groups;

insert into member_groups values ('inner-leadership-2018-nov', 1);

-- 2018-11-01 12:18:42.670602

select * from course_cohorts;

insert into course_cohorts values ('inner-leadership-2018-nov', 'inner-leadership', to_date('2018-11-06', 'YYYY-MM-DD'), to_date('2019-02-15', 'YYYY-MM-DD'));

delete from member_groups g where list_member_id in (9129, 6671, 6911, 6995, 6996, 7021) and g.name = 'inner-leadership-2018-nov';

select s.list_member_id, l.first_name, l.email, t.subject, t.id, t.created from list_members l, sent_emails s, email_templates t where s.list_member_id in (9129, 6671, 6911, 6995, 6996, 7021) and t.id = s.email_template_id and l.id = s.list_member_id and t.id = 2871018535200300250 order by created desc;

select t.subject, count(*) from sent_emails s, email_templates t where t.id = s.email_template_id group by t.subject order by count(*) desc;
 

select * from email_templates where id = 2630011874810195904;

select * from tracking_hits where referrer_url is not null limit 10;


select l.* from list_member_locations l, member_groups g where g.name = 'inner-leadership-2018-nov' and g.list_member_id = l.id;

select l.* from list_member_locations l, member_groups g where g.name = 'test_course_1' and g.list_member_id = l.id;

select * from list_member_locations;

select * from member_groups where name = 'test_course_1';

select * from member_groups;

insert into list_member_locations values (2, 'a', 'a', 'a', 'a', '+01:00'), (3, '', '', '', '', '-08:00');

select * from list_member_locations where id < 10;

select * from sent_emails order by created desc;

update sent_emails set list_member_id = 5 where id >= 183;

delete from email_actions where sent_email_id > 183;

delete from tracked_urls where sent_email_id > 183;

select * from sent_emails where email_template_id = -436652262288456380;

select * from member_cookies;

update member_cookies set list_member_id = 5;