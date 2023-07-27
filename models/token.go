package models

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/dhinogz/my-tunes/rand"
)

type TokenManager struct {
	// BytesPerToken is used to determine how many bytes to use when generating
	// each session token. If this value is not set or is less than the
	// MinBytesPerToken const it will be ignored and MinBytesPerToken will be used
	BytesPerToken int
}

func (tm TokenManager) New() (token, tokenHash string, err error) {
	bytesPerToken := tm.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err = rand.String(bytesPerToken)
	if err != nil {
		return "", "", fmt.Errorf("new token: %w", err)
	}
	tokenHash = tm.hash(token)
	return token, tokenHash, nil
}

func (tm *TokenManager) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
