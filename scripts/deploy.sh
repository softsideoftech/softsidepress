#!/usr/bin/env bash
cd ~/go/src/softside/
gzip softside
scp softside.gz root@softsideoftech.com:
ssh root@softsideoftech.com 'gunzip -f softside.gz'
ssh root@softsideoftech.com 'service softside restart'