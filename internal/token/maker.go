package token

import "time"

// A TokenMaker is an interface for managing tokens.
type TokenMaker interface {
	// CreateToken creates a new specific token for a user ID and duration
	CreateToken(duration time.Duration, userID int64) (string, *Payload, error)

	// VerifyToken verifies if a token is valid or not.
	VerifyToken(token string) (*Payload, error)
}
