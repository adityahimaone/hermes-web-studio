package httpapi

import (
	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestModelCatalogReturnsExplicitUnavailableState(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir()}
	api := NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler()
	response := httptest.NewRecorder()
	api.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/models/catalog", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"status":"unavailable"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
