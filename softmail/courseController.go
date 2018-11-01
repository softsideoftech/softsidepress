package softmail

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

type Session struct {
	Name        string
	Day         int
	Description string
	Url         string
	VideoUrl    string
	Course      *CourseConfig
}

type Emails struct {
	SendHour int
}

type CourseConfig struct {
	Name     string
	Sessions []*Session
	Emails   Emails
	Url      string
}

type ConfigObj interface {
}

type CoursePageParams struct {
	TrackingRequestParams
	ConfigObj ConfigObj
	Url       string
	CourseDay int
}

type NotLoggedInError struct {
	msg string
}

type NoSuchCourseError struct {
	msg string
}

type NotRegisteredForCourseError struct {
	msg string
}

type CourseNotStartedError struct {
	msg       string
	StartDate time.Time
}

var confMux = &sync.Mutex{}
var courses map[string]CourseConfig = nil

func (e NotLoggedInError) Error() string {
	return e.msg
}
func (e NoSuchCourseError) Error() string {
	return e.msg
}
func (e CourseNotStartedError) Error() string {
	return e.msg
}
func (e NotRegisteredForCourseError) Error() string {
	return e.msg
}

func loadCourses(coursesDirPath string) map[string]CourseConfig {
	courses := make(map[string]CourseConfig)

	coursesDir, err := os.Stat(coursesDirPath)
	if coursesDir == nil || !coursesDir.IsDir() {
		log.Printf("Didn't find a directory of courses named: %s. Error message: %v\n", coursesDirPath, err)
		return courses
	}

	courseFiles, err := ioutil.ReadDir(coursesDirPath)
	if courseFiles == nil || len(courseFiles) == 0 {
		log.Printf("Didn't find any courses in the directory: %s\n", coursesDirPath)
		return courses
	}

	for _, curCourseDir := range courseFiles {
		coursePathName := curCourseDir.Name()
		courseCfgPath := coursesDirPath + "/" + coursePathName + "/course.yml"
		courseCfgBytes, err := ioutil.ReadFile(courseCfgPath)
		if err != nil {
			log.Printf("ERROR reading course config file: %s, error: %v", courseCfgPath, err)
		}
		var course CourseConfig
		err = yaml.Unmarshal(courseCfgBytes, &course)
		if err != nil {
			log.Printf("ERROR parsing course config file for course: %s, error: %v", coursePathName, err)
		}

		for _, session := range course.Sessions {
			session.VideoUrl = fmt.Sprintf("%s/courses/%s/%s", CDNUrl, coursePathName, session.Url)
			session.Course = &course
		}
		course.Url = "/" + coursePathName

		courses[coursePathName] = course
	}

	return courses
}

func (ctx *RequestContext) GetCourseForCurListMember(courseName string) (*CourseConfig, *CourseCohort, error) {

	course := ctx.GetCourse(courseName)

	if course == nil {
		return nil, nil, NoSuchCourseError{"There is no course named: " + courseName}
	}

	if ctx.MemberCookie == nil || ctx.MemberCookie.ListMemberId == 0 || ctx.MemberCookie.LoggedIn == nil {
		return course, nil, NotLoggedInError{"No logged in user for current request"}
	}

	// Try to find a CourseCohort that matches this ListMemberId and CourseName
	var courseCohort CourseCohort
	_, err := ctx.DB.Query(&courseCohort, `
			select c.* from member_groups g, course_cohorts c 
			where g.name = c.name and c.course_name = ? and g.list_member_id = ?`,
		courseName, ctx.MemberCookie.ListMemberId)

	if err != nil {
		panic(fmt.Sprintf("DB problem while retrieving course cohort named: %s for user: %d", courseName, ctx.MemberCookie.ListMemberId))
	}
	
	if courseCohort.Name == "" {

		// See if this course exists at all
		_, err := ctx.DB.Query(&courseCohort, "select c.* from course_cohorts c where c.course_name = ? limit 1", courseName)

		if err != nil {
			panic(fmt.Sprintf("DB problem while retrieving course cohort named: %s", courseName))
		}
		
		if len(courseCohort.Name) > 0 {
			// If at least one such CourseCohort exists, it may be that the person isn't signed up for the course or is logged in with the wrong email.
			return course, nil, NotRegisteredForCourseError{fmt.Sprintf("Not registered for course named: %s for user: %d ", courseName, ctx.MemberCookie.ListMemberId)}	
		} else {
			// Otherwise, the course simply doesn't exist. 
			return course, nil, NoSuchCourseError{fmt.Sprintf("No started course named: %s for user: %d ", courseName, ctx.MemberCookie.ListMemberId)}
		}
	}

	if courseCohort.StartDate.After(time.Now()) {
		return nil, nil, CourseNotStartedError{
			msg:       fmt.Sprintf("Course doesn't start until a future date for cohort: %s", courseCohort.Name),
			StartDate: courseCohort.StartDate}
	}

	return course, &courseCohort, nil
}

func (ctx *RequestContext) GetCourse(courseName string) *CourseConfig {
	confMux.Lock()
	defer confMux.Unlock()
	if courses == nil || ctx.DevMode {
		courses = loadCourses(ctx.GetFilePath("/courses"))
	}
	course := courses[courseName]
	return &course
}

func (courseConfig *CourseConfig) getSession(sessionUrlName string) *Session {
	for _, session := range courseConfig.Sessions {
		if session.Url == sessionUrlName {
			return session
		}
	}
	return nil
}
