package sampleTranslationHtml


const TranslationForwarded = `
Return-Path: <vlad@giverts.com>
Received: from mail-lj1-f179.google.com (mail-lj1-f179.google.com [209.85.208.179])
 by inbound-smtp.us-west-2.amazonaws.com with SMTP id s3guf52n3mef53rccj4fs4stce1dmo0hf80p5b01
 for vlad@m.softsideoftech.com;
 Mon, 13 Aug 2018 19:19:21 +0000 (UTC)
X-SES-Spam-Verdict: PASS
X-SES-Virus-Verdict: PASS
Received-SPF: none (spfCheck: 209.85.208.179 is neither permitted nor denied by domain of giverts.com) client-ip=209.85.208.179; envelope-from=vlad@giverts.com; helo=mail-lj1-f179.google.com;
Authentication-Results: amazonses.com;
 spf=none (spfCheck: 209.85.208.179 is neither permitted nor denied by domain of giverts.com) client-ip=209.85.208.179; envelope-from=vlad@giverts.com; helo=mail-lj1-f179.google.com;
 dkim=pass header.i=@giverts-com.20150623.gappssmtp.com;
X-SES-RECEIPT: AEFBQUFBQUFBQUFIaW90bjlwSXNhVkp0QUUxZFdhZ3liRU1EN294aXYvQThRNVZ6RzdSVjJwY1M5bHhsWklFZy9HcytLQXMwbmdxKzZwWldWVnFEUnM4T0lSOFVwMEJHQWRVNkVnVDFOZUFCb0FiV3RReVY5bk92Mk9QZHM5ZGpDTDBMcVpZeWMyRVRSbkZUUTBXTXF6MllVSUMzZDhHVHA0VkVsb2lsWnh0b2NmMmNyU3Y4bjVHb0pGNlduQ2E5RWx2b0VjSE82dm4rd3FsVE1jcmJ6YUsyU3ByVTBXNmxPZzhyWEFPZmFYNmErRTNDa2tBN2pSY1ZlTWRDazYxb2tHVWNvMUhGTWJIditCS1Nkb09TQzBORk9IcmJ1NG5YME5ld0N5VlZSVitWak5rODFJdVUxc1E9PQ==
X-SES-DKIM-SIGNATURE: a=rsa-sha256; q=dns/txt; b=WrDWR6uk+oNbsMSieWOQmgljlMFmK4ox9FiAFmZC/lleq4BiyIkNOuB0MACT1MD3sZOGp308aEmdo+Ie0ect8HTy9v4AEdN9Drt7EGilVZZDSHZuyn/x1J5amYHO42CE8fsfiGnGiIPmmV59L6xCRQ9JheNR80XU7/VPUl0gdY4=; c=relaxed/simple; s=7v7vs6w47njt4pimodk5mmttbegzsi6n; d=amazonses.com; t=1534187961; v=1; bh=H3qvB3GFjIcur5IGnPobp+e052M+BmeuFqvPhNeLNME=; h=From:To:Cc:Bcc:Subject:Date:Message-ID:MIME-Version:Content-Type:X-SES-RECEIPT;
Received: by mail-lj1-f179.google.com with SMTP id l15-v6so13450317lji.6
        for <vlad@m.softsideoftech.com>; Mon, 13 Aug 2018 12:19:20 -0700 (PDT)
DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=giverts-com.20150623.gappssmtp.com; s=20150623;
        h=mime-version:from:date:message-id:subject:to;
        bh=3agQ7t0uOKVPzy665nUstVT3rLz6Pwq+84Kdh34MptY=;
        b=FuvNE0ZnSuISG/iSg8g9oTT0VCdRT7uY2mk7GjhFzjrS66Y07v+yNS0kOk8RGYxxnS
         M+jnxAqHll+DneZ4XpktqQ2RzeE9kHyxA5gZ7QwmVZTwaE6jBi6/vEmKQYyEh4GgaAJK
         /wGv9GAuo0ixX/oAzqvJsgoYAvKMJaAkCWfof53dY9CjsYdg1rQducDuFWjopxMhtbn0
         rrqCoAW+CxrwxeQ8tqNCgB+a4PUxjlMWx0+8ejUlg15ez9gjlsqveNnndbCmO6nLLppH
         MrtnJOIApkIL/YJC3CbgNkgVBh0rj59zWAJg+9e7S2yfIHPQN93uRmxHMD3U5PRy3ZqV
         kXFQ==
X-Google-DKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed;
        d=1e100.net; s=20161025;
        h=x-gm-message-state:mime-version:from:date:message-id:subject:to;
        bh=3agQ7t0uOKVPzy665nUstVT3rLz6Pwq+84Kdh34MptY=;
        b=l240/pzeWyAuhCGuml7UBhAawl45OntYtXiEl5FssKrPmqiXEvRic+kDXEwDdfMNYe
         zPUtoj3YUjVTRtDLuj3kKNtxKWptemcyWD4ZxhBBrIn3SHzF1PHf23dRN76KNVCqfNLB
         wWqprj7QJAtGTFegE7F4GJs+AZi6zV7R4MMc9AiBA96ch43SDq3XJYphpW07eVbOexI/
         nZgMKSFe5W2y6oKAULSqVO+xY2PvscMPdTRxeXc1dNOsjxl4wNdTOrSy5FrQjH5i1OQi
         adh7AkmXhdPkFQ54d69lC3PAOwkbRdSHQCj/FYX/qcRUjqMXuY+K54LL3DmNEfsBv7YG
         Ic3A==
X-Gm-Message-State: AOUpUlEzDxfYkLQjISACzhP9tR+5I3pElum1x53hnXUSl2hmGBk9yM/q
	p29zpAUP8Fpb7rMfZ+6+0jyk3AL2Zq83ZStj
X-Google-Smtp-Source: AA+uWPzLg2tR5qb1wSShaIM/eWFxtJmL3y+LqTU/B0aSYniVhZ3pYhLS18wnAmHF8Uyj0LUUkAQ3zA==
X-Received: by 2002:a2e:2114:: with SMTP id h20-v6mr13937943ljh.135.1534187958746;
        Mon, 13 Aug 2018 12:19:18 -0700 (PDT)
Return-Path: <vlad@giverts.com>
Received: from mail-lj1-f179.google.com (mail-lj1-f179.google.com. [209.85.208.179])
        by smtp.gmail.com with ESMTPSA id a14-v6sm3188669ljb.49.2018.08.13.12.19.17
        for <vlad@m.softsideoftech.com>
        (version=TLS1_2 cipher=ECDHE-RSA-AES128-GCM-SHA256 bits=128/128);
        Mon, 13 Aug 2018 12:19:18 -0700 (PDT)
Received: by mail-lj1-f179.google.com with SMTP id y17-v6so13465938ljy.8
        for <vlad@m.softsideoftech.com>; Mon, 13 Aug 2018 12:19:17 -0700 (PDT)
X-Received: by 2002:a2e:8514:: with SMTP id j20-v6mr12493278lji.10.1534187957431;
 Mon, 13 Aug 2018 12:19:17 -0700 (PDT)
MIME-Version: 1.0
From: Vlad Giverts <vlad@giverts.com>
Date: Mon, 13 Aug 2018 21:19:05 +0200
X-Gmail-Original-Message-ID: <CAP-N5vTzxKe91GeLh5JMbsv6Ebg1pG6mQ5qerOmT-asJ8YSQ6g@mail.gmail.com>
Message-ID: <CAP-N5vTzxKe91GeLh5JMbsv6Ebg1pG6mQ5qerOmT-asJ8YSQ6g@mail.gmail.com>
Subject: test 4
To: "vlad@m.softsideoftech.com" <vlad@m.softsideoftech.com>
Content-Type: multipart/alternative; boundary="00000000000038b5f0057355f81e"

--00000000000038b5f0057355f81e
Content-Type: text/plain; charset="UTF-8"

my mail body!
-- 
Purposeful Leadership Coaching
softsideoftech.com

--00000000000038b5f0057355f81e
Content-Type: text/html; charset="UTF-8"

<div dir="ltr">my mail body!</div>-- <br><div dir="ltr" class="gmail_signature" data-smartmail="gmail_signature"><div dir="ltr">Purposeful Leadership Coaching<div><a href="http://softsideoftech.com">softsideoftech.com</a></div></div></div>

--00000000000038b5f0057355f81e--
`