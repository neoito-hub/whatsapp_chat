package mailer

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	"github.com/appblocks-hub/SHIELD/functions/general"
)

type VerifyEmailUserStruct struct {
	Email            string
	UserName         string
	VerificationCode string
}

type VerifyEmailResStruct struct {
	Message string
	AppLink string
}

// Get the smtp auth data
func getSmtpAuth() (smtp.Auth, string, string) {
	// Sender data.

	from := general.Envs["SHIELD_MAILER_EMAIL"]
	password := general.Envs["SHIELD_MAILER_PASSWORD"]

	// smtp server configuration.
	smtpHost := general.Envs["SHIELD_MAILER_HOST"]
	smtpPort := general.Envs["SHIELD_MAILER_PORT"]

	smtpUrl := smtpHost + ":" + smtpPort

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	return auth, smtpUrl, from
}

// Function returns the email verification template with the provided user data
func getTemplateBody(userMailData VerifyEmailUserStruct) bytes.Buffer {
	t, _ := template.ParseFiles("static/src/templates/verifyEmailMailer.html")

	var body bytes.Buffer

	fromMail := fmt.Sprintf("From: Appblocks <%s>\r\n", general.Envs["SHIELD_MAILER_EMAIL"])
	toMail := fmt.Sprintf("To: <%s>\r\n", userMailData.Email)

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subjectData := fmt.Sprintf("Subject: Appblocks Email Verification \n%s\n\n", mimeHeaders)

	mailData := fromMail + toMail + subjectData

	fmt.Printf("mailData %s", mailData)

	body.Write([]byte(mailData))
	// body.Write([]byte(fmt.Sprintf(subjectData, mimeHeaders)))

	t.Execute(&body, userMailData)

	return body
}

// Send  email verification and return success/error message
func EmailVerficationMailer(userData VerifyEmailUserStruct) (string, error) {

	// Get mail template body
	to := []string{
		userData.Email,
	}

	// Get mail template body
	body := getTemplateBody(userData)

	// Get smtp auth data
	auth, smtpUrl, from := getSmtpAuth()

	// Sending email.
	err := smtp.SendMail(smtpUrl, auth, from, to, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return "Error", err
	}

	fmt.Println("Email Sent!")
	return "Done", err
}

func getPasswordRecoveryTemplateBody(userMailData general.ResetPasswordMail) bytes.Buffer {
	t, _ := template.ParseFiles("static/src/templates/passwordRecoveryMailer.html")

	var body bytes.Buffer

	fromMail := fmt.Sprintf("From: Appblocks <%s>\r\n", general.Envs["SHIELD_MAILER_EMAIL"])
	toMail := fmt.Sprintf("To: <%s>\r\n", userMailData.Email)

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subjectData := fmt.Sprintf("Subject: Appblocks Password Recovery \n%s\n\n", mimeHeaders)

	mailData := fromMail + toMail + subjectData

	fmt.Printf("mailData %s", mailData)

	body.Write([]byte(mailData))
	// body.Write([]byte(fmt.Sprintf(subjectData, mimeHeaders)))

	t.Execute(&body, userMailData)

	return body
}

func PasswordRecoveryMailer(userData general.ResetPasswordMail) (string, error) {

	// Get mail template body
	to := []string{
		userData.Email,
	}

	// Get mail template body
	body := getPasswordRecoveryTemplateBody(userData)

	// Get smtp auth data
	auth, smtpUrl, from := getSmtpAuth()

	// Sending email.
	err := smtp.SendMail(smtpUrl, auth, from, to, body.Bytes())
	if err != nil {
		fmt.Println(err)
		return "Error", err
	}

	fmt.Println("Email Sent!")
	return "Done", err
}
