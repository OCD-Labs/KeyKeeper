package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/OCD-Labs/KeyKeeper/internal/util"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T) User {
	// Set up test user parameters.
	arg := CreateUserParams{
		FullName:       fmt.Sprintf("%s %s", util.RandomString(6), util.RandomString(6)),
		HashedPassword: util.RandomPasswordHash(12),
		Email:          util.RandomEmail(),
	}

	// Create the user.
	user, err := testQuerier.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	// Assert that the user was created with the correct properties.
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.True(t, user.IsActivated)
	require.Zero(t, user.PasswordChangedAt)
	require.NotZero(t, user.CreatedAt)

	return user
}
func TestCreateUser(t *testing.T) {
	createTestUser(t)
}

func TestDeactivateUser(t *testing.T) {
	// Create a test user.
	user1 := createTestUser(t)

	// Set up parameters to deactivate the test user.
	arg := DeactivateUserParams{
		ID:    user1.ID,
		Email: user1.Email,
	}

	// Deactivate the user and check for errors
	user, err := testQuerier.DeactivateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	// Check that the user is deactivated
	require.False(t, user.IsActivated)
}

func TestGetUser(t *testing.T) {
	// Create a test user.
	user := createTestUser(t)

	// Retrieve the test user from the database.
	user1, err := testQuerier.GetUser(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user1)

	// Assert that the retrieved user has the same properties as the test user.
	require.Equal(t, user.ID, user1.ID)
	require.Equal(t, user.FullName, user1.FullName)
	require.Equal(t, user.Email, user1.Email)
	require.Equal(t, user.HashedPassword, user1.HashedPassword)
	require.Equal(t, user.IsActivated, user1.IsActivated)

	// Assert that the retrieved user's timestamps are within one
	// second of the test user's timestamps.
	require.WithinDuration(t, user.CreatedAt, user1.CreatedAt, time.Second)
	require.WithinDuration(t, user.PasswordChangedAt, user1.PasswordChangedAt, time.Second)
}

func TestChangePassword(t *testing.T) {
	// Create a test user.
	user := createTestUser(t)

	// Set up parameters to change the test user password.
	arg := ChangePasswordParams{
		HashedPassword: util.RandomPasswordHash(16),
		Email:          user.Email,
	}

	// Change the user password and check for errors
	user1, err := testQuerier.ChangePassword(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user1)

	// Assert that the retrieved user has the same properties as
	// the test user.
	require.Equal(t, user.ID, user1.ID)
	require.Equal(t, user.Email, user1.Email)

	// Assert that the retrieved user's password changed.
	require.NotEqual(t, user.HashedPassword, user1.HashedPassword)
	require.Equal(t, arg.HashedPassword, user1.HashedPassword)
}

func TestChangeEmail(t *testing.T) {
	// Create a test user.
	user := createTestUser(t)

	// Set up parameter to change the test user email.
	arg := ChangeEmailParams{
		Email: util.RandomEmail(),
		ID:    user.ID,
	}

	// Change the user email and check for errors
	user1, err := testQuerier.ChangeEmail(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user1)

	// Assert that the retrieved user's email changed.
	require.Equal(t, user.ID, user1.ID)
	require.NotEqual(t, user.Email, user1.Email)
	require.Equal(t, arg.Email, user1.Email)
}
