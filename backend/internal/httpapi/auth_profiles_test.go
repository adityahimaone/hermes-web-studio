package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestPasswordAuthProtectsAPIAndSetsHttpOnlyCookie(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	handler := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	request := httptest.NewRequest(http.MethodPost, "/api/onboarding/password", strings.NewReader(`{"password":"correct horse battery"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("setup=%d %s", recorder.Code, recorder.Body.String())
	}
	request = httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("protected=%d", recorder.Code)
	}
	request = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"password":"correct horse battery"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Header().Get("Set-Cookie"), "HttpOnly") {
		t.Fatalf("login=%d cookie=%q", recorder.Code, recorder.Header().Get("Set-Cookie"))
	}
	var payload map[string]any
	_ = json.Unmarshal(recorder.Body.Bytes(), &payload)
}

func TestOnboardingSetupCanRecoverAfterValidationFailure(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	handler := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()

	request := httptest.NewRequest(http.MethodPost, "/api/onboarding/password", strings.NewReader(`{"password":"too-short"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://attacker.example")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), "origin_rejected") {
		t.Fatalf("cross-origin setup=%d %s", recorder.Code, recorder.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/api/onboarding/password", strings.NewReader(`{"password":"too-short"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "password_setup_failed") {
		t.Fatalf("invalid setup=%d %s", recorder.Code, recorder.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/api/onboarding/password", strings.NewReader(`{"password":"correct horse battery"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("recovered setup=%d %s", recorder.Code, recorder.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/api/onboarding", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"configured":true`) {
		t.Fatalf("onboarding after recovery=%d %s", recorder.Code, recorder.Body.String())
	}
}

func TestLoginRotatesCookieAndLogoutRejectsCrossOrigin(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	handler := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	setup := httptest.NewRequest(http.MethodPost, "/api/onboarding/password", strings.NewReader(`{"password":"correct horse battery"}`))
	setup.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), setup)

	login := func() *http.Cookie {
		request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"password":"correct horse battery"}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Forwarded-Proto", "https")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("login=%d %s", recorder.Code, recorder.Body.String())
		}
		parsed := &http.Cookie{Name: sessionCookie}
		parsed.Value = strings.TrimPrefix(strings.Split(recorder.Header().Get("Set-Cookie"), ";")[0], sessionCookie+"=")
		if !strings.Contains(recorder.Header().Get("Set-Cookie"), "Secure") || !strings.Contains(recorder.Header().Get("Set-Cookie"), "HttpOnly") {
			t.Fatalf("proxy cookie flags=%q", recorder.Header().Get("Set-Cookie"))
		}
		return parsed
	}
	first, second := login(), login()
	if first.Value == second.Value {
		t.Fatal("login did not rotate the cookie")
	}

	logout := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logout.AddCookie(second)
	logout.Header.Set("Origin", "http://attacker.example")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, logout)
	if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), "origin_rejected") {
		t.Fatalf("cross-origin logout=%d %s", recorder.Code, recorder.Body.String())
	}
}

func TestProfileAndProviderCRUDRedactsCredentials(t *testing.T) {
	api := newTestServer(t, "http://127.0.0.1:1", "")
	defer api.Close()
	profile := postJSON(t, api.URL+"/api/profiles", map[string]any{"id": "writing", "name": "Writing", "model": "q4", "provider_id": "custom"})
	if profile.StatusCode != http.StatusCreated {
		t.Fatalf("profile create=%d", profile.StatusCode)
	}
	_ = profile.Body.Close()
	provider := postJSON(t, api.URL+"/api/providers", map[string]any{"id": "custom", "name": "Local", "kind": "openai-compatible", "base_url": "http://127.0.0.1:9000/v1", "api_key": "do-not-return"})
	if provider.StatusCode != http.StatusCreated {
		t.Fatalf("provider create=%d", provider.StatusCode)
	}
	body, _ := io.ReadAll(provider.Body)
	_ = provider.Body.Close()
	if strings.Contains(string(body), "do-not-return") || !strings.Contains(string(body), `"has_key":true`) {
		t.Fatalf("provider leaked credential: %s", body)
	}
	providers, err := api.Client().Get(api.URL + "/api/providers")
	if err != nil {
		t.Fatal(err)
	}
	providerList := readBody(providers)
	if strings.Contains(providerList, "do-not-return") || strings.Contains(providerList, "api_key") {
		t.Fatalf("provider list exposed credential fields: %s", providerList)
	}

	settingsRequest, err := http.NewRequest(http.MethodPut, api.URL+"/api/preferences", strings.NewReader(`{"theme":"dark","api_key":"should-not-persist","password":"should-not-persist"}`))
	if err != nil {
		t.Fatal(err)
	}
	settingsRequest.Header.Set("Content-Type", "application/json")
	settingsResponse, err := api.Client().Do(settingsRequest)
	if err != nil {
		t.Fatal(err)
	}
	settingsBody := readBody(settingsResponse)
	if settingsResponse.StatusCode != http.StatusOK || strings.Contains(settingsBody, "should-not-persist") || !strings.Contains(settingsBody, `"theme":"dark"`) {
		t.Fatalf("unsafe settings response=%d body=%s", settingsResponse.StatusCode, settingsBody)
	}
	response, err := api.Client().Get(api.URL + "/api/settings/capabilities")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("capabilities=%d", response.StatusCode)
	}
}

func TestActiveProfileCannotBeLeftWithoutItsProvider(t *testing.T) {
	api := newTestServer(t, "http://127.0.0.1:1", "")
	defer api.Close()

	provider := postJSON(t, api.URL+"/api/providers", map[string]any{
		"id": "local", "name": "Local", "kind": "openai-compatible",
		"base_url": "http://127.0.0.1:9000/v1", "api_key": "server-only-key",
	})
	if provider.StatusCode != http.StatusCreated {
		t.Fatalf("provider create=%d %s", provider.StatusCode, readBody(provider))
	}
	_ = provider.Body.Close()

	profile := postJSON(t, api.URL+"/api/profiles", map[string]any{
		"id": "local-profile", "name": "Local", "model": "q4", "provider_id": "local",
	})
	if profile.StatusCode != http.StatusCreated {
		t.Fatalf("profile create=%d %s", profile.StatusCode, readBody(profile))
	}
	_ = profile.Body.Close()

	switchResponse := postJSON(t, api.URL+"/api/profiles/active", map[string]any{"id": "local-profile"})
	if switchResponse.StatusCode != http.StatusOK {
		t.Fatalf("profile switch=%d %s", switchResponse.StatusCode, readBody(switchResponse))
	}
	_ = switchResponse.Body.Close()

	deleteRequest, err := http.NewRequest(http.MethodDelete, api.URL+"/api/providers?id=local", nil)
	if err != nil {
		t.Fatal(err)
	}
	deleteResponse, err := api.Client().Do(deleteRequest)
	if err != nil {
		t.Fatal(err)
	}
	if deleteResponse.StatusCode != http.StatusConflict || !strings.Contains(readBody(deleteResponse), "provider_in_use") {
		t.Fatalf("active provider delete=%d", deleteResponse.StatusCode)
	}

	update := postJSON(t, api.URL+"/api/profiles/active", map[string]any{"id": "missing"})
	if update.StatusCode != http.StatusNotFound {
		t.Fatalf("missing profile switch=%d", update.StatusCode)
	}
	_ = update.Body.Close()

	gatewayDelete, err := api.Client().Do(mustRequest(t, http.MethodDelete, api.URL+"/api/providers?id=gateway"))
	if err != nil {
		t.Fatal(err)
	}
	if gatewayDelete.StatusCode != http.StatusBadRequest || !strings.Contains(readBody(gatewayDelete), "provider_delete_rejected") {
		t.Fatalf("gateway delete=%d", gatewayDelete.StatusCode)
	}
}

func TestProfileSwitchRejectsUnavailableProviderWithoutChangingActiveProfile(t *testing.T) {
	cfg := config.Config{
		GatewayBaseURL: "http://127.0.0.1:1",
		DefaultModel:   "default",
		ProfilesJSON:   `[{"id":"safe","name":"Safe","model":"default"},{"id":"broken","name":"Broken","model":"q4","provider_id":"missing"}]`,
		StateDir:       t.TempDir(),
	}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()

	response := postJSON(t, api.URL+"/api/profiles/active", map[string]any{"id": "broken"})
	if response.StatusCode != http.StatusConflict || !strings.Contains(readBody(response), "profile_provider_unavailable") {
		t.Fatalf("broken profile switch=%d", response.StatusCode)
	}

	profiles, err := api.Client().Get(api.URL + "/api/profiles")
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Active string `json:"active"`
	}
	decode(t, profiles.Body, &payload)
	_ = profiles.Body.Close()
	if payload.Active != "safe" {
		t.Fatalf("active profile changed after rejected switch: %q", payload.Active)
	}
}

func mustRequest(t *testing.T, method, url string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func readBody(response *http.Response) string {
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	return string(body)
}
