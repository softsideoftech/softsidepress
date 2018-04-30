#!/usr/bin/env bash
## Assumes that:
## - pg_hba.conf is set to 'trust' for local connections
## IP db is downloaded and unzipped to /root
### https://lite.ip2location.com/database/ip-country-region-city-latitude-longitude-zipcode-timezone

# Dowload and install the latest Postgres
echo 'deb http://apt.postgresql.org/pub/repos/apt/ xenial-pgdg main' >> /etc/apt/sources.list.d/pgdg.list
sudo apt-get update
sudo apt-get install postgresql-10 postgresql-10-contrib

sudo -u postgres psql -c "create user softside with encrypted password 'asdf'";
sudo -u postgres psql -c "create database softside";
sudo -u postgres psql -c "grant all privileges on database softside to softside";

cp /root/go/src/softside/scripts/dbSchema.sql /var/lib/postgresql/
chmod a+r /var/lib/postgresql/dbSchema.sql

sudo -u postgres psql -d softside -U softside -f /var/lib/postgresql/dbSchema.sql

chmod a+r /root/IP2LOCATION-LITE-DB11.CSV
mv /root/IP2LOCATION-LITE-DB11.CSV /var/lib/postgresql/
sudo -u postgres psql -d softside -U softside -c "\copy ip2location FROM '/var/lib/postgresql/IP2LOCATION-LITE-DB11.CSV' WITH CSV QUOTE AS '\"'"


# Import initial list member list. Assumes csv list is downloaded to /root
cp /root/member-list.csv /var/lib/postgresql/
chmod a+r /var/lib/postgresql/member-list.csv
create temporary table csv_import (first_name text, last_name text, member_role text, company text, position text, email text);
\copy csv_import from '/var/lib/postgresql/member-list.csv' WITH CSV HEADER DELIMITER AS ',';
insert into list_members (first_name, member_role, company, position, email) select first_name, member_role, company, position, email from csv_import;
drop table csv_import;