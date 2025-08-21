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
	saltLength = 16
	keyLength  = 32
	timeParam  = 1
	memory     = 64 * 1024
	threads    = 4
)

type HashParams struct {
	Memory  uint32
	Time    uint32
	Threads uint8
	Salt    []byte
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, timeParam, memory, threads, keyLength)

	params := HashParams{
		Memory:  memory,
		Time:    timeParam,
		Threads: threads,
		Salt:    salt,
	}

	encodedHash := encodeHash(hash, params)
	return encodedHash, nil
}

func VerifyPassword(password, hashedPassword string) (bool, error) {
	hash, params, err := decodeHash(hashedPassword)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), params.Salt, params.Time, params.Memory, params.Threads, keyLength)

	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

func encodeHash(hash []byte, params HashParams) string {
	b64Salt := base64.RawStdEncoding.EncodeToString(params.Salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, params.Memory, params.Time, params.Threads, b64Salt, b64Hash)
}

func decodeHash(hashedPassword string) (hash []byte, params HashParams, err error) {
	// Проверяем формат хеша
	if !strings.HasPrefix(hashedPassword, "$argon2id$") {
		return nil, params, fmt.Errorf("invalid hash format")
	}

	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 {
		return nil, params, fmt.Errorf("invalid hash parts count: expected 6, got %d", len(parts))
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, params, fmt.Errorf("failed to parse version: %w", err)
	}

	if version != argon2.Version {
		return nil, params, fmt.Errorf("incompatible argon2 version: %d", version)
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Time, &params.Threads)
	if err != nil {
		return nil, params, fmt.Errorf("failed to parse parameters: %w", err)
	}

	params.Salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, params, fmt.Errorf("failed to decode salt: %w", err)
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, params, fmt.Errorf("failed to decode hash: %w", err)
	}

	return hash, params, nil
}
