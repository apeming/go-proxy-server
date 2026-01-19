package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	// Dummy hash for timing attack protection (generated at init)
	dummyHash []byte
)

func init() {
	// Generate dummy hash at initialization for timing attack protection
	// This prevents attackers from distinguishing between valid and invalid usernames
	var err error
	dummyHash, err = bcrypt.GenerateFromPassword([]byte(""), bcrypt.DefaultCost)
	if err != nil {
		// Fallback to a pre-computed hash if generation fails
		dummyHash = []byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")
	}
}

// VerifyCredentials verifies username and password against stored credentials
// Uses constant-time comparison to prevent timing attacks
func VerifyCredentials(username string, password []byte) error {
	// Lock-free read using atomic.Value - no type assertion overhead
	creds := getCredentials()
	expectedPassword, ok := creds[username]

	// To prevent timing attacks, always perform bcrypt comparison
	// even if username doesn't exist. Use the dynamically generated dummy hash.
	if !ok {
		// Use the dummy hash generated at init time
		// This ensures consistent timing regardless of username existence
		bcrypt.CompareHashAndPassword(dummyHash, password)
		return fmt.Errorf("invalid credentials")
	}

	// Compare the received password with the expected password
	if err := bcrypt.CompareHashAndPassword(expectedPassword, password); err != nil {
		return fmt.Errorf("invalid credentials")
	}

	return nil
}
