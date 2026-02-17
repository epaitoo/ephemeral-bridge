package auth

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CloudflareClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type certResponse struct {
	Keys        json.RawMessage `json:"keys"`
	PublicCerts []publicCert    `json:"public_certs"`
}

type publicCert struct {
	Kid  string `json:"kid"`
	Cert string `json:"cert"`
}

type CloudflareVerifier struct {
	teamDomain    string
	audience      string
	mu            sync.RWMutex
	cachedKeys    map[string]crypto.PublicKey
	lastFetch     time.Time
	cacheDuration time.Duration
}

func NewCloudflareVerifier(teamDomain, audience string) *CloudflareVerifier {
	return &CloudflareVerifier{
		teamDomain:    teamDomain,
		audience:      audience,
		cacheDuration: 1 * time.Hour,
	}
}

func (v *CloudflareVerifier) VerifyToken(tokenString string) (*CloudflareClaims, error) {
	keys, err := v.getKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to get public keys: %w", err)
	}

	claims, err := v.parseAndVerify(tokenString, keys)
	if err != nil {
		// Retry with fresh keys in case of key rotation
		freshKeys, fetchErr := v.refreshKeys()
		if fetchErr != nil {
			return nil, err
		}
		claims, err = v.parseAndVerify(tokenString, freshKeys)
		if err != nil {
			return nil, err
		}
	}

	return claims, nil
}

func (v *CloudflareVerifier) parseAndVerify(tokenString string, keys map[string]crypto.PublicKey) (*CloudflareClaims, error) {
	claims := &CloudflareClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		key, exists := keys[kid]
		if !exists {
			return nil, fmt.Errorf("unknown kid: %s", kid)
		}

		return key, nil
	}, jwt.WithValidMethods([]string{"ES256", "RS256"}))

	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Validate issuer
	expectedIssuer := fmt.Sprintf("https://%s", v.teamDomain)
	issuer, err := claims.GetIssuer()
	if err != nil || issuer != expectedIssuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	// Validate audience
	audiences, err := claims.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("invalid audience")
	}
	audienceValid := false
	for _, aud := range audiences {
		if aud == v.audience {
			audienceValid = true
			break
		}
	}
	if !audienceValid {
		return nil, fmt.Errorf("audience mismatch")
	}

	if claims.Email == "" {
		return nil, fmt.Errorf("missing email claim")
	}

	return claims, nil
}

func (v *CloudflareVerifier) getKeys() (map[string]crypto.PublicKey, error) {
	v.mu.RLock()
	if v.cachedKeys != nil && time.Since(v.lastFetch) < v.cacheDuration {
		keys := v.cachedKeys
		v.mu.RUnlock()
		return keys, nil
	}
	v.mu.RUnlock()

	return v.refreshKeys()
}

func (v *CloudflareVerifier) refreshKeys() (map[string]crypto.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Double-check after acquiring write lock
	if v.cachedKeys != nil && time.Since(v.lastFetch) < v.cacheDuration {
		return v.cachedKeys, nil
	}

	keys, err := v.fetchKeys()
	if err != nil {
		return nil, err
	}

	v.cachedKeys = keys
	v.lastFetch = time.Now()
	return keys, nil
}

func (v *CloudflareVerifier) fetchKeys() (map[string]crypto.PublicKey, error) {
	url := fmt.Sprintf("https://%s/cdn-cgi/access/certs", v.teamDomain)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Cloudflare certs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status fetching certs: %d", resp.StatusCode)
	}

	var certResp certResponse
	if err := json.NewDecoder(resp.Body).Decode(&certResp); err != nil {
		return nil, fmt.Errorf("failed to decode cert response: %w", err)
	}

	keys := make(map[string]crypto.PublicKey)
	for _, pc := range certResp.PublicCerts {
		block, _ := pem.Decode([]byte(pc.Cert))
		if block == nil {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		keys[pc.Kid] = cert.PublicKey
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no valid public keys found")
	}

	return keys, nil
}
