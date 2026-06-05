// Package secret provides AES-256-GCM encryption for at-rest secrets.
//
// Currently used by: (none). Reserved for encrypting OAuth client_secrets
// when site_settings-backed admin UI provider configuration is implemented.
// See: https://github.com/atbeta/picfast/issues/44
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("secret: key must be 32 bytes for AES-256-GCM, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("secret: aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secret: cipher.NewGCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("secret: generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("secret: key must be 32 bytes for AES-256-GCM, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("secret: aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secret: cipher.NewGCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("secret: ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("secret: gcm.Open: %w", err)
	}
	return plaintext, nil
}

func ParseKey(hexKey string) ([]byte, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("secret: parse key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("secret: key must be 32 bytes for AES-256-GCM, got %d", len(key))
	}
	return key, nil
}
