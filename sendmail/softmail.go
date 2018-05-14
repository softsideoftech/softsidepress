package sendmail

import (
	"softside/softmail"
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	subject := args[0]
	emailTemplateFile := args[1]
	fromEmail := args[2]
	memberGroupName := args[3]

	err := softmail.Sendmail(subject, emailTemplateFile, fromEmail, memberGroupName)
	if err != nil {
		fmt.Println(err)
	}
}
