package sampleEmails

const EmailWithMultiCc = `
Received: by mail-wr0-f173.google.com with SMTP id p18-v6so10368966wrm.1
        for <stacyap@gmail.com>; Fri, 18 May 2018 13:16:39 -0700 (PDT)
X-Gm-Message-State: ALKqPweIPxWJ1vCKcbs5C9N+1cxJRt71/UfWmkEHKXzA146ndKWlhe5+
	dFqMdC0z+cW9AnYChnSPUhuPa0JO/5K1V+izDrM=
X-Google-Smtp-Source: AB8JxZqYEitgpUhWrH8RjewNTZxGY//+liL5c7BX0wJI22bdRiOHaMX1iVnKtRZsMaC/xYsvOKX0Uulw6Nx2h5pLj6o=
X-Received: by 2002:adf:a789:: with SMTP id j9-v6mr8341811wrc.95.1526674598724;
 Fri, 18 May 2018 13:16:38 -0700 (PDT)
MIME-Version: 1.0
Received: by 10.223.149.129 with HTTP; Fri, 18 May 2018 13:16:38 -0700 (PDT)
From: Send Test <send-test@softsideoftech.com>
Date: Fri, 18 May 2018 22:16:38 +0200
X-Gmail-Original-Message-ID: <CAP-N5vRGcrAEQBJ8sz8NMg6SHwHCLRPF0v3arF=2hA9Evs=vuA@mail.gmail.com>
Message-ID: <CAP-N5vRGcrAEQBJ8sz8NMg6SHwHCLRPF0v3arF=2hA9Evs=vuA@mail.gmail.com>
Subject: good morning
To: stacya <stacyap@gmail.com>
Cc: Vlad2 <vlad@cloudmars.com>, Vlad3 <vlad@giverts.com>
Content-Type: multipart/alternative; boundary="000000000000250ab4056c80a132"

--000000000000250ab4056c80a132
Content-Type: text/plain; charset="UTF-8"

love you baby...

--000000000000250ab4056c80a132
Content-Type: text/html; charset="UTF-8"

<div dir="ltr">love you baby...</div>

--000000000000250ab4056c80a132--

`