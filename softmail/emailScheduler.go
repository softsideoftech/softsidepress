package softmail

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"
)

func (ctx RequestContext) StartEmailScheduler() {
	defer ctx.runEmailAgain()
	if ctx.DevMode {
		time.Sleep(time.Second)
	} else {
		time.Sleep(time.Minute * 5)
	}

	cohorts, err := ctx.GetCurrentCohorts()
	if err != nil {
		ctx.SendOwnerErrorEmail("Problem retrieving current course cohorts.", err)
	}
	//log.Printf("DEBUG: Course cohorts: %v", cohorts)

	for _, courseCohort := range cohorts {
		course := ctx.GetCourse(courseCohort.CourseName)
		for _, session := range course.Sessions {

			memberLocations, err := ctx.GetCohortMemberLocations(courseCohort.Name)

			for _, memberLocation := range memberLocations {

				
				var memberTime = SystemTime{time.Now()}.GetMemberTime(memberLocation)
//				log.Printf("DEBUG member time: %v", memberTime)
				memberHour := memberTime.Hour()
				sendHour := course.Emails.SendHour
				//log.Printf("DEBUG: emailController - memberHour: %d, sendHour: %d", memberHour, sendHour)
				if memberHour == sendHour {

					// If today is the day for this session in this cohort, 
					// then see if we have any emails to send
					courseDay := courseCohort.GetCourseDay(memberTime)
					sessionDay := session.Day

					//log.Printf("DEBUG: emailController - courseDay: %d, sessionDay: %d", courseDay, sessionDay)
					if sessionDay == courseDay && courseDay > 0 {

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

						if err != nil {
							panic(fmt.Sprintf("ERROR: Unable to send course email due to a problem obtaining cohort member locations (timezones). Course: %s, Cohort: %s, Message: %v", course.Name, courseCohort.Name, err))
						}

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
