package sampleEmails
// --([0-9a-f]+)\nContent-Type: text/plain[\S\s]*?\n\n([\S\s]*?)\n\n--\1\nContent-Type: text/html[\S\s]*?\n\n([\S\s]*?)\n\n--\1

const EmailSampleWithInlineAttachment = `
Received: by mail-wr0-f169.google.com with SMTP id a15-v6so4132271wrm.0
        for <vgiverts@gmail.com>; Sat, 19 May 2018 02:22:13 -0700 (PDT)
X-Gm-Message-State: ALKqPwes2YQXbEqUbhGEQhd5Y2o61xwWsGn7/5mhyL7PBS9EFm+xudkA
	G/NRnhJikLuEFvyKHKxx5lllU7c2bmJbK3a9pCo=
X-Google-Smtp-Source: AB8JxZrdl2QY5fUFuLAsSL/nfQCWqjGUs0W60WBmxhFs7Vf2aJjrPtpHZxYr1xVKt+4Fx/FNTaKLAy13bKXkFdMMnHI=
X-Received: by 2002:adf:94a5:: with SMTP id 34-v6mr8177087wrr.43.1526721732613;
 Sat, 19 May 2018 02:22:12 -0700 (PDT)
MIME-Version: 1.0
From: Send Test <send-test@softsideoftech.com>
Date: Sat, 19 May 2018 11:22:01 +0200
X-Gmail-Original-Message-ID: <CAP-N5vQufAi=Tsn9M4zNaUirrxXb1Cefjv4360nNNov-pq+HGQ@mail.gmail.com>
Message-ID: <CAP-N5vQufAi=Tsn9M4zNaUirrxXb1Cefjv4360nNNov-pq+HGQ@mail.gmail.com>
Subject: inline attachment
To: Vlad Giverts <vgiverts@gmail.com>
Content-Type: multipart/related; boundary="0000000000008b36f9056c8b9ae2"

--0000000000008b36f9056c8b9ae2
Content-Type: multipart/alternative; boundary="0000000000008b36f6056c8b9ae1"

--0000000000008b36f6056c8b9ae1
Content-Type: text/plain; charset="UTF-8"

asdf[image: favicon.bmp]
-- 
Purposeful Leadership Coaching
softsideoftech.com

--0000000000008b36f6056c8b9ae1
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div dir=3D"ltr">asdf<img src=3D"cid:16377b4a7fde70feab3" alt=3D"favicon.bm=
p" class=3D"" style=3D"max-width: 100%; opacity: 1;"></div>-- <br><div dir=
=3D"ltr" class=3D"gmail_signature" data-smartmail=3D"gmail_signature"><div =
dir=3D"ltr">Purposeful Leadership Coaching<div><a href=3D"http://softsideof=
tech.com">softsideoftech.com</a></div></div></div>

--0000000000008b36f6056c8b9ae1--
--0000000000008b36f9056c8b9ae2
Content-Type: image/bmp; name="favicon.bmp"
Content-Disposition: inline; filename="favicon.bmp"
Content-Transfer-Encoding: base64
Content-ID: <16377b4a7fde70feab3>
X-Attachment-Id: 16377b4a7fde70feab3

AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAABILAAASCwAAAAAA
AAAAAAAAAAAAs3BxALNwcRSzcXKrs3Fy6bNxcX2zcHAKuGpkALhtbgGzb3AGs3BxbbNxcuWzcXK6
s3BxHbNwcQAAAAAAAAAAALNxcgCzcHFKs3Fy+LNxcv+zcXL1s3FypLNxcpWzcXKZs3FynrNxcvCz
cXL/s3Fy/bNxcmCzcXIAAAAAAAAAAACzcXIAs3ByMLNxct+zcXL/s3Fy/7Nxcv+zcXL/s3Fy/7Nx
cv+zcXL/s3Fy/7NxcuqzcHFAs3FxAAAAAAAAAAAAtG9xALVnbQCzcHFLs3Fy57Nxcv+zcXL/s3Fy
/7Nxcv+zcXL/s3Fy/7NxcvGzcHFds25tArNvcAAAAAAAsG1xALBucQGwb3AHs3BxKLNxcuGzcXL/
s3Fy/7Nxcv+zcXL/s3Fy/7Nxcv+zcXLws3BxO7NxcgAAAAAAAAAAALp7fACzcXJYs3FyurNxcraz
cXL5s3Fy/7Nxcv+zcXL/s3Fy/7Nxcv+zcXL/s3Fy/bNxcl60cXIAAAAAAAAAAACxbW8Hs3FytLNx
cv+zcXL/s3Fy/7Nxcv+zcXL/s3Fy/7Nxcv+zcXL/s3Fy/7NxcvqzcHJQs3FyAAAAAAAAAAAApFFj
ALNxcmizcXLis3Fy/7Nxcv+zcXL/s3Fy/7Nxcv+zcXL/s3Fy/7Nxcv+zcXL8s3Bxk7JwcRiycXIA
s21uALJvcACyb28Cs3BxK7NxcomzcXLcs3Fy+7Nxcv+zcXL/s3Fy/7Nxcv+zcXL/s3Fy/7Nxcv2z
cXLAs3BxNbRycgAAAAAAAAAAALJvcACxbG4Bs3BxLrNxcsmzcXL/s3Fy/7Nxcv+zcXL/s3Fy/rNx
cvqzcXL/s3Fy/7Nxcrqzb3EOAAAAAAAAAACybXAAs3JzALNwcmKzcXL2s3Fy/7Nxcv+zcXL/s3Fy
/7Nxcv+zcHG+s3FyxbNxcvqzcXKdtG9wCAAAAAAAAAAAsW1vALFsbQOzcXKps3Fy/7Nxcv+zcXL/
s3Fy/7Nxcv+zcXL/s3Fyt7NwcSOzcHFCsnBxFbNycgAAAAAAAAAAALNucACzaGsBs3FymLNxcv+z
cXL/s3Fy/7Nxcv+zcXL/s3Fy/7NxcqqybW8Fsm5wAAAAAAAAAAAAAAAAAAAAAACzcXIAs3ByNLNx
ctWzcXL/s3Fy/7Nxcv+zcXL/s3Fy/7Nxcv+zcXLfs3BxQbNxcgCvY2MAAAAAAAAAAAAAAAAAs3Fy
ALNxcmazcXL+s3Fy+7Nxcq6zcXKos3Fyq7NxcqmzcXL4s3Fy/7NwcXezcnMAsGRkAAAAAAAAAAAA
AAAAALNwcQCycHEhs3FyrrNxcqqzcHEcs25uArBubgOzb3EUs3BynrNxcraycHErs3FxAK9gYAAA
AAAAwAMAAMADAADAAwAAwAMAAIAHAACABwAAAAcAAAADAACAAQAA4AAAAPAAAADgAQAA4AcAAOAH
AADgBwAA4AcAAA==
--0000000000008b36f9056c8b9ae2--

`
