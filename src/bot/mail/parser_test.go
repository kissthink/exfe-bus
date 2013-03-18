package mail

import (
	"bytes"
	"model"
	"net/mail"
	"testing"
)

var config model.Config

func init() {
	config.Email.Prefix = "x"
	config.Email.Domain = "0d0f.com"
}

func TestInit(t *testing.T) {
	var str = `Delivered-To: panda@0d0f.com
Received: by 10.70.0.234 with SMTP id 10csp193484pdh;
        Sun, 17 Mar 2013 23:15:37 -0700 (PDT)
X-Received: by 10.152.113.34 with SMTP id iv2mr13078933lab.20.1363587335984;
        Sun, 17 Mar 2013 23:15:35 -0700 (PDT)
Return-Path: <googollee@hotmail.com>
Received: from bay0-omc2-s22.bay0.hotmail.com (bay0-omc2-s22.bay0.hotmail.com. [65.54.190.97])
        by mx.google.com with ESMTP id ia4si6157124lab.180.2013.03.17.23.15.34;
        Sun, 17 Mar 2013 23:15:35 -0700 (PDT)
Received-SPF: pass (google.com: domain of googollee@hotmail.com designates 65.54.190.97 as permitted sender) client-ip=65.54.190.97;
Authentication-Results: mx.google.com;
       spf=pass (google.com: domain of googollee@hotmail.com designates 65.54.190.97 as permitted sender) smtp.mail=googollee@hotmail.com
Received: from BAY152-DS11 ([65.54.190.124]) by bay0-omc2-s22.bay0.hotmail.com with Microsoft SMTPSVC(6.0.3790.4675);
	 Sun, 17 Mar 2013 23:15:34 -0700
X-EIP: [cqX6eeEhbwCCOhFZWRs2ZHcUKc1BKGRH]
X-Originating-Email: [googollee@hotmail.com]
Message-ID: <BAY152-ds11D2686FF50F86678FFD1AA0E80@phx.gbl>
Return-Path: googollee@hotmail.com
From: "Lee Googol Lee" <googollee@hotmail.com>
To: googollee@gmail.com;panda@0d0f.com
Date: Sun, 17 Mar 2013 23:15:33 -0700
Subject: =?utf-8?B?5rS75Yqo?=
Content-Type: multipart/alternative;
	boundary="_32936406-5bb3-4fc0-b17a-11cc15875002_"
MIME-Version: 1.0
X-OriginalArrivalTime: 18 Mar 2013 06:15:34.0264 (UTC) FILETIME=[0008CF80:01CE23A0]

--_32936406-5bb3-4fc0-b17a-11cc15875002_
Content-Type: text/plain; charset="utf-8"
Content-Transfer-Encoding: base64


--_32936406-5bb3-4fc0-b17a-11cc15875002_
Content-Type: text/html; charset="utf-8"
Content-Transfer-Encoding: base64

DQoNCjxzdHlsZSB0eXBlPSJ0ZXh0L2NzcyI+DQogICAgDQogICAgDQogICAgLk1haW5Db250YWlu
ZXINCiAgICB7DQogICAgICAgIGZvbnQtZmFtaWx5OiA7IC8qIFRoZSB0b3RhbCB3aWR0aCBpcyA3
MDBweCwgd2hpY2ggd2UgZ2V0IGJ5IGhhdmluZyBwYWRkaW5nIG9mIDUwcHggb24gZWFjaCBsZWZ0
IGFuZCByaWdodCBzaWRlIGFuZCBhIDYwMCB3aWR0aCBzZXR0aW5nICovDQogICAgICAgIGJhY2tn
cm91bmQtY29sb3I6ICNFNkU2RTY7DQogICAgfQ0KDQogICAgLk1haW5WZXJ0aWNhbEJvcmRlcg0K
ICAgIHsNCiAgICAgICAgd2lkdGg6IDUwcHg7DQogICAgfQ0KICAgIC5NYWluSG9yaXpvbnRhbEJv
cmRlcg0KICAgIHsNCiAgICAgICAgaGVpZ2h0OiAyMHB4Ow0KICAgIH0NCiAgICANCiAgICAuSW5u
ZXJIb3Jpem9udGFsQm9yZGVyDQogICAgew0KICAgICAgICBoZWlnaHQ6IDIwcHg7DQogICAgfQ0K
ICAgIC5Jbm5lclZlcnRpY2FsQm9yZGVyDQogICAgew0KICAgICAgICB3aWR0aDogNDBweDsNCiAg
ICAgICAgYmFja2dyb3VuZC1jb2xvcjogI0ZGRkZGRjsgICAgICAgIA0KICAgIH0NCiAgICAuQ29u
dGVudENvbnRhaW5lcg0KICAgIHsNCiAgICAgICAgd2lkdGg6IDUyMHB4Ow0KICAgICAgICBiYWNr
Z3JvdW5kLWNvbG9yOiAjRkZGRkZGOyAgICAgICAgDQogICAgfQ0KICAgIA0KICAgIC5Jbm5lclZl
cnRpY2FsQm9yZGVyV2lkdGgNCiAgICB7DQogICAgICAgIHdpZHRoOiA0MHB4Ow0KICAgIH0NCiAg
ICAuQ29udGVudFdpZHRoDQogICAgew0KICAgICAgICB3aWR0aDogNTIwcHg7DQogICAgfQ0KICAg
IA0KICAgIC5Gb290ZXJDb250YWluZXINCiAgICB7DQogICAgICAgIHRleHQtYWxpZ246IHJpZ2h0
Ow0KICAgIH0NCjwvc3R5bGU+DQoNCg0KPHRhYmxlIGNlbGxwYWRkaW5nPSIwIiBjZWxsc3BhY2lu
Zz0iMCIgY2xhc3M9Ik1haW5Db250YWluZXIiPg0KICAgIDx0ciBjbGFzcz0iTWFpbkhvcml6b250
YWxCb3JkZXIiPg0KICAgICAgICA8dGQgY2xhc3M9Ik1haW5WZXJ0aWNhbEJvcmRlciI+PC90ZD4g
ICAgDQogICAgICAgIDx0ZCBjbGFzcz0iSW5uZXJWZXJ0aWNhbEJvcmRlcldpZHRoIj48L3RkPg0K
ICAgICAgICA8dGQgY2xhc3M9IkNvbnRlbnRXaWR0aCI+PC90ZD4NCiAgICAgICAgPHRkIGNsYXNz
PSJJbm5lclZlcnRpY2FsQm9yZGVyV2lkdGgiPjwvdGQ+DQogICAgICAgIDx0ZCBjbGFzcz0iTWFp
blZlcnRpY2FsQm9yZGVyIj48L3RkPiAgICAgICAgDQogICAgPC90cj4NCiAgICA8dHIgY2xhc3M9
IklubmVySG9yaXpvbnRhbEJvcmRlciI+DQogICAgICAgIDx0ZCBjbGFzcz0iTWFpblZlcnRpY2Fs
Qm9yZGVyIj48L3RkPg0KICAgICAgICA8dGQgY2xhc3M9IklubmVyVmVydGljYWxCb3JkZXIiPjwv
dGQ+DQogICAgICAgIDx0ZCBjbGFzcz0iQ29udGVudENvbnRhaW5lciI+PC90ZD4NCiAgICAgICAg
PHRkIGNsYXNzPSJJbm5lclZlcnRpY2FsQm9yZGVyIj48L3RkPg0KICAgICAgICA8dGQgY2xhc3M9
Ik1haW5WZXJ0aWNhbEJvcmRlciI+PC90ZD4NCiAgICA8L3RyPg0KICAgIDx0cj4NCiAgICAgICAg
PHRkIGNsYXNzPSJNYWluVmVydGljYWxCb3JkZXIiPjwvdGQ+DQogICAgICAgIDx0ZCBjbGFzcz0i
SW5uZXJWZXJ0aWNhbEJvcmRlciI+PC90ZD4NCiAgICAgICAgPHRkIGNsYXNzPSJDb250ZW50Q29u
dGFpbmVyIj4NCiAgICAgICAgICAgIDxkaXYgY2xhc3M9IkNvbnRlbnRDb250YWluZXIiPg0KICAg
ICAgICAgICAgICAgIA0KDQo8c3R5bGUgdHlwZT0idGV4dC9jc3MiPg0KICAgIA0KICAgIC5NZWV0
aW5nUmVxdWVzdEhlYWRlcg0KICAgIHsNCiAgICAgICAgZm9udC1zaXplOiAyMnB4Ow0KICAgICAg
ICBmb250LXdlaWdodDpib2xkOw0KICAgICAgICBjb2xvcjojNDQ0NDQ0Ow0KICAgIH0NCiAgICAu
TWVldGluZ1JlcXVlc3RNZXNzYWdlQ29udGFpbmVyDQogICAgew0KICAgICAgICBwYWRkaW5nLXRv
cDoxNnB4Ow0KICAgIH0NCiAgICAuTWVldGluZ1JlcXVlc3RNZXNzYWdlDQogICAgew0KICAgICAg
ICBmb250LXNpemU6IDIycHg7DQogICAgICAgIGNvbG9yOiNGNDc5M0E7DQogICAgfQ0KICAgIC5N
ZWV0aW5nUmVxdWVzdFF1b3RlDQogICAgew0KICAgICAgICBmb250LWZhbWlseTogOw0KICAgICAg
ICBmb250LXdlaWdodDpib2xkOw0KICAgICAgICBmb250LXNpemU6MjRwdDsNCiAgICAgICAgY29s
b3I6Izg4ODg4ODsNCiAgICB9DQogICAgLk1lZXRpbmdSZXF1ZXN0RGVzY3JpcHRpb24NCiAgICB7
DQogICAgICAgIGZvbnQtZmFtaWx5OiA7DQogICAgICAgIGNvbG9yOiM0NDQ0NDQ7DQogICAgICAg
IGZvbnQtc2l6ZToxM3B4Ow0KICAgIH0NCiAgICAuTWVldGluZ1JlcXVlc3RIUnVsZQ0KICAgIHsN
CiAgICAgICAgYmFja2dyb3VuZC1jb2xvcjogI0VCRUJFQjsNCiAgICAgICAgZm9udC1zaXplOiAx
cHg7DQogICAgICAgIGhlaWdodDoxcHg7DQogICAgICAgIHdpZHRoOjEwMCU7DQogICAgICAgIG1h
cmdpbjoxNnB4IDBweDsNCiAgICB9DQogICAgLk1lZXRpbmdSZXF1ZXN0VGFibGUNCiAgICB7DQog
ICAgICAgIHdpZHRoOjEwMCU7DQogICAgICAgIGJvcmRlci1jb2xsYXBzZTpjb2xsYXBzZTsNCiAg
ICB9DQogICAgLk1lZXRpbmdSZXF1ZXN0VGFibGUgVEQNCiAgICB7DQogICAgICAgIHBhZGRpbmct
dG9wOjE2cHg7DQogICAgfQ0KICAgIC5NZWV0aW5nUmVxdWVzdFRpbWVMb2NhdGlvbkNvbnRhaW5l
cg0KICAgIHsNCiAgICAgICAgZm9udC1zaXplOjE2cHg7DQogICAgICAgIGNvbG9yOiM4ODg4ODg7
DQogICAgICAgIHdpZHRoOjEwMCU7DQogICAgfQ0KICAgIC5NZWV0aW5nUmVxdWVzdENhbmNlbA0K
ICAgIHsNCiAgICAgICAgaGVpZ2h0OjQ4cHg7DQogICAgICAgIHdpZHRoOjQ4cHg7DQogICAgICAg
IG1hcmdpbi1yaWdodDoxMnB4Ow0KICAgIH0NCiAgICANCjwvc3R5bGU+DQoNCjxkaXYgY2xhc3M9
Ik1lZXRpbmdSZXF1ZXN0SGVhZGVyIj5MZWUgR29vZ29sIExlZSDlkJHkvaDlj5HpgIHkuobigJwm
IzI3OTYzOyYjMjExNjA74oCd55qE6YKA6K+3PC9kaXY+DQo8dGFibGUgY2xhc3M9Ik1lZXRpbmdS
ZXF1ZXN0VGFibGUiPg0KICAgIDx0cj4NCiAgICA8dGQgY2xhc3M9Ik1lZXRpbmdSZXF1ZXN0VGlt
ZUxvY2F0aW9uQ29udGFpbmVyIj4NCiAgICAgICAgPGRpdj4yMDEz5bm0M+aciDE55pelPC9kaXY+
PGRpdj7kuIrljYg5OjAwIC0g5LiK5Y2IMTA6MDA8L2Rpdj4NCiAgICAgICAgPGRpdj4mIzIyMzIw
OyYjMjg4NTc7PC9kaXY+DQogICAgPC90ZD4NCiAgICA8L3RyPg0KPC90YWJsZT4NCg0KDQo8dGFi
bGUgY2xhc3M9Ik1lZXRpbmdSZXF1ZXN0VGFibGUiPg0KICAgIDx0cj4NCiAgICA8dGQgY2xhc3M9
Ik1lZXRpbmdSZXF1ZXN0VGltZUxvY2F0aW9uQ29udGFpbmVyIj48ZGl2PuatpOa0u+WKqOWPkeeU
n+S6jiAoR01UKzA4OjAwKSDljJfkuqzvvIzph43luobvvIzpppnmuK/nibnliKvooYzmlL/ljLrv
vIzkuYzpsoHmnKjpvZA8L2Rpdj48L3RkPg0KICAgIDwvdHI+DQo8L3RhYmxlPg0KPGRpdiBjbGFz
cz0iTWVldGluZ1JlcXVlc3RIUnVsZSI+PC9kaXY+DQoNCiAgICAgICAgICAgIDwvZGl2Pg0KICAg
ICAgICAgICAgPGJyIC8+DQogICAgICAgICAgICA8ZGl2IGNsYXNzPSJGb290ZXJDb250YWluZXIi
Pg0KICAgICAgICAgICAgICAgIDxpbWcgc3JjPSJodHRwczovL2dmeDUuaG90bWFpbC5jb20vY2Fs
LzExLjAwL3VwZGF0ZWJldGEvbHRyL2xvZ29fd2xfaG90bWFpbF8xMjAuZ2lmIiBhbHQ9IldpbmRv
d3MgTGl2ZSI+PC9pbWc+IA0KICAgICAgICAgICAgPC9kaXY+DQogICAgICAgIDwvdGQ+DQogICAg
ICAgIDx0ZCBjbGFzcz0iSW5uZXJWZXJ0aWNhbEJvcmRlciI+PC90ZD4NCiAgICAgICAgPHRkIGNs
YXNzPSJNYWluVmVydGljYWxCb3JkZXIiPjwvdGQ+DQogICAgPC90cj4NCiAgICA8dHIgY2xhc3M9
IklubmVySG9yaXpvbnRhbEJvcmRlciI+DQogICAgICAgIDx0ZCBjbGFzcz0iTWFpblZlcnRpY2Fs
Qm9yZGVyIj48L3RkPg0KICAgICAgICA8dGQgY2xhc3M9IklubmVyVmVydGljYWxCb3JkZXIiPjwv
dGQ+DQogICAgICAgIDx0ZCBjbGFzcz0iQ29udGVudENvbnRhaW5lciI+PC90ZD4NCiAgICAgICAg
PHRkIGNsYXNzPSJJbm5lclZlcnRpY2FsQm9yZGVyIj48L3RkPg0KICAgICAgICA8dGQgY2xhc3M9
Ik1haW5WZXJ0aWNhbEJvcmRlciI+PC90ZD4NCiAgICA8L3RyPg0KICAgIDx0ciBjbGFzcz0iTWFp
bkhvcml6b250YWxCb3JkZXIiPg0KICAgICAgICA8dGQgY2xhc3M9Ik1haW5WZXJ0aWNhbEJvcmRl
ciI+PC90ZD4gICAgDQogICAgICAgIDx0ZCBjbGFzcz0iSW5uZXJWZXJ0aWNhbEJvcmRlcldpZHRo
Ij48L3RkPg0KICAgICAgICA8dGQgY2xhc3M9IkNvbnRlbnRXaWR0aCI+PC90ZD4NCiAgICAgICAg
PHRkIGNsYXNzPSJJbm5lclZlcnRpY2FsQm9yZGVyV2lkdGgiPjwvdGQ+DQogICAgICAgIDx0ZCBj
bGFzcz0iTWFpblZlcnRpY2FsQm9yZGVyIj48L3RkPiAgICAgICAgDQogICAgPC90cj4NCjwvdGFi
bGU+DQo=

--_32936406-5bb3-4fc0-b17a-11cc15875002_
Content-Type: text/calendar; charset="utf-8"; method=REQUEST
Content-Transfer-Encoding: base64

QkVHSU46VkNBTEVOREFSDQpNRVRIT0Q6UkVRVUVTVA0KVkVSU0lPTjoyLjANClBST0RJRDotLy9N
aWNyb3NvZnQgQ29ycG9yYXRpb24vL1dpbmRvd3MgTGl2ZSBDYWxlbmRhci8vRU4NCkJFR0lOOlZU
SU1FWk9ORQ0KVFpJRDpDaGluYSBTdGFuZGFyZCBUaW1lDQpCRUdJTjpTVEFOREFSRA0KRFRTVEFS
VDoyMDA4MDEwMVQwMDAwMDANClRaT0ZGU0VUVE86KzA4MDANClRaT0ZGU0VURlJPTTorMDgwMA0K
RU5EOlNUQU5EQVJEDQpFTkQ6VlRJTUVaT05FDQpCRUdJTjpWRVZFTlQNClVJRDpiZTcxYWUxNS01
NjY0LTQ1MWEtODRiMy1hMzFjM2Y1NWFjZmMNCkRUU1RBTVA6MjAxMzAzMThUMDYxNTMzWg0KQ0xB
U1M6UFVCTElDDQpYLU1JQ1JPU09GVC1DRE8tQlVTWVNUQVRVUzpCVVNZDQpUUkFOU1A6T1BBUVVF
DQpTRVFVRU5DRTowDQpEVFNUQVJUO1RaSUQ9Q2hpbmEgU3RhbmRhcmQgVGltZToyMDEzMDMxOVQw
OTAwMDANCkRURU5EO1RaSUQ9Q2hpbmEgU3RhbmRhcmQgVGltZToyMDEzMDMxOVQxMDAwMDANClNV
TU1BUlk65rS75YqoDQpMT0NBVElPTjrlnLDngrkNClBSSU9SSVRZOjANCkFUVEVOREVFO0NVVFlQ
RT1JTkRJVklEVUFMO1JPTEU9UkVRLVBBUlRJQ0lQQU5UO1BBUlRTVEFUPU5FRURTLUFDVElPTjtS
U1ZQPQ0KIFRSVUU6TUFJTFRPOmdvb2dvbGxlZUBnbWFpbC5jb20NCkFUVEVOREVFO0NVVFlQRT1J
TkRJVklEVUFMO1JPTEU9UkVRLVBBUlRJQ0lQQU5UO1BBUlRTVEFUPU5FRURTLUFDVElPTjtSU1ZQ
PQ0KIFRSVUU6TUFJTFRPOnBhbmRhQDBkMGYuY29tDQpPUkdBTklaRVI7Q049TGVlIEdvb2dvbCBM
ZWU6TUFJTFRPOmdvb2dvbGxlZUBob3RtYWlsLmNvbQ0KQkVHSU46VkFMQVJNDQpBQ1RJT046RElT
UExBWQ0KVFJJR0dFUjotUFQxNU0NCkVORDpWQUxBUk0NCkJFR0lOOlZBTEFSTQ0KQUNUSU9OOkRJ
U1BMQVkNClRSSUdHRVI6LVBUMTVNDQpFTkQ6VkFMQVJNDQpFTkQ6VkVWRU5UDQpFTkQ6VkNBTEVO
REFSDQo=

--_32936406-5bb3-4fc0-b17a-11cc15875002_--`

	buf := bytes.NewBufferString(str)
	msg, err := mail.ReadMessage(buf)
	if err != nil {
		t.Fatal(err)
	}
	parser, err := NewParser(msg, &config)
	if err != nil {
		t.Fatal(err)
	}
	t.Errorf("%+v", parser)
}
