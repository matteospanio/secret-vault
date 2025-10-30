package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	password := "test-password-123"
	plaintext := []byte("This is a secret message")

	// Encrypt
	ciphertext, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Verify ciphertext is different from plaintext
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("Ciphertext should be different from plaintext")
	}

	// Decrypt
	decrypted, err := Decrypt(ciphertext, password)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	// Verify decrypted matches original
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted text does not match original.\nExpected: %s\nGot: %s", plaintext, decrypted)
	}
}

func TestDecryptWithWrongPassword(t *testing.T) {
	password := "correct-password"
	wrongPassword := "wrong-password"
	plaintext := []byte("Secret data")

	ciphertext, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(ciphertext, wrongPassword)
	if err == nil {
		t.Error("Expected error when decrypting with wrong password")
	}
}

func TestDecryptInvalidData(t *testing.T) {
	password := "test-password"

	// Too short data
	_, err := Decrypt([]byte("short"), password)
	if err == nil {
		t.Error("Expected error for too short ciphertext")
	}

	// Random invalid data
	invalidData := make([]byte, 100)
	for i := range invalidData {
		invalidData[i] = byte(i)
	}

	_, err = Decrypt(invalidData, password)
	if err == nil {
		t.Error("Expected error for invalid ciphertext")
	}
}

func TestEncryptDeterminism(t *testing.T) {
	password := "test-password"
	plaintext := []byte("Test message")

	// Encrypt same data twice
	ciphertext1, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("First encrypt failed: %v", err)
	}

	ciphertext2, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Second encrypt failed: %v", err)
	}

	// Ciphertexts should be different due to random salt and nonce
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Two encryptions of same data should produce different ciphertexts")
	}

	// But both should decrypt to the same plaintext
	decrypted1, _ := Decrypt(ciphertext1, password)
	decrypted2, _ := Decrypt(ciphertext2, password)

	if !bytes.Equal(decrypted1, plaintext) || !bytes.Equal(decrypted2, plaintext) {
		t.Error("Both ciphertexts should decrypt to original plaintext")
	}
}

func TestEmptyData(t *testing.T) {
	password := "test-password"
	plaintext := []byte("")

	ciphertext, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt empty data failed: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, password)
	if err != nil {
		t.Fatalf("Decrypt empty data failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted empty data should match original")
	}
}

func TestLargeData(t *testing.T) {
	password := "test-password"
	plaintext := make([]byte, 1024*1024) // 1 MB
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	ciphertext, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt large data failed: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, password)
	if err != nil {
		t.Fatalf("Decrypt large data failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted large data does not match original")
	}
}
