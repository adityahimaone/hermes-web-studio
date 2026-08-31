package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/control"
)

func validateMCPServer(server control.MCPServer) error {
	if strings.TrimSpace(server.Name) == "" || len(server.Name) > 100 {
		return errors.New("name_required")
	}
	switch server.Transport {
	case "stdio":
		if strings.TrimSpace(server.Command) == "" || len(server.Command) > 240 {
			return errors.New("command_required")
		}
	case "http", "sse":
		parsed, err := url.Parse(server.Endpoint)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil {
			return errors.New("endpoint_invalid")
		}
	default:
		return errors.New("transport_invalid")
	}
	if len(server.Args) > 32 || len(server.Tools) > 256 {
		return errors.New("limits_exceeded")
	}
	for _, tool := range server.Tools {
		if strings.TrimSpace(tool.Name) == "" || len(tool.Name) > 120 {
			return errors.New("tool_invalid")
		}
	}
	return nil
}

func (s *Server) handleMCPServers(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]any{"servers": store.MCPServers(), "available": true})
		return
	}
	var input control.MCPServer
	if !decodeBody(w, r, &input) {
		return
	}
	if err := validateMCPServer(input); err != nil {
		writeError(w, http.StatusBadRequest, "mcp_server_invalid", "The MCP server configuration is invalid.")
		return
	}
	if r.Method == http.MethodPost {
		created, err := store.CreateMCPServer(input)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "mcp_server_save_failed", "The MCP server could not be saved.")
			return
		}
		writeJSON(w, http.StatusCreated, created)
		return
	}
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "The requested method is not supported.")
		return
	}
	servers := store.MCPServers()
	var existing *control.MCPServer
	for i := range servers {
		if servers[i].ID == r.PathValue("id") {
			existing = &servers[i]
			break
		}
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "mcp_server_not_found", "The MCP server was not found.")
		return
	}
	merged := *existing
	if input.Name != "" {
		merged.Name = input.Name
	}
	if input.Transport != "" {
		merged.Transport = input.Transport
	}
	if input.Endpoint != "" {
		merged.Endpoint = input.Endpoint
	}
	if input.Command != "" {
		merged.Command = input.Command
	}
	if input.Args != nil {
		merged.Args = input.Args
	}
	if input.Tools != nil {
		merged.Tools = input.Tools
	}
	if input.Settings != nil {
		merged.Settings = input.Settings
	}
	merged.Enabled = input.Enabled
	if err := validateMCPServer(merged); err != nil {
		writeError(w, http.StatusBadRequest, "mcp_server_invalid", "The MCP server configuration is invalid.")
		return
	}
	updated, err := store.UpdateMCPServer(merged.ID, merged)
	if errors.Is(err, control.ErrNotFound) {
		writeError(w, http.StatusNotFound, "mcp_server_not_found", "The MCP server was not found.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "mcp_server_save_failed", "The MCP server could not be saved.")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleMCPServerDelete(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	if err := store.DeleteMCPServer(r.PathValue("id")); errors.Is(err, control.ErrNotFound) {
		writeError(w, http.StatusNotFound, "mcp_server_not_found", "The MCP server was not found.")
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "mcp_server_delete_failed", "The MCP server could not be deleted.")
	} else {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func (s *Server) handleMCPServerTools(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	for _, server := range store.MCPServers() {
		if server.ID == r.PathValue("id") {
			tools := server.Tools
			if tools == nil {
				tools = []control.MCPTool{}
			}
			writeJSON(w, http.StatusOK, map[string]any{"server_id": server.ID, "status": server.Status, "tools": tools, "read_only": true})
			return
		}
	}
	writeError(w, http.StatusNotFound, "mcp_server_not_found", "The MCP server was not found.")
}

func (s *Server) handlePluginUpdate(w http.ResponseWriter, r *http.Request) {
	store, ok := s.controlReady(w)
	if !ok {
		return
	}
	var input control.Plugin
	if !decodeBody(w, r, &input) {
		return
	}
	if len(input.Settings) > 64 {
		writeError(w, http.StatusBadRequest, "plugin_settings_invalid", "Plugin settings exceed the supported limit.")
		return
	}
	id := r.PathValue("id")
	if !validRegistrationName(id) {
		writeError(w, http.StatusBadRequest, "plugin_invalid", "The plugin id is invalid.")
		return
	}
	updated, err := store.UpdatePlugin(id, input)
	if errors.Is(err, control.ErrNotFound) {
		for _, item := range discoverFiles(filepath.Join(s.config.StateDir, "plugins"), "") {
			if item["path"] == id {
				input.ID, input.Name, input.Path, input.Status = id, item["name"], id, "available"
				if upsertErr := store.UpsertPlugin(input); upsertErr != nil {
					writeError(w, http.StatusInternalServerError, "plugin_settings_failed", "Plugin settings could not be saved.")
					return
				}
				updated, err = store.UpdatePlugin(id, input)
				break
			}
		}
	}
	if errors.Is(err, control.ErrNotFound) {
		writeError(w, http.StatusNotFound, "plugin_not_found", "The plugin was not found.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "plugin_settings_failed", "Plugin settings could not be saved.")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
