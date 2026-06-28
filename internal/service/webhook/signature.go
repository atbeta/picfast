package webhook

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/atbeta/picfast/internal/secret"
)

type SecretHash = string

func GenerateSecret(encKey []byte) (plain string, hash SecretHash, ciphertext []byte, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", nil, err
	}
	plain = "whsec_" + hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(sum[:])

	ciphertext, err = secret.Encrypt([]byte(plain), encKey)
	if err != nil {
		return "", "", nil, err
	}
	return plain, hash, ciphertext, nil
}

func DecryptSecret(ciphertext []byte, encKey []byte) (string, error) {
	plain, err := secret.Decrypt(ciphertext, encKey)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func ComputeSignature(secret string, timestamp string, body []byte) string {
	payload := timestamp + "." + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func NormalizeEvents(events json.RawMessage) ([]string, error) {
	if len(events) == 0 || string(events) == "[]" || string(events) == "null" {
		return nil, nil
	}
	var evts []string
	if err := json.Unmarshal(events, &evts); err != nil {
		return nil, err
	}
	return evts, nil
}

func EventMatches(subscribed []string, eventType string) bool {
	if len(subscribed) == 0 {
		return false
	}
	for _, s := range subscribed {
		if s == eventType {
			return true
		}
	}
	return false
}
