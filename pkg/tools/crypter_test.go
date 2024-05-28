package tools

// Encrypts data using AES-128-CBC with valid key and IV
import (
	"bytes"
	"testing"
)

func TestAES128EncryptValidKeyAndIV(t *testing.T) {
	// Arrange
	origData := []byte("data to encrypt")
	key := []byte("1234567890123456")
	iv := []byte("abcdefghijklmnop")

	// Act
	encryptedData, err := AES128Encrypt(origData, key, iv)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if len(encryptedData) == 0 {
		t.Error("Expected encrypted data, but got empty slice")
	}
}

func TestAES128EncryptEmptyKey(t *testing.T) {
	// Arrange
	origData := []byte("data to encrypt")
	key := []byte{}
	iv := []byte("abcdefghijklmnop")

	// Act
	encryptedData, err := AES128Encrypt(origData, key, iv)

	// Assert
	if err == nil {
		t.Error("Expected error, but got nil")
	}

	if len(encryptedData) > 0 {
		t.Error("Expected empty encrypted data, but got non-empty slice")
	}
}

func TestAES128DecryptValidKeyAndIV(t *testing.T) {
	origData := []byte("data to encrypt")
	key := []byte("1234567890123456")
	iv := []byte("abcdefghijklmnop")

	encryptedData, err := AES128Encrypt(origData, key, iv)

	decryptedData, err := AES128Decrypt(encryptedData, key, iv)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	expectedData := []byte("data to encrypt")
	if !bytes.Equal(decryptedData, expectedData) {
		t.Errorf("Expected decrypted data to be %v but got %v", expectedData, decryptedData)
	}
}
