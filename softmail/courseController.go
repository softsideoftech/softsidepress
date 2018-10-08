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
	Name string
	Day  int
}

type Emails struct {
	SendHour int
}

type CourseConfig struct {
	Name     string
	Sessions []Session
	Emails   Emails
}

type CourseParams struct {
	TrackingRequestParams
	*CourseConfig
	Url string
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
		courseName := curCourseDir.Name()
		courseCfgPath := coursesDirPath + "/" + courseName + "/course.yml"
		courseCfgBytes, err := ioutil.ReadFile(courseCfgPath)
		if err != nil {
			log.Printf("ERROR reading course config file: %s, error: %v", courseCfgPath, err)
		}
		var course CourseConfig
		err = yaml.Unmarshal(courseCfgBytes, &course)
		if err != nil {
			log.Printf("ERROR parsing course config file for course: %s, error: %v", courseName, err)
		}
		courses[courseName] = course
	}

	return courses
}

func (ctx *RequestContext) GetCourseForCurListMember(courseName string) (*CourseConfig, error) {

	course := ctx.GetCourse(courseName)

	if course == nil {
		return nil, NoSuchCourseError{"There is no course named: " + courseName}
	}

	if ctx.MemberCookie == nil || ctx.MemberCookie.ListMemberId == 0 || ctx.MemberCookie.LoggedIn != nil {
		return course, NotLoggedInError{"No logged in user for current request"}
	}

	var courseCohort CourseCohort

	ctx.DB.Query(&courseCohort, `
	select c.* from member_groups g, course_cohorts c 
	where g.name = c.cohort_name and c.course_name = ? and g.list_member_id = ?`,
		courseName, ctx.MemberCookie.ListMemberId)

	if courseCohort.Name == "" {
		return course, NoSuchCourseError{fmt.Sprintf("No started course named: %s for user: %d ", courseName, ctx.MemberCookie.ListMemberId)}
	}

	if courseCohort.StartDate.After(time.Now()) {
		return nil, CourseNotStartedError{
			msg:       fmt.Sprintf("Course doesn't start until a future date for cohort: %s", courseCohort.Name),
			StartDate: courseCohort.StartDate}
	}

	return course, nil
}

func (ctx *RequestContext) GetCourse(courseName string) *CourseConfig {
	confMux.Lock()
	defer confMux.Unlock()
	if courses == nil {
		courses = loadCourses(ctx.GetFilePath("/courses"))
	}
	course := courses[courseName]
	return &course
}

func (ctx *RequestContext) GetCoursePageParams(coursePath string, trackingParams TrackingRequestParams) (*CourseParams, error) {

	courseConfig, err := ctx.GetCourseForCurListMember(coursePath)

	return &CourseParams{
		trackingParams,
		courseConfig,
		"/" + coursePath,
	}, err
}
