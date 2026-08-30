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
	streamID  string
	sessionID string
	cancel    context.CancelFunc
	events    []replayEvent
	subs      map[chan replayEvent]struct{}
	done      bool
	finished  time.Time
}

type replayEvent struct {
	ID    int64
	Event gateway.Event
}

type startRequest struct {
	SessionID     string   `json:"session_id"`
	Message       string   `json:"message"`
	Model         string   `json:"model"`
	Provider      string   `json:"provider"`
	AttachmentIDs []string `json:"attachment_ids"`
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
	mux.HandleFunc("POST /api/sessions/{session_id}/truncate", s.handleSessionTruncate)
	mux.HandleFunc("DELETE /api/sessions/{session_id}", s.handleSessionDelete)
	mux.HandleFunc("POST /api/chat/start", s.handleChatStart)
	mux.HandleFunc("GET /api/chat/stream", s.handleChatStream)
	mux.HandleFunc("POST /api/chat/cancel", s.handleChatCancel)
	mux.HandleFunc("POST /api/runs/{run_id}/approval", s.handleApproval)
	mux.HandleFunc("POST /api/attachments", s.handleAttachmentUpload)
	mux.HandleFunc("GET /api/attachments/{attachment_id}", s.handleAttachmentDownload)
	return securityHeaders(requestLog(mux))
}

func (s *Server) handleSessions(w http.ResponseWriter, _ *http.Request) {
	sessions, err := s.sessions.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sessions_unavailable", "Session history is unavailable.")
		return
	}
	rows := make([]map[string]any, 0, len(sessions))
	for _, item := range sessions {
		rows = append(rows, summaryPayload(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": rows})
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
	writeJSON(w, http.StatusOK, sessionPayload(item))
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
	writeJSON(w, http.StatusCreated, sessionPayload(item))
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
	writeJSON(w, http.StatusOK, sessionPayload(item))
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
	writeJSON(w, http.StatusOK, sessionPayload(item))
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
	writeJSON(w, http.StatusOK, sessionPayload(item))
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
	writeJSON(w, http.StatusOK, sessionPayload(item))
}

func (s *Server) handleSessionDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.sessions.Delete(r.PathValue("session_id")); err != nil {
		sessionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "session_id": r.PathValue("session_id")})
}

func (s *Server) handleSessionTruncate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Count *int `json:"count"`
	}
	if !decodeBody(w, r, &input) || input.Count == nil {
		if input.Count == nil {
			writeError(w, http.StatusBadRequest, "count_required", "Message count is required.")
		}
		return
	}
	if err := s.sessions.TruncateMessages(r.PathValue("session_id"), *input.Count); err != nil {
		sessionError(w, err)
		return
	}
	item, err := s.sessions.Load(r.PathValue("session_id"))
	if err != nil {
		sessionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sessionPayload(item))
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
	attachments, err := s.sessions.LoadAttachments(input.AttachmentIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_attachment", "One or more attachments are unavailable.")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	item := &turn{streamID: streamID, sessionID: input.SessionID, cancel: cancel, subs: make(map[chan replayEvent]struct{})}
	s.mu.Lock()
	s.turns[streamID] = item
	s.mu.Unlock()

	gatewayAttachments := make([]gateway.Attachment, 0, len(attachments))
	for _, attachment := range attachments {
		gatewayAttachments = append(gatewayAttachments, gateway.Attachment{Name: attachment.Name, MIME: attachment.MIME, Data: attachment.Bytes()})
	}
	go s.runTurn(ctx, item, gateway.ChatRequest{
		SessionID: input.SessionID, Message: input.Message, Model: input.Model, Provider: input.Provider, Attachments: gatewayAttachments,
	})

	writeJSON(w, http.StatusAccepted, map[string]any{"stream_id": streamID, "session_id": input.SessionID})
}

func (s *Server) runTurn(ctx context.Context, item *turn, input gateway.ChatRequest) {
	defer s.finishTurn(item)
	if _, err := s.sessions.Load(input.SessionID); errors.Is(err, os.ErrNotExist) {
		_, _ = s.sessions.Create(input.SessionID, input.Message, nil)
	}
	userMessage := mustMessage("user", input.Message)
	_ = s.sessions.AppendMessages(input.SessionID, userMessage)
	stream := s.gateway.Stream
	if s.config.UseRunsAPI {
		stream = s.gateway.RunStream
	}
	answer, err := stream(ctx, input, func(event gateway.Event) {
		s.publish(item, event)
	})
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			s.publish(item, gateway.Event{Name: "cancel", Data: map[string]any{"message": "Cancelled by user"}})
			return
		}
		code := "gateway_unavailable"
		message := "Hermes Gateway is unavailable. Check that it is running and reachable."
		var httpErr *gateway.HTTPError
		if errors.As(err, &httpErr) {
			code, message = httpErr.Code, httpErr.Message
		}
		s.publish(item, gateway.Event{Name: "apperror", Data: map[string]any{"code": code, "message": message}})
		return
	}
	_ = s.sessions.AppendMessages(input.SessionID, mustMessage("assistant", answer))
	s.publish(item, gateway.Event{Name: "done", Data: map[string]any{"answer": answer, "session_id": item.sessionID}})
}

