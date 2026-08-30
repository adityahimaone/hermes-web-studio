package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/session"
)

type Server struct {
	config   config.Config
	gateway  *gateway.Client
	sessions *session.Store
	turns    map[string]*turn
	mu       sync.Mutex
}

type turn struct {
	sessionID string
	events    chan gateway.Event
	cancel    context.CancelFunc
}

type startRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
	Model     string `json:"model"`
	Provider  string `json:"provider"`
}

func New(cfg config.Config) *Server {
	return NewWithGateway(cfg, gateway.New(gateway.Config{
		BaseURL:     cfg.GatewayBaseURL,
		APIKey:      cfg.GatewayAPIKey,
		ReadTimeout: cfg.ReadTimeout,
	}))
}

func NewWithGateway(cfg config.Config, client *gateway.Client) *Server {
	stateDir := cfg.StateDir
	if stateDir == "" {
		stateDir, _ = session.ResolveStateDir(os.Getenv, os.UserHomeDir)
	}
	return &Server{config: cfg, gateway: client, sessions: session.NewStore(stateDir), turns: make(map[string]*turn)}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/health/hermes", s.handleHermesHealth)
	mux.HandleFunc("GET /api/sessions", s.handleSessions)
	mux.HandleFunc("GET /api/sessions/{session_id}", s.handleSession)
	mux.HandleFunc("POST /api/sessions", s.handleSessionCreate)
	mux.HandleFunc("PATCH /api/sessions/{session_id}", s.handleSessionUpdate)
	mux.HandleFunc("POST /api/sessions/{session_id}/rename", s.handleSessionRename)
	mux.HandleFunc("POST /api/sessions/{session_id}/pin", s.handleSessionPin)
	mux.HandleFunc("POST /api/sessions/{session_id}/archive", s.handleSessionArchive)
	mux.HandleFunc("DELETE /api/sessions/{session_id}", s.handleSessionDelete)
	mux.HandleFunc("POST /api/chat/start", s.handleChatStart)
	mux.HandleFunc("GET /api/chat/stream", s.handleChatStream)
	mux.HandleFunc("POST /api/chat/cancel", s.handleChatCancel)
	return securityHeaders(requestLog(mux))
}

