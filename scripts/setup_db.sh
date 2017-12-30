#!/usr/bin/env bash

sudo -u postgres psql -c "create user softside with encrypted password 'asdf'";
sudo -u postgres psql -c "create database softside";
sudo -u postgres psql -c "grant all privileges on database softside to softside";

sudo -u postgres psql -u softside -d softside - AND NOW WHAT?

wget http://www.ip2location.com/download/?token=$2&file=DB11LITEBIN;

sudo -u postgres copy