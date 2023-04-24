package mailer

import (
	"testing"

	"github.com/OCD-Labs/KeyKeeper/internal/utils"
	"github.com/stretchr/testify/require"
)

// TestSendEmailWithGmail tests the send user email verification.
func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := utils.ParseConfigs("../..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A test email"
	content := `
	<h1>Hello!</h1>
	<p>This is a test message from <a href="https://github.com/OCD-Labs/">OCD Labs</a></p>
	`
	to := []string{"mrikehchukwuka@gmail.com"}
	// attachFiles := []string{"../../README.md"}

	err = sender.SendEmail(subject, content, to, nil, nil, nil)
	require.NoError(t, err)
}
