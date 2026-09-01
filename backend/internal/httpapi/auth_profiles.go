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
	http.SetCookie(w, sessionCookieFor(r, token, 86400))
	writeJSON(w, 200, map[string]any{"authenticated": true})
}
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(w, r) {
		return
	}
	http.SetCookie(w, sessionCookieFor(r, "", -1))
	writeJSON(w, 200, map[string]any{"authenticated": false})
}

func sessionCookieFor(r *http.Request, value string, maxAge int) *http.Cookie {
	secure := r.TLS != nil || strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
	return &http.Cookie{Name: sessionCookie, Value: value, Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: maxAge}
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
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	writeJSON(w, 200, map[string]any{"profiles": s.profiles, "active": s.activeProfile})
}

func (s *Server) handleProfileCreate(w http.ResponseWriter, r *http.Request) {
	var input profile
	if !decodeBody(w, r, &input) || strings.TrimSpace(input.ID) == "" || strings.TrimSpace(input.Name) == "" {
		writeError(w, http.StatusBadRequest, "profile_invalid", "Profile id and name are required.")
		return
	}
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	for _, item := range s.profiles {
		if item.ID == input.ID {
			writeError(w, http.StatusConflict, "profile_exists", "The profile already exists.")
			return
		}
	}
	input.Health = "configured"
	s.profiles = append(s.profiles, input)
	writeJSON(w, http.StatusCreated, input)
}

func (s *Server) handleProfileUpdate(w http.ResponseWriter, r *http.Request) {
	var patch profile
	if !decodeBody(w, r, &patch) {
		return
	}
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	providerAvailable := true
	if patch.ProviderID != "" {
		providerAvailable = false
		for _, item := range s.providers {
			if item.ID == patch.ProviderID {
				providerAvailable = true
				break
			}
		}
	}
	for i := range s.profiles {
		if s.profiles[i].ID == patch.ID {
			if patch.ProviderID != "" && s.profiles[i].ID == s.activeProfile && !providerAvailable {
				writeError(w, http.StatusConflict, "profile_provider_unavailable", "The active profile cannot reference an unavailable provider.")
				return
			}
			if patch.Name != "" {
				s.profiles[i].Name = patch.Name
			}
			if patch.Model != "" {
				s.profiles[i].Model = patch.Model
			}
			if patch.Provider != "" {
				s.profiles[i].Provider = patch.Provider
			}
			if patch.ProviderID != "" {
				s.profiles[i].ProviderID = patch.ProviderID
			}
			writeJSON(w, http.StatusOK, s.profiles[i])
			return
		}
	}
	writeError(w, http.StatusNotFound, "profile_not_found", "The profile was not found.")
}

func (s *Server) handleProfileDelete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if id == "" || id == s.activeProfile || len(s.profiles) == 1 {
		writeError(w, http.StatusBadRequest, "profile_delete_rejected", "The active or last profile cannot be deleted.")
		return
	}
	for i, item := range s.profiles {
		if item.ID == id {
			s.profiles = append(s.profiles[:i], s.profiles[i+1:]...)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
			return
		}
	}
	writeError(w, http.StatusNotFound, "profile_not_found", "The profile was not found.")
}

func (s *Server) handleProviders(w http.ResponseWriter, _ *http.Request) {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{"providers": s.providers})
}

func (s *Server) handleModelCatalog(w http.ResponseWriter, r *http.Request) {
	models, err := s.gateway.Models(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": "unavailable", "models": []any{}, "message": "Hermes Gateway model catalog is unavailable."})
		return
	}
	rows := make([]map[string]any, 0, len(models))
	for _, model := range models {
		rows = append(rows, map[string]any{"id": model.ID, "name": model.ID, "provider": model.Provider, "aliases": model.Aliases, "available": true})
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "models": rows})
}

func (s *Server) handleProviderCreate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Kind    string `json:"kind"`
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
	}
	if !decodeBody(w, r, &input) || strings.TrimSpace(input.ID) == "" || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.BaseURL) == "" {
		writeError(w, http.StatusBadRequest, "provider_invalid", "Provider id, name, and base URL are required.")
		return
	}
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	for _, item := range s.providers {
		if item.ID == input.ID {
			writeError(w, http.StatusConflict, "provider_exists", "The provider already exists.")
			return
		}
	}
	item := provider{ID: input.ID, Name: input.Name, Kind: input.Kind, BaseURL: input.BaseURL, HasKey: strings.TrimSpace(input.APIKey) != "", Health: "configured"}
	s.providers = append(s.providers, item)
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) handleProviderDelete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" || id == "gateway" {
		writeError(w, http.StatusBadRequest, "provider_delete_rejected", "The gateway provider cannot be deleted.")
		return
	}
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	for _, profile := range s.profiles {
		if profile.ID == s.activeProfile && (profile.ProviderID == id || profile.Provider == id) {
			writeError(w, http.StatusConflict, "provider_in_use", "The active profile is using this provider.")
			return
		}
	}
	for i, item := range s.providers {
		if item.ID == id {
			s.providers = append(s.providers[:i], s.providers[i+1:]...)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
			return
		}
	}
	writeError(w, http.StatusNotFound, "provider_not_found", "The provider was not found.")
}

func (s *Server) handleSettingsCapabilities(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"sections": []string{"conversation", "appearance", "preferences", "providers", "plugins", "extensions", "system", "help"}, "locales": []string{"en", "id", "de", "es", "fr", "it", "ja", "ko", "pt-BR", "ru", "zh-CN", "zh-TW", "ar", "hi", "tr"}})
}
func (s *Server) handleProfileSwitch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID string `json:"id"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	for _, p := range s.profiles {
		if p.ID == input.ID {
			providerAvailable := p.ProviderID == ""
			for _, provider := range s.providers {
				if provider.ID == p.ProviderID {
					providerAvailable = true
					break
				}
			}
			if !providerAvailable {
				writeError(w, http.StatusConflict, "profile_provider_unavailable", "The profile references an unavailable provider.")
				return
			}
			s.activeProfile = p.ID
			writeJSON(w, 200, map[string]any{"active": p})
			return
		}
	}
	writeError(w, 404, "profile_not_found", "The profile was not found.")
}

func (s *Server) providerExists(id string) bool {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	for _, item := range s.providers {
		if item.ID == id {
			return true
		}
	}
	return false
}
func (s *Server) providerIDs() map[string]bool {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	ids := make(map[string]bool, len(s.providers))
	for _, item := range s.providers {
		ids[item.ID] = true
	}
	return ids
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
