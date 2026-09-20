package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type ArgonParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultArgonParams = &ArgonParams{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// Menghasilkan hash kata sandi yang aman menggunakan algoritma Argon2id.
func HashPassword(password string) (string, error) {
	salt := make([]byte, DefaultArgonParams.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("gagal menghasilkan salt acak: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		DefaultArgonParams.Iterations,
		DefaultArgonParams.Memory,
		DefaultArgonParams.Parallelism,
		DefaultArgonParams.KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		DefaultArgonParams.Memory,
		DefaultArgonParams.Iterations,
		DefaultArgonParams.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

// Memverifikasi kecocokan antara kata sandi teks polos dengan format hash Argon2id.
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("format hash terenkripsi tidak valid")
	}

	if parts[1] != "argon2id" {
		return false, errors.New("varian argon2 tidak didukung")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("gagal mem-parsing versi argon2: %w", err)
	}

	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("gagal mem-parsing parameter argon2: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("gagal mendekode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("gagal mendekode hash: %w", err)
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(expectedHash)),
	)

	if subtle.ConstantTimeCompare(actualHash, expectedHash) == 1 {
		return true, nil
	}

	return false, nil
}
