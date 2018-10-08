package softside

import (
	"softside/softmail"
	"testing"
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