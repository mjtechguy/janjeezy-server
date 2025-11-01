package password

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	saltLen             = 16
)

// Hash generates an Argon2id hash for the provided password.
func Hash(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", errors.New("password cannot be empty")
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemory, argonTime, argonThreads, encodedSalt, encodedHash), nil
}

// Verify compares a password with an encoded Argon2id hash.
func Verify(password, encodedHash string) (bool, error) {
	if strings.TrimSpace(encodedHash) == "" {
		return false, errors.New("encoded hash is empty")
	}

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		return false, errors.New("invalid hash format")
	}

	var memory uint32
	var time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, fmt.Errorf("invalid hash parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, fmt.Errorf("invalid salt: %w", err)
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid hash: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expectedHash)))

	if len(hash) != len(expectedHash) {
		return false, nil
	}

	var diff uint8
	for i := 0; i < len(hash); i++ {
		diff |= hash[i] ^ expectedHash[i]
	}

	return diff == 0, nil
}
