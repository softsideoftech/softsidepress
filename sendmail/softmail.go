package main

import (
	"log"
	"os"
	"runtime/debug"
	"softside/softmail"
)

func main() {
	defer (func() {
		if r := recover(); r != nil {
			log.Printf("PANIC: %v\n", r, string(debug.Stack()))
		}
	})();
	args := os.Args[1:]
	subject := args[0]
	emailTemplateFile := args[1]
	memberEmailOrGroupName := args[2]
	login := args[3]
	suffix := args[4]

	softmail.NewRawRequestCtx().SendTemplatedEmail(
		subject,
		emailTemplateFile,
		memberEmailOrGroupName,
		softmail.SendEmailOpts{
			Login:     login == "login",
			UseSuffix: suffix == "suffix",
		})
}
