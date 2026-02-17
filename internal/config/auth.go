package config

import (
	"fmt"
	"os"
	"strconv"
)

type AuthConfig struct {
	APIKey               string
	AllowedEmail         string
	CookieName           string
	CookieSecret         string
	CookieMaxAge         int // seconds
	SkipCloudflareAuth   bool
	CloudflareTeamDomain string
	CloudflareAudience   string
}

func LoadAuthConfig() (*AuthConfig, error) {
	apiKey := os.Getenv("APP_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("APP_API_KEY is not set")
	}

	allowedEmail := os.Getenv("APP_ALLOWED_EMAIL")
	if allowedEmail == "" {
		return nil, fmt.Errorf("APP_ALLOWED_EMAIL is not set")
	}

	cookieSecret := os.Getenv("APP_COOKIE_SECRET")
	if cookieSecret == "" {
		return nil, fmt.Errorf("APP_COOKIE_SECRET is not set")
	}

	skipCF := os.Getenv("APP_SKIP_CLOUDFLARE_AUTH") == "true"

	cfDomain := os.Getenv("CLOUDFLARE_TEAM_DOMAIN")
	cfAudience := os.Getenv("CLOUDFLARE_AUDIENCE")
	if !skipCF && (cfDomain == "" || cfAudience == "") {
		return nil, fmt.Errorf("CLOUDFLARE_TEAM_DOMAIN and CLOUDFLARE_AUDIENCE are required when Cloudflare auth is enabled")
	}

	cookieMaxAge := 604800 // 7 days
	if v := os.Getenv("APP_COOKIE_MAX_AGE"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("could not convert APP_COOKIE_MAX_AGE to int")
		}
		cookieMaxAge = parsed
	}

	return &AuthConfig{
		APIKey:               apiKey,
		AllowedEmail:         allowedEmail,
		CookieName:           "ephemeral_session",
		CookieSecret:         cookieSecret,
		CookieMaxAge:         cookieMaxAge,
		SkipCloudflareAuth:   skipCF,
		CloudflareTeamDomain: cfDomain,
		CloudflareAudience:   cfAudience,
	}, nil
}
