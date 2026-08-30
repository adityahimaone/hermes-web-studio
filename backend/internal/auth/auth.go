package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const iterations = 120000

var ErrRateLimited = errors.New("rate limited")

type persisted struct {
	PasswordHash string `json:"password_hash"`
	Secret       string `json:"secret"`
}

type Service struct {
	path     string
	mu       sync.RWMutex
	data     persisted
	attempts map[string][]time.Time
}

func New(stateDir string) (*Service, error) {
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return nil, err
	}
	s := &Service{path: filepath.Join(stateDir, "auth.json"), attempts: make(map[string][]time.Time)}
	data, err := os.ReadFile(s.path)
	if err == nil {
		if json.Unmarshal(data, &s.data) != nil {
			return nil, errors.New("auth state is invalid")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if s.data.Secret == "" {
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return nil, err
		}
		s.data.Secret = hex.EncodeToString(raw)
		if err := s.persist(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Service) Enabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.PasswordHash != ""
}
func (s *Service) Setup(password string) error {
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.PasswordHash != "" {
		return errors.New("password already configured")
	}
	s.data.PasswordHash = hash(password, s.data.Secret)
	return s.persist()
}
func (s *Service) Login(ip, password string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	recent := s.attempts[ip][:0]
	for _, at := range s.attempts[ip] {
		if now.Sub(at) < time.Minute {
			recent = append(recent, at)
		}
	}
	s.attempts[ip] = recent
	if len(recent) >= 5 {
		return "", ErrRateLimited
	}
	s.attempts[ip] = append(s.attempts[ip], now)
	if s.data.PasswordHash == "" || !hmac.Equal([]byte(s.data.PasswordHash), []byte(hash(password, s.data.Secret))) {
		return "", errors.New("invalid credentials")
	}
	return s.sign(now.Add(24 * time.Hour)), nil
}
func (s *Service) Verify(cookie string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	parts := strings.Split(cookie, ".")
	if len(parts) != 2 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expiry, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil || time.Now().Unix() > expiry {
		return false
	}
	expected := s.signature(parts[0])
	return hmac.Equal([]byte(parts[1]), []byte(expected))
}
func (s *Service) sign(expiry time.Time) string {
	payload := base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(expiry.Unix(), 10)))
	return payload + "." + s.signature(payload)
}
func (s *Service) signature(value string) string {
	mac := hmac.New(sha256.New, []byte(s.data.Secret))
	_, _ = mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func (s *Service) persist() error {
	data, _ := json.Marshal(s.data)
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
func hash(password, secret string) string {
	sum := sha256.Sum256([]byte(secret + "\x00" + password))
	for i := 0; i < iterations; i++ {
		next := sha256.Sum256(append(sum[:], []byte(password)...))
		sum = next
	}
	return fmt.Sprintf("sha256$%s", hex.EncodeToString(sum[:]))
}
