package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 parameters
	saltLen   = 32
	keyLen    = 32
	timeParam = 1
	memory    = 64 * 1024 // 64 MB
	threads   = 4
)

// PasswordConfig holds the argon2id configuration
type PasswordConfig struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

// DefaultPasswordConfig returns a secure default configuration
func DefaultPasswordConfig() PasswordConfig {
	return PasswordConfig{
		Time:    timeParam,
		Memory:  memory,
		Threads: threads,
		KeyLen:  keyLen,
		SaltLen: saltLen,
	}
}

// HashPassword hashes a password using argon2id
func HashPassword(password string) (string, error) {
	config := DefaultPasswordConfig()

	salt := make([]byte, config.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, config.Time, config.Memory, config.Threads, config.KeyLen)

	// Encode the hash in a format: $argon2id$v=19$m=memory,t=time,p=threads$salt$hash
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		config.Memory, config.Time, config.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))

	return encoded, nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(password, hashedPassword string) (bool, error) {
	// Parse the encoded hash
	salt, hash, config, err := parseHash(hashedPassword)
	if err != nil {
		return false, fmt.Errorf("invalid hash format: %w", err)
	}

	// Hash the input password with the same salt and config
	testHash := argon2.IDKey([]byte(password), salt, config.Time, config.Memory, config.Threads, config.KeyLen)

	// Compare using constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(hash, testHash) == 1, nil
}

// parseHash parses an encoded argon2id hash
func parseHash(encoded string) (salt, hash []byte, config PasswordConfig, err error) {
	var version int
	var memory, time uint32
	var threads uint8
	var saltB64, hashB64 string

	// Parse the format: $argon2id$v=19$m=memory,t=time,p=threads$salt$hash
	n, err := fmt.Sscanf(encoded, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		&version, &memory, &time, &threads, &saltB64, &hashB64)
	if err != nil || n != 6 {
		return nil, nil, config, fmt.Errorf("invalid hash format")
	}

	if version != 19 {
		return nil, nil, config, fmt.Errorf("unsupported argon2 version: %d", version)
	}

	salt, err = base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return nil, nil, config, fmt.Errorf("invalid salt encoding: %w", err)
	}

	hash, err = base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return nil, nil, config, fmt.Errorf("invalid hash encoding: %w", err)
	}

	// Validate lengths to prevent integer overflow
	if len(hash) > 0xFFFFFFFF || len(salt) > 0xFFFFFFFF {
		return nil, nil, config, fmt.Errorf("hash or salt length exceeds maximum allowed size")
	}

	config = PasswordConfig{
		Time:    time,
		Memory:  memory,
		Threads: threads,
		KeyLen:  uint32(len(hash)),  // #nosec G115 - bounds checked above
		SaltLen: uint32(len(salt)), // #nosec G115 - bounds checked above
	}

	return salt, hash, config, nil
}
