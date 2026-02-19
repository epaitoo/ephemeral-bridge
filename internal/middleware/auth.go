package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/epaitoo/ephermalbridge/internal/auth"
	"github.com/epaitoo/ephermalbridge/internal/config"
)

type contextKey string

const EmailContextKey contextKey = "auth_email"

// CloudflareMiddleware verifies Cloudflare Access JWT only (Layer 1).
// Used on auth endpoints that need Cloudflare identity but not API key/session.
func CloudflareMiddleware(cfg *config.AuthConfig, verifier *auth.CloudflareVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.SkipCloudflareAuth {
				ctx := context.WithValue(r.Context(), EmailContextKey, cfg.AllowedEmail)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			email, err := verifyCloudflare(r, cfg, verifier)
			if err != nil {
				unauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), EmailContextKey, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthMiddleware verifies Cloudflare Access JWT (Layer 1) AND API key or session cookie (Layer 2).
// Used on all protected routes.
func AuthMiddleware(cfg *config.AuthConfig, verifier *auth.CloudflareVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Layer 1: Cloudflare verification
			if !cfg.SkipCloudflareAuth {
				_, err := verifyCloudflare(r, cfg, verifier)
				if err != nil {
					unauthorized(w)
					return
				}
			}

			// Layer 2: API key or session cookie
			if !validBearerToken(r, cfg) && !validSessionCookie(r, cfg) {
				unauthorized(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func verifyCloudflare(r *http.Request, cfg *config.AuthConfig, verifier *auth.CloudflareVerifier) (string, error) {
	// Check for Service Token headers (bypasses JWT verification)
	clientID := r.Header.Get("CF-Access-Client-Id")
	clientSecret := r.Header.Get("CF-Access-Client-Secret")
	if clientID != "" && clientSecret != "" {
		// Service token present - Cloudflare already validated it at the edge
		return cfg.AllowedEmail, nil
	}

	// Regular JWT verification
	jwtToken := r.Header.Get("CF-Access-JWT-Assertion")
	if jwtToken == "" {
		return "", errUnauthorized
	}

	claims, err := verifier.VerifyToken(jwtToken)
	if err != nil {
		return "", errUnauthorized
	}

	// Service token JWTs have no email claim — signature/issuer/audience already verified above
	if claims.Email == "" {
		return cfg.AllowedEmail, nil
	}

	if !strings.EqualFold(claims.Email, cfg.AllowedEmail) {
		return "", errUnauthorized
	}

	return claims.Email, nil
}

func validBearerToken(r *http.Request, cfg *config.AuthConfig) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return false
	}

	token := parts[1]
	return subtle.ConstantTimeCompare([]byte(token), []byte(cfg.APIKey)) == 1
}

func validSessionCookie(r *http.Request, cfg *config.AuthConfig) bool {
	cookie, err := r.Cookie(cfg.CookieName)
	if err != nil {
		return false
	}

	_, err = auth.ValidateSessionToken(cookie.Value, cfg.CookieSecret)
	return err == nil
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized"}`))
}

var errUnauthorized = errors.New("unauthorized")
