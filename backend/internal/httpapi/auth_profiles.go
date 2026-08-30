package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/auth"
)

const sessionCookie = "hermes_web_session"

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/onboarding" || r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/me" || r.URL.Path == "/health" || r.URL.Path == "/api/health/hermes" {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/api/onboarding/password" {
			if !sameOrigin(w, r) {
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		if s.config.AuthTrustedHeader != "" && strings.TrimSpace(r.Header.Get(s.config.AuthTrustedHeader)) != "" {
			if sameOrigin(w, r) {
				next.ServeHTTP(w, r)
			}
			return
		}
		if s.authErr != nil {
			writeError(w, http.StatusServiceUnavailable, "auth_unavailable", "Authentication is unavailable.")
			return
		}
		if !loopbackHost(s.config.Host) && s.auth != nil && !s.auth.Enabled() {
			writeError(w, http.StatusServiceUnavailable, "authentication_setup_required", "Configure authentication before remote access.")
			return
		}
		if s.auth != nil && s.auth.Enabled() {
			cookie, err := r.Cookie(sessionCookie)
			if err != nil || !s.auth.Verify(cookie.Value) {
				writeError(w, http.StatusUnauthorized, "authentication_required", "Sign in to continue.")
				return
			}
			if r.Method != http.MethodGet && r.Method != http.MethodHead && !sameOrigin(w, r) {
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func sameOrigin(w http.ResponseWriter, r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" || u.Host != r.Host || (u.Scheme != "http" && u.Scheme != "https") {
		writeError(w, http.StatusForbidden, "origin_rejected", "Request origin is not allowed.")
		return false
	}
	return true
}

func (s *Server) handleOnboarding(w http.ResponseWriter, _ *http.Request) {
	if s.authErr != nil {
		writeError(w, 503, "auth_unavailable", "Authentication is unavailable.")
		return
	}
	writeJSON(w, 200, map[string]any{"configured": s.auth.Enabled(), "password_required": !s.auth.Enabled(), "providers": map[string]bool{"password": true, "trusted_header": s.config.AuthTrustedHeader != "", "oidc": s.config.OIDCIssuer != "", "passkey": false}})
}
func (s *Server) handlePasswordSetup(w http.ResponseWriter, r *http.Request) {
	if s.authErr != nil {
		writeError(w, 503, "auth_unavailable", "Authentication is unavailable.")
		return
	}
	var input struct {
		Password string `json:"password"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	if err := s.auth.Setup(input.Password); err != nil {
		writeError(w, 400, "password_setup_failed", err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"configured": true})
}
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if s.authErr != nil {
		writeError(w, 503, "auth_unavailable", "Authentication is unavailable.")
		return
	}
	var input struct {
		Password string `json:"password"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	token, err := s.auth.Login(clientIP(r), input.Password)
	if err != nil {
		if errors.Is(err, auth.ErrRateLimited) {
			writeError(w, 429, "rate_limited", "Too many login attempts.")
		} else {
			writeError(w, 401, "invalid_credentials", "The password is incorrect.")
		}
		return
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode, MaxAge: 86400})
	writeJSON(w, 200, map[string]any{"authenticated": true})
}
func (s *Server) handleLogout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	writeJSON(w, 200, map[string]any{"authenticated": false})
}
func (s *Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	authenticated := false
	if s.config.AuthTrustedHeader != "" && r.Header.Get(s.config.AuthTrustedHeader) != "" {
		authenticated = true
	} else if s.auth != nil && s.auth.Enabled() {
		if c, err := r.Cookie(sessionCookie); err == nil {
			authenticated = s.auth.Verify(c.Value)
		}
	}
	writeJSON(w, 200, map[string]any{"authenticated": authenticated, "user": "local"})
}
func (s *Server) handleProfiles(w http.ResponseWriter, _ *http.Request) {
	s.profileMu.RLock()
	defer s.profileMu.RUnlock()
	writeJSON(w, 200, map[string]any{"profiles": s.profiles, "active": s.activeProfile})
}
func (s *Server) handleProfileSwitch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID string `json:"id"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	s.profileMu.Lock()
	defer s.profileMu.Unlock()
	for _, p := range s.profiles {
		if p.ID == input.ID {
			s.activeProfile = p.ID
			writeJSON(w, 200, map[string]any{"active": p})
			return
		}
	}
	writeError(w, 404, "profile_not_found", "The profile was not found.")
}
func (s *Server) handleAuthProviders(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"password": true, "trusted_header": s.config.AuthTrustedHeader != "", "oidc": s.config.OIDCIssuer != "", "passkey": false, "oidc_issuer": s.config.OIDCIssuer})
}
func clientIP(r *http.Request) string {
	host, _, err := strings.Cut(r.RemoteAddr, ":")
	if err {
		return r.RemoteAddr
	}
	return host
}

func loopbackHost(host string) bool {
	return host == "" || host == "127.0.0.1" || host == "::1" || host == "localhost"
}