func (s *Server) publish(item *turn, event gateway.Event) {
	s.mu.Lock()
	record := replayEvent{ID: int64(len(item.events) + 1), Event: event}
	item.events = append(item.events, record)
	for subscriber := range item.subs {
		select {
		case subscriber <- record:
		default:
		}
	}
	s.mu.Unlock()
}

func (s *Server) finishTurn(item *turn) {
	s.mu.Lock()
	item.done = true
	item.finished = time.Now()
	for subscriber := range item.subs {
		close(subscriber)
	}
	item.subs = make(map[chan replayEvent]struct{})
	s.mu.Unlock()
	// Keep completed turns briefly so EventSource can reconnect with Last-Event-ID.
	time.AfterFunc(5*time.Minute, func() {
		s.mu.Lock()
		if current := s.turns[item.streamID]; current == item && item.done {
			delete(s.turns, item.streamID)
		}
		s.mu.Unlock()
	})
}

func (s *Server) handleAttachmentUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(session.MaxAttachmentSize + 1<<20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_attachment", "The attachment upload is invalid.")
		return
	}
	if rawSessionID := strings.TrimSpace(r.FormValue("session_id")); rawSessionID != "" {
		if err := session.ValidateID(rawSessionID); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_session_id", "The session ID is invalid.")
			return
		}
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file_required", "An attachment file is required.")
		return
	}
	_ = file.Close()
	attachment, err := s.sessions.SaveAttachment(header)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_attachment", "The attachment type or size is not allowed.")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": attachment.ID, "name": attachment.Name, "mime": attachment.MIME, "size": attachment.Size})
}

func (s *Server) handleAttachmentDownload(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("attachment_id"))
	attachments, err := s.sessions.LoadAttachments([]string{id})
	if err != nil || len(attachments) != 1 {
		writeError(w, http.StatusNotFound, "attachment_not_found", "The attachment was not found.")
		return
	}
	attachment := attachments[0]
	w.Header().Set("Content-Type", attachment.MIME)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", attachment.Name))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", attachment.Size))
	_, _ = w.Write(attachment.Bytes())
}

func (s *Server) handleApproval(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Choice   string `json:"choice"`
		Decision string `json:"decision"`
	}
	if !decodeBody(w, r, &input) {
		return
	}
	choice := input.Choice
	if choice == "" {
		choice = input.Decision
	}
	if choice == "approved" {
		choice = "once"
	}
	if choice == "denied" {
		choice = "deny"
	}
	if err := s.gateway.ResolveApproval(r.Context(), r.PathValue("run_id"), choice); err != nil {
		writeError(w, http.StatusBadGateway, "approval_failed", "Hermes could not apply the approval decision.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "run_id": r.PathValue("run_id"), "choice": choice})
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
	lastID := int64(0)
	if raw := r.Header.Get("Last-Event-ID"); raw != "" {
		_, _ = fmt.Sscan(raw, &lastID)
	}
	if raw := r.URL.Query().Get("after"); raw != "" {
		_, _ = fmt.Sscan(raw, &lastID)
	}
	subscriber := make(chan replayEvent, 256)
	s.mu.Lock()
	for _, event := range item.events {
		if event.ID > lastID {
			subscriber <- event
		}
	}
	if !item.done {
		item.subs[subscriber] = struct{}{}
	} else {
		close(subscriber)
	}
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(item.subs, subscriber); s.mu.Unlock() }()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case record, open := <-subscriber:
			if !open {
				return
			}
			payload, _ := json.Marshal(record.Event.Data)
			_, _ = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", record.ID, record.Event.Name, payload)
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

func summaryPayload(item session.Summary) map[string]any {
	payload := map[string]any{"session_id": item.ID, "title": item.Title}
	for key, raw := range item.Metadata {
		if key == "session_id" || key == "title" || key == "messages" {
			continue
		}
		var value any
		if json.Unmarshal(raw, &value) == nil {
			payload[key] = value
		}
	}
	return payload
}

func sessionPayload(item session.Session) map[string]any {
	payload := summaryPayload(item.Summary)
	payload["messages"] = item.Messages
	return payload
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
