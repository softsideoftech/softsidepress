package sampleEmails

const EmailWithMultiTo = `
Received: by mail-wr0-f171.google.com with SMTP id l41-v6so14260872wre.7
        for <vgiverts@gmail.com>; Sat, 26 May 2018 12:05:29 -0700 (PDT)
X-Gm-Message-State: ALKqPwfZNys5qZPawllRqz8H22Ee/PRu0H1qd8s8L8d7piW8tAU6EAe/
	a+kjTzMeONSHneUYr/D1rQdCDe/SVdniXfU+glg=
X-Google-Smtp-Source: AB8JxZqsxMDbbUxU2L7jQ1khMjIgLu6hEYSo14N+CKTQ9u6Ktuf3aGLu5nFVrEp5vepCRKSeH4xmNKGBTbw54aM9KIg=
X-Received: by 2002:adf:a789:: with SMTP id j9-v6mr5480275wrc.95.1527361529253;
 Sat, 26 May 2018 12:05:29 -0700 (PDT)
MIME-Version: 1.0
From: Send Test <send-test@softsideoftech.com>
Date: Sat, 26 May 2018 21:05:17 +0200
X-Gmail-Original-Message-ID: <CAP-N5vQ3-dCqLos03O+GDwtiiqMUNTeGyPeYyk_OnJwaYEBzcA@mail.gmail.com>
Message-ID: <CAP-N5vQ3-dCqLos03O+GDwtiiqMUNTeGyPeYyk_OnJwaYEBzcA@mail.gmail.com>
Subject: test multi-to
To: Vlad Giverts <vgiverts@gmail.com>, Vlad Giverts <vlad@giverts.com>
Cc: "vlad@cloudmars.com" <vlad@cloudmars.com>
Content-Type: multipart/alternative; boundary="000000000000651b64056d2091af"

--000000000000651b64056d2091af
Content-Type: text/plain; charset="UTF-8"

body of test multi-to
--
Purposeful Leadership Coaching
softsideoftech.com

--000000000000651b64056d2091af
Content-Type: text/html; charset="UTF-8"

<div dir="ltr">body of test multi-to</div>-- <br><div dir="ltr" class="gmail_signature" data-smartmail="gmail_signature"><div dir="ltr">Purposeful Leadership Coaching<div><a href="http://softsideoftech.com">softsideoftech.com</a></div></div></div>

--000000000000651b64056d2091af--

`