func (s *Server) handleSessions(w http.ResponseWriter, _ *http.Request) {
	sessions, err := s.sessions.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sessions_unavailable", "Session history is unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("session_id")
	item, err := s.sessions.Load(id)
	if err != nil {
		if errors.Is(err, session.ErrInvalidSessionID) {
			writeError(w, http.StatusBadRequest, "invalid_session_id", "The session ID is invalid.")
			return
		}
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "session_not_found", "The requested session was not found.")
			return
		}
		writeError(w, http.StatusInternalServerError, "session_unavailable", "The requested session could not be loaded.")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleSessionCreate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SessionID string `json:"session_id"`
		Title     string `json:"title"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	if input.SessionID == "" {
		input.SessionID = newID()
	}
	item, err := s.sessions.Create(input.SessionID, strings.TrimSpace(input.Title), nil)
	if err != nil {
		sessionError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) handleSessionUpdate(w http.ResponseWriter, r *http.Request) {
	patch, ok := decodeRawObject(w, r)
	if !ok {
		return
	}
	item, err := s.sessions.Update(r.PathValue("session_id"), patch)
	if err != nil {
		sessionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleSessionRename(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	item, err := s.sessions.Rename(r.PathValue("session_id"), strings.TrimSpace(input.Title))
	if err != nil {
		sessionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleSessionPin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Pinned *bool `json:"pinned"`
	}
	if !decodeBody(w, r, &input) || input.Pinned == nil {
		if input.Pinned == nil {
			writeError(w, http.StatusBadRequest, "pinned_required", "Pinned is required.")
		}
		return
	}
	item, err := s.sessions.SetPinned(r.PathValue("session_id"), *input.Pinned)
	if err != nil {
		sessionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleSessionArchive(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Archived *bool `json:"archived"`
	}
	if !decodeBody(w, r, &input) || input.Archived == nil {
		if input.Archived == nil {
			writeError(w, http.StatusBadRequest, "archived_required", "Archived is required.")
		}
		return
	}
	item, err := s.sessions.SetArchived(r.PathValue("session_id"), *input.Archived)
	if err != nil {
		sessionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleSessionDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.sessions.Delete(r.PathValue("session_id")); err != nil {
		sessionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "session_id": r.PathValue("session_id")})
}

func sessionError(w http.ResponseWriter, err error) {
	if errors.Is(err, session.ErrInvalidSessionID) {
		writeError(w, http.StatusBadRequest, "invalid_session_id", "The session ID is invalid.")
		return
	}
	if errors.Is(err, session.ErrSessionExists) {
		writeError(w, http.StatusConflict, "session_exists", "The session already exists.")
		return
	}
	if errors.Is(err, os.ErrNotExist) {
		writeError(w, http.StatusNotFound, "session_not_found", "The requested session was not found.")
		return
	}
	if strings.Contains(err.Error(), "required") {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "session_unavailable", "The session could not be changed.")
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "hermes-web-studio"})
}

func (s *Server) handleHermesHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	err := s.gateway.Health(ctx)
	payload := map[string]any{
		"ok": err == nil, "configured": s.gateway.Configured(), "reachable": err == nil,
		"base_url": publicGatewayURL(s.gateway.BaseURL()),
	}
	if err != nil {
		payload["message"] = "Hermes Gateway is not reachable. Start Hermes or check HERMES_WEBUI_GATEWAY_BASE_URL."
	}
	writeJSON(w, http.StatusOK, payload)
}

func publicGatewayURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "configured"
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func (s *Server) handleChatStart(w http.ResponseWriter, r *http.Request) {
	var input startRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "The chat request is invalid.")
		return
	}
	input.Message = strings.TrimSpace(input.Message)
	if input.Message == "" {
		writeError(w, http.StatusBadRequest, "message_required", "A message is required.")
		return
	}
	if input.SessionID == "" {
		input.SessionID = newID()
	}
	if err := session.ValidateID(input.SessionID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_session_id", "The session ID is invalid.")
		return
	}
	if input.Model == "" {
		input.Model = s.config.DefaultModel
	}
	if input.Provider == "" {
		input.Provider = s.config.DefaultProvider
	}

	streamID := newID()
	ctx, cancel := context.WithCancel(context.Background())
	item := &turn{sessionID: input.SessionID, events: make(chan gateway.Event, 256), cancel: cancel}
	s.mu.Lock()
	s.turns[streamID] = item
	s.mu.Unlock()

	go s.runTurn(ctx, item, gateway.ChatRequest{
		SessionID: input.SessionID, Message: input.Message, Model: input.Model, Provider: input.Provider,
	})

	writeJSON(w, http.StatusAccepted, map[string]any{"stream_id": streamID, "session_id": input.SessionID})
}

func (s *Server) runTurn(ctx context.Context, item *turn, input gateway.ChatRequest) {
	defer close(item.events)
	if _, err := s.sessions.Load(input.SessionID); errors.Is(err, os.ErrNotExist) {
		_, _ = s.sessions.Create(input.SessionID, input.Message, nil)
	}
	userMessage := mustMessage("user", input.Message)
	_ = s.sessions.AppendMessages(input.SessionID, userMessage)
	answer, err := s.gateway.Stream(ctx, input, func(event gateway.Event) {
		select {
		case item.events <- event:
		case <-ctx.Done():
		}
	})
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			item.events <- gateway.Event{Name: "cancel", Data: map[string]any{"message": "Cancelled by user"}}
			return
		}
		code := "gateway_unavailable"
		message := "Hermes Gateway is unavailable. Check that it is running and reachable."
		var httpErr *gateway.HTTPError
		if errors.As(err, &httpErr) {
			code, message = httpErr.Code, httpErr.Message
		}
		item.events <- gateway.Event{Name: "apperror", Data: map[string]any{"code": code, "message": message}}
		return
	}
	_ = s.sessions.AppendMessages(input.SessionID, mustMessage("assistant", answer))
	item.events <- gateway.Event{Name: "done", Data: map[string]any{"answer": answer, "session_id": item.sessionID}}
}

func mustMessage(role, content string) json.RawMessage {
	data, _ := json.Marshal(map[string]string{"role": role, "content": content})
	return data
}

func decodeBody(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "The request body is invalid.")
		return false
	}
	return true
}

func decodeRawObject(w http.ResponseWriter, r *http.Request) (map[string]json.RawMessage, bool) {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	var value map[string]json.RawMessage
	if err := decoder.Decode(&value); err != nil || value == nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "The request body is invalid.")
		return nil, false
	}
	return value, true
}

func (s *Server) handleChatStream(w http.ResponseWriter, r *http.Request) {
	streamID := strings.TrimSpace(r.URL.Query().Get("stream_id"))
	s.mu.Lock()
	item := s.turns[streamID]
	s.mu.Unlock()
	if item == nil {
		writeError(w, http.StatusNotFound, "stream_not_found", "The requested stream does not exist or has expired.")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "stream_unsupported", "Streaming is unavailable on this server.")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	defer func() {
		s.mu.Lock()
		delete(s.turns, streamID)
		s.mu.Unlock()
	}()
	for {
		select {
		case event, open := <-item.events:
			if !open {
				return
			}
			payload, _ := json.Marshal(event.Data)
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Name, payload)
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleChatCancel(w http.ResponseWriter, r *http.Request) {
	streamID := strings.TrimSpace(r.URL.Query().Get("stream_id"))
	s.mu.Lock()
	item := s.turns[streamID]
	s.mu.Unlock()
	if item == nil {
		writeError(w, http.StatusNotFound, "stream_not_found", "The requested stream does not exist or has expired.")
		return
	}
	item.cancel()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func newID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buffer)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"code": code, "message": message})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
	})
}
