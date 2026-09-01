package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

func TestSpacesP027RegistryLifecycleAndRemoteState(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), WorkspaceRoot: t.TempDir()}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	create := func(body string) (int, map[string]any) {
		response, err := http.Post(api.URL+"/api/spaces", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		return response.StatusCode, payload
	}
	if status, _ := create(`{"name":"Remote Mac","location_kind":"remote","workspace_ref":"macbook-project","order":2}`); status != http.StatusCreated {
		t.Fatalf("remote create status=%d", status)
	}
	status, local := create(`{"name":"Local","location_kind":"local","workspace_ref":"project","order":1}`)
	if status != http.StatusCreated {
		t.Fatalf("local create status=%d", status)
	}
	if local["id"] == "" || local["health"] != "ready" || local["profile_id"] != "default" {
		t.Fatalf("local=%#v", local)
	}
	response, err := http.Get(api.URL + "/api/spaces")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var listed struct {
		Active string           `json:"active"`
		Spaces []map[string]any `json:"spaces"`
	}
	if err := json.NewDecoder(response.Body).Decode(&listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Spaces) != 2 || listed.Spaces[0]["name"] != "Local" {
		t.Fatalf("listed=%#v", listed)
	}
	if status, _ := create(`{"name":"bad","location_kind":"remote","workspace_ref":""}`); status != http.StatusBadRequest {
		t.Fatalf("invalid remote status=%d", status)
	}
	request, _ := http.NewRequest(http.MethodPost, api.URL+"/api/spaces/active", bytes.NewBufferString(`{"id":"missing"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("activate missing status=%d", response.StatusCode)
	}
	response.Body.Close()
}

func TestSpacesP027ProtectsActiveDeletion(t *testing.T) {
	cfg := config.Config{GatewayBaseURL: "http://127.0.0.1:1", StateDir: t.TempDir(), WorkspaceRoot: t.TempDir()}
	api := httptest.NewServer(NewWithGateway(cfg, gateway.New(gateway.Config{BaseURL: cfg.GatewayBaseURL})).Handler())
	defer api.Close()
	request, _ := http.NewRequest(http.MethodPost, api.URL+"/api/spaces", bytes.NewBufferString(`{"name":"Local","location_kind":"local","workspace_ref":"project"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var item map[string]any
	json.NewDecoder(response.Body).Decode(&item)
	response.Body.Close()
	request, _ = http.NewRequest(http.MethodDelete, api.URL+"/api/spaces/"+item["id"].(string), nil)
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("delete active status=%d", response.StatusCode)
	}
}
