
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

-- additions for video course
alter table member_groups ADD COLUMN created TIMESTAMP DEFAULT current_timestamp NOT NULL;
alter table member_groups ADD PRIMARY KEY (list_member_id, name);
CREATE INDEX member_groups__list_member_id ON member_groups (list_member_id);
CREATE TABLE course_cohorts (
  course_name VARCHAR(128) NOT NULL,
  start_date  TIMESTAMP    NOT NULL,
  end_date    TIMESTAMP    NOT NULL,
  PRIMARY KEY (course_name, start_date)
);


select * from course_cohorts;

select id, first_name, last_name, email from list_members l where email in ('akilburn924@gmail.com', 'gregsilin@gmail.com', 'myblake@gmail.com', 'julie.michelle.smith@gmail.com', 'michael.dore@gmail.com', 'dustin@dustinbuss.com', 'ferhat.hatay@gmail.com', 'alexcloudcto@gmail.com', 'kringotime@me.com', 'cshenoy@gmail.com', 'benvnguyen@gmail.com', 'john.celenza@gmail.com', 'brendan.hayes@gmail.com', 'shane.kelly@gmail.com', 'evan.hourigan@gmail.com', 'endre.soos@gmail.com', 'armen.abrahamian@gmail.com');

update list_members set first_name = 'Alex' where email = 'alexcloudcto@gmail.com';

select * from member_groups;



select * from tracking_hits where referrer_url is not null limit 10;

country_code CHARACTER(2)           NOT NULL,
country_name CHARACTER VARYING(64)  NOT NULL,
region_name  CHARACTER VARYING(128) NOT NULL,
city_name    CHARACTER VARYING(128) NOT NULL,
time_zone

explain verbose select ip.country_code, ip.country_name, ip.region_name, ip.city_name, ip.time_zone, ip_address, list_member_id from tracking_hits h, ip2location ip where ip_address >= ip.ip_from and ip_address <= ip_to and list_member_id is not null limit 20;

with hits as (select list_member_id, ip_address, count(*) from tracking_hits where list_member_id is not null group by list_member_id, ip_address),
user_locations as (select ip.country_code, ip.country_name, ip.region_name, ip.city_name, ip.time_zone, ip_address, list_member_id from hits h, ip2location ip where ip_address >= ip.ip_from and ip_address <= ip_to)
select city_name, list_member_id, count(*) from user_locations group by city_name, list_member_id limit 20;


with user_ips as (select list_member_id, ip_address from tracking_hits where list_member_id is not null),
    user_ips2 as (select list_member_id, mode() within group (order by ip_address)
    from user_ips
    group by list_member_id)
select list_member_id, ip.country_code, ip.country_name, ip.region_name, ip.city_name, ip.time_zone from user_ips2 u, ip2location ip where u.mode >= ip.ip_from and u.mode <= ip_to limit 10;

select * from list_member_locations;

truncate table list_member_locations;

insert into list_member_locations (with user_ips as (select list_member_id, ip_address from tracking_hits where list_member_id is not null),
    user_ips2 as (select list_member_id, mode() within group (order by ip_address)
                  from user_ips
                  group by list_member_id)
select list_member_id, ip.country_code, ip.country_name, ip.region_name, ip.city_name, ip.time_zone from user_ips2 u, ip2location ip where u.mode >= ip.ip_from and u.mode <= ip_to);

insert into list_member_locations (select 1,  ip.country_code, ip.country_name, ip.region_name, ip.city_name, ip.time_zone from ip2location ip where 1386970622 >= ip_from and 1386970622 <= ip_to);

delete from list_member_locations where id = 1;

select region_name, count(*) from list_member_locations group by region_name order by count(*) desc;

select count(*) from list_member_locations;