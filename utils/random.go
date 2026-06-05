package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateRandomString generates a cryptographically secure random string of a given length.
// We use this to assign strong, random passwords to users signing in via Google.
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		// Securely pick a random index from our charset
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback character if something goes wrong (very rare)
			result[i] = charset[0]
			continue
		}
		result[i] = charset[num.Int64()]
	}

	return string(result)
}
