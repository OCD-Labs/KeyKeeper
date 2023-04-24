package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	// Authentication server address.
	smtpAuthAddress = "smtp.gmail.com"

	// Gmail server address.
	smtpServerAddress = "smtp.gmail.com:587"
)

// EmailSender defines the required interface to send a mail.
type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

// GmailSender represents the gmail sender's details
type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

// NewGmailSender instantiates a new GmailSender
func NewGmailSender(name string, fromEmailAddress string, fromEmailPassword string) EmailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

// SendEmail encapsulates the logic needed to send mail.
func (sender *GmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPassword, smtpAuthAddress)
	return e.Send(smtpServerAddress, smtpAuth)
}
