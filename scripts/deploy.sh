#!/usr/bin/env bash
gzip /tmp/softside
scp /tmp/softside.gz root@softsideoftech.com:
ssh root@softsideoftech.com 'gunzip -f softside.gz'
ssh root@softsideoftech.com 'service softside restart'