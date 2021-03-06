CREATE TABLE email_templates (
  id      BIGINT PRIMARY KEY,
  subject VARCHAR(256)                        NOT NULL,
  body    TEXT                                NOT NULL,
  created TIMESTAMP DEFAULT current_timestamp NOT NULL
);

CREATE TABLE member_roles (
  id VARCHAR(128) PRIMARY KEY
);
INSERT INTO member_roles (id)
VALUES ('founder'),
       ('executive'),
       ('manager'),
       ('engineer'),
       ('ic');

CREATE TABLE email_actions_enum (
  action CHAR(16) PRIMARY KEY
);

INSERT INTO email_actions_enum (action)
VALUES ('sent'),
       ('delivered'),
       ('opened'),
       ('clicked'),
       ('hard_bounce'),
       ('soft_bounce'),
       ('complaint');

CREATE TABLE tracked_urls (
  id            BIGINT PRIMARY KEY,
  url           VARCHAR(1024)                       NOT NULL UNIQUE,
  target_url    VARCHAR(1024),
  sent_email_id INT REFERENCES sent_emails (id),
  created       TIMESTAMP DEFAULT current_timestamp NOT NULL,
  login_id      INT REFERENCES list_members (id)
);

alter table tracked_urls drop column login;
alter table tracked_urls add column login_id      INT REFERENCES list_members (id);

CREATE TABLE IF NOT EXISTS list_members (
  id           INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  first_name   VARCHAR(128)                        NOT NULL,
  last_name    VARCHAR(128),
  company      VARCHAR(128),
  position     VARCHAR(128),
  created      TIMESTAMP DEFAULT current_timestamp NOT NULL,
  subscribed   TIMESTAMP,
  unsubscribed TIMESTAMP,
  updated      TIMESTAMP DEFAULT current_timestamp NOT NULL,
  email        VARCHAR(128) UNIQUE                 NOT NULL,
  member_role  VARCHAR(128) REFERENCES member_roles (id),
  timezone
);

CREATE TABLE IF NOT EXISTS list_member_locations (
  id INT PRIMARY KEY references list_members(id),
  country_code CHARACTER(2)           NOT NULL,
  country_name CHARACTER VARYING(64)  NOT NULL,
  region_name  CHARACTER VARYING(128) NOT NULL,
  city_name    CHARACTER VARYING(128) NOT NULL,
  time_zone    CHARACTER VARYING(8)   NOT NULL
);

CREATE TABLE sent_emails (
  id                INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  email_template_id BIGINT REFERENCES email_templates (id),
  list_member_id    INT REFERENCES list_members (id),
  third_party_id    CHAR(64),
  created           TIMESTAMP DEFAULT current_timestamp NOT NULL,
);
CREATE INDEX sent_email__third_party_id
  ON sent_emails (third_party_id);
CREATE INDEX sent_email__list_member_id__email_template_id
  ON sent_emails (list_member_id, email_template_id);

CREATE TABLE email_actions (
  id            INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  sent_email_id INT REFERENCES sent_emails (id),
  action        CHAR(16) REFERENCES email_actions_enum (action) NOT NULL,
  metadata      VARCHAR(1024),
  created       TIMESTAMP DEFAULT current_timestamp             NOT NULL
);

CREATE TABLE member_cookies (
  id             BIGINT PRIMARY KEY,
  list_member_id INT REFERENCES list_members (id)    NOT NULL,
  created        TIMESTAMP DEFAULT current_timestamp NOT NULL,
  updated        TIMESTAMP DEFAULT current_timestamp NOT NULL,
  logged_in      TIMESTAMP
);

CREATE TABLE tracking_hits (
  id                INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  tracked_url_id    BIGINT REFERENCES tracked_urls (id)             NOT NULL,
  list_member_id    INT REFERENCES list_members (id),
  member_cookie_id  BIGINT, -- No foreign key constraint because we only create member cookies when we have a listMemberId to associate with it
  ip_address        BIGINT                                          NOT NULL,
  ip_address_string VARCHAR(128)                                    NOT NULL,
  referrer_url      VARCHAR(1024),
  created           TIMESTAMP DEFAULT current_timestamp             NOT NULL,
  time_on_page      INT DEFAULT 0
);

CREATE TABLE ip2location (
  ip_from      BIGINT                 NOT NULL,
  ip_to        BIGINT                 NOT NULL,
  country_code CHARACTER(2)           NOT NULL,
  country_name CHARACTER VARYING(64)  NOT NULL,
  region_name  CHARACTER VARYING(128) NOT NULL,
  city_name    CHARACTER VARYING(128) NOT NULL,
  latitude     REAL                   NOT NULL,
  longitude    REAL                   NOT NULL,
  zip_code     CHARACTER VARYING(30)  NOT NULL,
  time_zone    CHARACTER VARYING(8)   NOT NULL,
  CONSTRAINT ip2location_pkey PRIMARY KEY (ip_from, ip_to)
);

CREATE TABLE member_groups (
  name           VARCHAR(128)                        NOT NULL,
  list_member_id INT REFERENCES list_members (id)    NOT NULL,
  created        TIMESTAMP DEFAULT current_timestamp NOT NULL,
  PRIMARY KEY (list_member_id, name)
);
CREATE INDEX member_groups__list_member_id
  ON member_groups (list_member_id);

drop table course_cohorts;
CREATE TABLE course_cohorts (
  name        VARCHAR(128) PRIMARY KEY,
  course_name VARCHAR(128) NOT NULL,
  start_date  TIMESTAMP    NOT NULL,
  end_date    TIMESTAMP    NOT NULL
)
