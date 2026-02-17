package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const currentSessionVersion = 1

type SessionClaims struct {
	Version   int    `json:"v"`
	Email     string `json:"email"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func CreateSessionToken(email string, secret string, maxAge int) (string, error) {
	claims := SessionClaims{
		Version:   currentSessionVersion,
		Email:     email,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Duration(maxAge) * time.Second).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session claims: %w", err)
	}

	sig := computeHMAC(payload, []byte(secret))

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	encodedSig := base64.RawURLEncoding.EncodeToString(sig)

	return encodedPayload + "." + encodedSig, nil
}

func ValidateSessionToken(token string, secret string) (*SessionClaims, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid token encoding")
	}

	receivedSig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding")
	}

	expectedSig := computeHMAC(payload, []byte(secret))
	if !hmac.Equal(receivedSig, expectedSig) {
		return nil, fmt.Errorf("invalid token signature")
	}

	var claims SessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("invalid token payload")
	}

	if claims.Version != currentSessionVersion {
		return nil, fmt.Errorf("unsupported token version")
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

func computeHMAC(data, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
