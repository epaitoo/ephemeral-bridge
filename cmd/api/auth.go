package main

import (
	"net/http"
	"net/url"

	"github.com/epaitoo/ephermalbridge/internal/auth"
	"github.com/epaitoo/ephermalbridge/internal/middleware"
)

func (app *application) createSessionHandler(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value(middleware.EmailContextKey).(string)
	if !ok || email == "" {
		app.unauthorizedResponse(w, r)
		return
	}

	token, err := auth.CreateSessionToken(email, app.authConfig.CookieSecret, app.authConfig.CookieMaxAge)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     app.authConfig.CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   app.authConfig.CookieMaxAge,
		HttpOnly: true,
		Secure:   !app.authConfig.SkipCloudflareAuth,
		SameSite: http.SameSiteLaxMode,
	})

	redirectParam := r.URL.Query().Get("redirect")
	if redirectParam != "" {
		parsed, err := url.Parse(redirectParam)
		if err == nil && (parsed.Host == "ephemeralbridge.cc" || parsed.Host == "www.ephemeralbridge.cc") {
			http.Redirect(w, r, redirectParam, http.StatusFound)
			return
		}
	}

	app.writeJSON(w, http.StatusOK, envelope{"message": "session created"}, nil)
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     app.authConfig.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   !app.authConfig.SkipCloudflareAuth,
		SameSite: http.SameSiteLaxMode,
	})

	app.writeJSON(w, http.StatusOK, envelope{"message": "logged out"}, nil)
}
