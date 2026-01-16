package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters - recommended by OWASP
	argon2Time      = 1
	argon2Memory    = 64 * 1024 // 64 MB
	argon2Threads   = 4
	argon2KeyLength = 32
	saltLength      = 16
)

// HashPassword creates an Argon2id hash of the password
func HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate the hash
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLength)

	// Encode salt and hash to base64
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads, saltEncoded, hashEncoded), nil
}

// VerifyPassword verifies a password against an Argon2id hash
func VerifyPassword(password, hash string) error {
	if !strings.HasPrefix(hash, "$argon2id$") {
		return fmt.Errorf("unsupported hash format")
	}
	return verifyArgon2Password(password, hash)
}

func verifyArgon2Password(password, encodedHash string) error {
	// Parse the encoded hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return fmt.Errorf("invalid hash format")
	}

	// parts[0] is empty, parts[1] is "argon2id", parts[2] is version
	// parts[3] contains parameters, parts[4] is salt, parts[5] is hash

	// Parse parameters
	var memory, time uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode hash
	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("failed to decode hash: %w", err)
	}

	// Generate hash from password
	passwordHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(decodedHash)))

	// Compare using constant time comparison
	if subtle.ConstantTimeCompare(passwordHash, decodedHash) == 1 {
		return nil
	}

	return fmt.Errorf("invalid password")
}