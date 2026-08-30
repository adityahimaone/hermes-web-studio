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
	"strings"
	"sync"
	"time"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/config"
	"github.com/adityahimaone/hermes-web-studio/backend/internal/gateway"
)

type Server struct {
	config  config.Config
	gateway *gateway.Client
	turns   map[string]*turn
	mu      sync.Mutex
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
	return &Server{config: cfg, gateway: client, turns: make(map[string]*turn)}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/health/hermes", s.handleHermesHealth)
	mux.HandleFunc("POST /api/chat/start", s.handleChatStart)
	mux.HandleFunc("GET /api/chat/stream", s.handleChatStream)
	mux.HandleFunc("POST /api/chat/cancel", s.handleChatCancel)
	return securityHeaders(requestLog(mux))
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
		"base_url": s.gateway.BaseURL(),
	}
	if err != nil {
		payload["message"] = "Hermes Gateway is not reachable. Start Hermes or check HERMES_WEBUI_GATEWAY_BASE_URL."
	}
	writeJSON(w, http.StatusOK, payload)
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
	item.events <- gateway.Event{Name: "done", Data: map[string]any{"answer": answer, "session_id": item.sessionID}}
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
