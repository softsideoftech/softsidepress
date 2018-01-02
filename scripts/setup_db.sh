#!/usr/bin/env bash
# assumes pg_hba.conf is set to 'trust' for local connections

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

wget http://www.ip2location.com/download/?token=$2&file=DB11LITEBIN;

sudo -u postgres