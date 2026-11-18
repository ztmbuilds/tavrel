package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ztmbuilds/tavrel/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

const (
	sessionCookieName = "tavrel_session"
	sessionTTL        = 7 * 24 * time.Hour // 7 days
	sessionIssuer     = "tavrel"
	sessionSubject    = "operator"
	loginBodyLimit    = 4 << 10 // 4 KiB
	authRoutePrefix   = "/api/v1/auth/"
	authChallenge     = `Bearer realm="tavrel"`
)

type authSessionPayload struct {
	Enabled       bool `json:"enabled"`
	Authenticated bool `json:"authenticated"`
}

type sessionClaims struct {
	jwt.RegisteredClaims
}

type authHandler struct {
	enabled bool
	token   string
}

func newAuthHandler(cfg *config.Config) *authHandler {
	return &authHandler{
		enabled: cfg.AuthToken != "",
		token:   cfg.AuthToken,
	}
}

func (h *authHandler) middleware(next http.Handler) http.Handler {
	if !h.enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, authRoutePrefix) || h.authenticated(r) {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", authChallenge)
		writeError(w, Unauthorized("Authentication required", nil))
	})
}

func (h *authHandler) session(w http.ResponseWriter, r *http.Request) error {
	writeJson(w, http.StatusOK, &Response{
		Data: h.sessionPayload(r),
	})
	return nil
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) error {
	if !h.enabled {
		return BadRequest("Authentication is disabled", nil)
	}

	r.Body = http.MaxBytesReader(w, r.Body, loginBodyLimit)
	defer r.Body.Close()

	var payload struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return PayloadTooLarge("Request body too large", err)
		}
		return BadRequest("Invalid request body", err)
	}

	if !tokensEqual(h.token, payload.Token) {
		w.Header().Set("WWW-Authenticate", authChallenge)
		return Unauthorized("Invalid access token", nil)
	}

	now := time.Now()
	value, err := h.newSessionToken(now)
	if err != nil {
		return Internal(err)
	}

	h.setSessionCookie(w, r, value, now.Add(sessionTTL))

	writeJson(w, http.StatusOK, &Response{
		Data: authSessionPayload{Enabled: true, Authenticated: true},
	})

	return nil
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) error {
	clearSessionCookie(w, r)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *authHandler) sessionPayload(r *http.Request) authSessionPayload {
	if !h.enabled {
		return authSessionPayload{Enabled: false, Authenticated: true}
	}

	return authSessionPayload{Enabled: true, Authenticated: h.authenticated(r)}
}

func (h *authHandler) authenticated(r *http.Request) bool {
	if token, ok := bearerToken(r.Header.Get("Authorization")); ok && tokensEqual(h.token, token) {
		return true
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}

	return h.validSession(cookie.Value)
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	return parts[1], true
}

func tokensEqual(expected, provided string) bool {
	expectedHash := sha256.Sum256([]byte(expected))
	providedHash := sha256.Sum256([]byte(provided))
	return subtle.ConstantTimeCompare(expectedHash[:], providedHash[:]) == 1
}

func (h *authHandler) setSessionCookie(w http.ResponseWriter, r *http.Request, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(sessionTTL / time.Second),
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *authHandler) newSessionToken(now time.Time) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, sessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    sessionIssuer,
			Subject:   sessionSubject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(sessionTTL)),
		},
	}).SignedString([]byte(h.token))
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *authHandler) validSession(value string) bool {
	token, err := jwt.ParseWithClaims(
		value,
		&sessionClaims{},
		func(_ *jwt.Token) (any, error) {
			return []byte(h.token), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(sessionIssuer),
		jwt.WithExpirationRequired(),
	)

	return err == nil && token.Valid
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}

	proto := r.Header.Get("X-Forwarded-Proto")
	if idx := strings.IndexByte(proto, ','); idx >= 0 {
		proto = proto[:idx]
	}

	return strings.EqualFold(strings.TrimSpace(proto), "https")
}
