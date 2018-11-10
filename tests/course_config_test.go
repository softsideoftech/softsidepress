package softside

import (
	"softside/softmail"
	"testing"
	"time"
)

func TestCourseYmlUnmarshall(t *testing.T) {
	// TODO: refactor this so we can actually run this test
	ctx := softmail.NewRawRequestCtx()
	ctx.ContentPath = "/Users/vlad/go/src/softside/tests" 

	if ctx.DevMode {
		c1 := ctx.GetCourse("sampleCourseOne")
		c2 := ctx.GetCourse("sampleCourseTwo")
		if c1.Sessions[0].Day != 1 ||
			c1.Sessions[1].Day != 2 ||
			c1.Sessions[0].Name != "abe" ||
			c1.Sessions[1].Name != "bob" ||
			c1.Emails.SendHour != 1 {
			t.Errorf("Failed to parse config for sampleCourseOne: %v", c1)
		}
		if c2.Sessions[0].Day != 3 ||
			c2.Sessions[1].Day != 4 ||
			c2.Sessions[0].Name != "cat" ||
			c2.Sessions[1].Name != "dan" ||
			c2.Emails.SendHour != 2 {
			t.Errorf("Failed to parse config for sampleCourseTwo: %v", c2)
		}
	}
}

func TestGetCourseDay(t *testing.T) {
	day := time.Hour * 24
	fiveDays := day * 5
	fiveDaysAgo := time.Now().Add(fiveDays * -1)
	c := softmail.CourseCohort{
		StartDate: fiveDaysAgo,
	}

	
	now := time.Now()
	doCourseDayTest(c, now, t)

	memberLocation := softmail.ListMemberLocation{
		TimeZone: "+11:00",
	}
	systemTime := softmail.SystemTime{now}
	memberTime := systemTime.GetMemberTime(memberLocation)
	doCourseDayTest(c, memberTime, t)
}

func doCourseDayTest(c softmail.CourseCohort, currentTime time.Time, t *testing.T) {
	courseDay := c.GetCourseDay(currentTime)
	// Note, we consider the 0th day to be "course day 1".
	expectedCourseDay := 6
	if courseDay != expectedCourseDay {
		t.Errorf("Expected courseDay to be %d, but instead was %d.", expectedCourseDay, courseDay)
	}
}