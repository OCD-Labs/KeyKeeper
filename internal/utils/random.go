package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomNumber returns a random number between min
// and max integers.
func RandomNumber(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates random string value
func RandomString(length int) string {
	sb := strings.Builder{}
	k := len(alphabet)

	for i := 0; i < length; i++ {
		char := alphabet[rand.Intn(k)]
		sb.WriteByte(char)
	}

	return sb.String()
}

// RandomPasswordHash generates a random hash of a string.
func RandomPasswordHash(passwordLength int) string {
	stringVal, err := HashedPassword(RandomString(passwordLength))
	if err != nil {
		return ""
	}
	return stringVal
}

// RandomEmail returns a random email.
func RandomEmail() string {
	return fmt.Sprintf("%s@gmail.com", RandomString(10))
}

// RandomWebsiteURL returns a random website url.
func RandomWebsiteURL() string {
	return fmt.Sprintf("%s.%s", RandomString(6), RandomString(3))
}
