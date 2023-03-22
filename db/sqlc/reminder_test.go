package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/OCD-Labs/KeyKeeper/internal/util"
	"github.com/stretchr/testify/require"
)

type extension struct {
	Region string `json:"region"`
	Action string `json:"action"`
}

// createTestReminder creates a test reminder with the given user ID and returns it
func createTestReminder(t *testing.T, userID int64) Reminder {
	// Define an extension with some region and action
	ext := extension{
		Region: "Europe",
		Action: "Create reminder",
	}

	// Marshal the extension into a JSON buffer
	buf, err := json.Marshal(ext)
	require.NoError(t, err)

	// Define arguments for creating a reminder
	arg := CreateReminderParams{
		UserID:     userID,
		WebsiteUrl: util.RandomWebsiteURL(),
		Interval:   "2 weeks",
		Extension:  buf,
	}

	// Call the CreateReminder function with the arguments
	reminder, err := testQuerier.CreateReminder(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, reminder)

	// Check that the reminder was created correctly
	require.NotZero(t, reminder.ID)
	require.Equal(t, userID, reminder.UserID)
	require.Equal(t, arg.WebsiteUrl, reminder.WebsiteUrl)
	require.Equal(t, arg.Interval, reminder.Interval)
	require.NotZero(t, reminder.UpdatedAt)

	// Unmarshal the reminder's extension into an extension
	// struct and check that it matches the original extension
	var ext1 extension
	err = json.Unmarshal(reminder.Extension, &ext1)
	require.NoError(t, err)
	require.Equal(t, ext, ext1)

	// Return the created reminder
	return reminder
}

func TestCreateReminder(t *testing.T) {
	user := createTestUser(t)
	createTestReminder(t, user.ID)
}

func TestDeleteReminder(t *testing.T) {
	// Create a test user and reminder.
	user := createTestUser(t)
	reminder := createTestReminder(t, user.ID)

	// Define arguments for the DeleteReminder function
	arg := DeleteReminderParams{
		ID:         reminder.ID,
		WebsiteUrl: reminder.WebsiteUrl,
	}

	// Call the DeleteReminder function and check for errors
	err := testQuerier.DeleteReminder(context.Background(), arg)
	require.NoError(t, err)

	// Define arguments for the GetReminder function
	arg1 := GetReminderParams{
		ID:         reminder.ID,
		WebsiteUrl: reminder.WebsiteUrl,
	}

	// Call the GetReminder function and check for errors
	reminder1, err := testQuerier.GetReminder(context.Background(), arg1)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, reminder1) // Check that the reminder has been deleted
}

func TestGetReminder(t *testing.T) {
	// Create a test user and a test reminder associated with that user.
	user := createTestUser(t)
	reminder := createTestReminder(t, user.ID)

	// Prepare the arguments to call GetReminder.
	arg := GetReminderParams{
		ID:         reminder.ID,
		WebsiteUrl: reminder.WebsiteUrl,
	}

	// Call GetReminder and check for errors.
	reminder1, err := testQuerier.GetReminder(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, reminder1)

	// Check that the returned reminder matches the expected values.
	require.Equal(t, reminder.ID, reminder1.ID)
	require.Equal(t, reminder.UserID, reminder1.UserID)
	require.Equal(t, reminder.WebsiteUrl, reminder1.WebsiteUrl)
	require.Equal(t, reminder.Interval, reminder1.Interval)
	require.Equal(t, reminder.UpdatedAt, reminder1.UpdatedAt)
	require.Equal(t, reminder.Extension, reminder1.Extension)
}

func TestListReminders(t *testing.T) {
	// Create a test user.
	user := createTestUser(t)

	// Create 5 test reminders for the user.
	for i := 0; i < 5; i++ {
		createTestReminder(t, user.ID)
	}

	// Set up the arguments for ListReminders.
	arg := ListRemindersParams{
		UserID: user.ID,
		Limit:  5,
		Offset: 0,
	}

	// Call ListReminders and make sure there are 5 reminders.
	reminders, err := testQuerier.ListReminders(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, reminders, 5)

	// Check that each reminder belongs to the test user.
	for _, reminder := range reminders {
		require.NotEmpty(t, reminder)
		require.Equal(t, user.ID, reminder.UserID)
	}
}

func TestSetNewInterval(t *testing.T) {
	// Create a test user and reminder.
	user := createTestUser(t)
	reminder := createTestReminder(t, user.ID)

	// Set the arguments for the SetNewInterval function.
	arg := SetNewIntervalParams{
		NewInterval: "1 month",
		ID:          reminder.ID,
		WebsiteUrl:  reminder.WebsiteUrl,
	}

	// Call the SetNewInterval function and get the new reminder.
	reminder1, err := testQuerier.SetNewInterval(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, reminder1)

	// Check that there are expected changes.
	require.Equal(t, reminder.ID, reminder.ID)
	require.Equal(t, user.ID, reminder.UserID)

	require.Equal(t, arg.NewInterval, reminder1.Interval)
}

func TestUpdateReminderConfigs(t *testing.T) {
	// Create a test user and reminder.
	user := createTestUser(t)
	reminder := createTestReminder(t, user.ID)

	// Get the reminder configurations.
	reminderExt, err := testQuerier.GetReminderConfigs(context.Background(), GetReminderConfigsParams{
		ID:         reminder.ID,
		WebsiteUrl: reminder.WebsiteUrl,
	})
	require.NoError(t, err)
	require.NotEmpty(t, reminderExt)

	// Unmarshal the reminder extensions.
	ext := extension{}
	require.NoError(t, json.Unmarshal(reminderExt, &ext))

	// Update the reminder extensions.
	ext.Action = "Updating configs"
	ext.Region = "Africa"
	buf, err := json.Marshal(ext)
	require.NoError(t, err)

	// Set the reminder configurations.
	reminder1, err := testQuerier.SetReminderConfigs(context.Background(), SetReminderConfigsParams{
		ID:               reminder.ID,
		WebsiteUrl:       reminder.WebsiteUrl,
		UpdatedExtension: buf,
	})
	require.NoError(t, err)
	require.NotEmpty(t, reminder1)

	// Unmarshal the updated reminder extensions.
	ext1 := extension{}
	require.NoError(t, json.Unmarshal(reminder1.Extension, &ext1))

	// Check that the updated extensions match the original extensions.
	require.Equal(t, ext, ext1)
}

func TestUpdateReminder(t *testing.T) {
	// Create a test user and reminder.
	user := createTestUser(t)
	reminder := createTestReminder(t, user.ID)

	// Set the arguments for the UpdateReminder function.
	arg := UpdateReminderParams{
		ID:         reminder.ID,
		UpdatedAt:  time.Now(),
		WebsiteUrl: reminder.WebsiteUrl,
	}

	// Call the UpdateReminder function and get the updated reminder.
	reminder1, err := testQuerier.UpdateReminder(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, reminder1)

	// Check that there are expected changes.
	require.Equal(t, reminder.ID, reminder1.ID)
	require.Equal(t, reminder.WebsiteUrl, reminder1.WebsiteUrl)
	require.WithinDuration(t, reminder.UpdatedAt, reminder1.UpdatedAt, time.Second)
}
