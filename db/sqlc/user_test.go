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
	arg := CreateUserParams{
		FullName:       fmt.Sprintf("%s %s", util.RandomString(6), util.RandomString(6)),
		HashedPassword: util.RandomPasswordHash(12),
		Email:          util.RandomEmail(),
	}

	user, err := testQuerier.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

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
	user1 := createTestUser(t)

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
	user := createTestUser(t)

	user1, err := testQuerier.GetUser(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user1)

	require.Equal(t, user.ID, user1.ID)
	require.Equal(t, user.FullName, user1.FullName)
	require.Equal(t, user.Email, user1.Email)
	require.Equal(t, user.HashedPassword, user1.HashedPassword)
	require.Equal(t, user.IsActivated, user1.IsActivated)

	require.WithinDuration(t, user.CreatedAt, user1.CreatedAt, time.Second)
	require.WithinDuration(t, user.PasswordChangedAt, user1.PasswordChangedAt, time.Second)
}

// TODO: Add change password for user
