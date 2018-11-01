package softmail

import "time"

type ListMemberId = uint32
type ListMember struct {
	Id           ListMemberId
	FirstName    string
	LastName     string
	Company      string
	Position     string
	Email        string
	MemberRole   string
	Created      time.Time
	Updated      time.Time
	Subscribed   *time.Time
	Unsubscribed *time.Time
}

type EmailTemplateId = int64
type EmailTemplate struct {
	Id      EmailTemplateId
	Subject string
	Body    string
	Created time.Time
}

type SentEmailId = uint32
type SentEmail struct {
	Id              SentEmailId
	EmailTemplateId EmailTemplateId
	ListMemberId    ListMemberId
	Created         time.Time
	ThirdPartyId    string
}

type MemberCookieId = int64
type MemberCookie struct {
	Id           MemberCookieId
	ListMemberId ListMemberId
	Created      time.Time
	Updated      time.Time
	LoggedIn     *time.Time
}

type TrackedUrlId = int64
type TrackedUrl struct {
	Id          TrackedUrlId
	SentEmailId SentEmailId
	Url         string
	TargetUrl   string
	Created     time.Time
	LoginId 	ListMemberId
}

type TrackingHitId = uint32
type IpAddress = int64

type TrackingHit struct {
	Id              TrackingHitId
	TrackedUrlId    TrackedUrlId
	ListMemberId    ListMemberId
	MemberCookieId  MemberCookieId
	IpAddress       IpAddress
	IpAddressString string
	ReferrerUrl     string
	Created         time.Time
}

type EmailActionId = uint32
type EmailActionEnum = string
type EmailAction struct {
	Id          EmailActionId
	SentEmailId SentEmailId
	Action      EmailActionEnum
	Created     time.Time
	Metadata    string
}

type MemberGroupName = string
type MemberGroup struct {
	Name         MemberGroupName
	ListMemberId ListMemberId
	Created      time.Time
}

type CourseCohort struct {
	Name       MemberGroupName
	CourseName string
	StartDate  time.Time
	EndDate    time.Time
}
