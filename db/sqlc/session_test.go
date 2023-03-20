package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/OCD-Labs/KeyKeeper/internal/token"
	"github.com/OCD-Labs/KeyKeeper/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createTestSession(t *testing.T) Session {
	user := createTestUser(t)

	maker, err := token.NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	token, payload, err := maker.CreateToken(time.Minute, user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	id, err := uuid.NewRandom()
	require.NoError(t, err)

	ip := fmt.Sprintf(
		"%d.%d.%d.%d",
		util.RandomNumber(0, 255),
		util.RandomNumber(0, 255),
		util.RandomNumber(0, 255),
		util.RandomNumber(0, 255),
	)

	arg := CreateSessionParams{
		ID:           id,
		UserID:       user.ID,
		RefreshToken: token,
		UserAgent:    util.RandomString(9),
		ClientIp:     ip,
		IsBlocked:    false,
		ExpiresAt:    payload.ExpiredAt,
	}

	session, err := testQuerier.CreateSession(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	require.Equal(t, arg.ID, session.ID)
	require.Equal(t, user.ID, session.UserID)
	require.Equal(t, arg.RefreshToken, session.RefreshToken)
	require.Equal(t, ip, session.ClientIp)
	require.False(t, session.IsBlocked)
	require.WithinDuration(t, payload.ExpiredAt, session.ExpiresAt, time.Second)

	return session
}

func TestCreateSession(t *testing.T) {
	createTestSession(t)
}

func TestGetSession(t *testing.T) {
	session := createTestSession(t)

	session1, err := testQuerier.GetSession(context.Background(), session.ID)
	require.NoError(t, err)
	require.NotEmpty(t, session1)

	require.Equal(t, session.ID, session1.ID)
	require.Equal(t, session.UserID, session1.UserID)
	require.Equal(t, session.ClientIp, session1.ClientIp)
	require.Equal(t, session.UserAgent, session1.UserAgent)
	require.Equal(t, session.RefreshToken, session1.RefreshToken)
	require.Equal(t, session.IsBlocked, session1.IsBlocked)
	require.WithinDuration(t, session.CreatedAt, session1.CreatedAt, time.Second)
	require.WithinDuration(t, session.ExpiresAt, session1.ExpiresAt, time.Second)
}
