package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltSize   = 32
	nonceSize  = 12
	iterations = 100000
	keySize    = 32 // AES-256
)

// Encrypt encrypts data using AES-256-GCM with a password
func Encrypt(data []byte, password string) ([]byte, error) {
	// Generate a random salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, iterations, keySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Combine salt + nonce + ciphertext
	result := make([]byte, 0, saltSize+nonceSize+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt decrypts data using AES-256-GCM with a password
func Decrypt(data []byte, password string) ([]byte, error) {
	if len(data) < saltSize+nonceSize {
		return nil, fmt.Errorf("invalid ciphertext: too short")
	}

	// Extract salt, nonce, and ciphertext
	salt := data[:saltSize]
	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, iterations, keySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: incorrect password or corrupted data")
	}

	return plaintext, nil
}
