package softmail

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"
)

func (ctx RequestContext) StartEmailScheduler() {
	defer ctx.runEmailAgain()
	if ctx.DevMode{
		time.Sleep(time.Second)
	} else {
		time.Sleep(time.Minute * 5)
	}

	cohorts, err := ctx.GetCurrentCohorts()
	if err != nil {
		ctx.SendOwnerErrorEmail("Problem retrieving current course cohorts.", err)
	}

	for _, courseCohort := range cohorts {
		course := ctx.GetCourse(courseCohort.CourseName)
		for _, session := range course.Sessions {
			// If today is the day for this session in this cohort, 
			// then see if we have any emails to send
			courseDay := courseCohort.GetCourseDay()
			sessionDay := session.Day
			if sessionDay == courseDay {

				subject := fmt.Sprintf("%s Day %d: %s", course.Shortname, sessionDay, session.Name)
				opts := SendEmailOpts{DontDoubleSend: true, TemplateParams: session, Login: true}
				var emailTemplate string
				if session.VideoUrl != "" {
					// It's ok to use a relative URL here because it's first going to be 
					// turned into a fully qualified login URL (makes debugging easier)
					opts.DestinationUrl = course.Url + "/" + session.Url
					emailTemplate = courseVideoEmailTemplate
				} else {
					emailTemplate = courseDayEmailTemplate
				}

				memberLocations, err := ctx.GetCohortMemberLocations(courseCohort.Name)
				
				if err != nil {
					panic(fmt.Sprintf("ERROR: Unable to send course email due to a problem obtaining cohort member locations (timezones). Course: %s, Cohort: %s, Message: %v", course.Name, courseCohort.Name, err))
				}
				
				for _, memberLocation := range memberLocations {
					
					timeLayout := "Mon Jan 2 15:04:05 %s MST 2006"
					timeFormat := fmt.Sprintf(timeLayout, "-07:00")
					memberTimeStr := fmt.Sprintf(timeLayout, memberLocation.TimeZone)
					memberTime, _ := time.Parse(timeFormat, memberTimeStr)
					memberTimeLocation := memberTime.Location()
					
					timeAtLocation := time.Now().In(memberTimeLocation)

					if timeAtLocation.Hour() == course.Emails.SendHour {
						ctx.SendTemplatedEmailToId(subject, emailTemplate, memberLocation.Id, opts)
						// todo: make sending the same email idempotent is the DontDoubleSent opt is enabled
					}
				}

				
			}
		}
	}
}

func (ctx RequestContext) runEmailAgain() {
	stack := debug.Stack()
	if r := recover(); r != nil {
		log.Printf("ERROR scheduling email. %v\n%s", r, string(stack))
	}

	go ctx.StartEmailScheduler()
}

func (ctx RequestContext) SendOwnerErrorEmail(message string, err error) {
	// TODO
}
