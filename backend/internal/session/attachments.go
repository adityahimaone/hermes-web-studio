package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const MaxAttachmentSize int64 = 10 << 20

type Attachment struct {
	ID, Name, MIME string
	Size           int64
	Path           string
	data           []byte
}

// Bytes exposes the already validated attachment payload to the gateway adapter.
// HTTP handlers must use the metadata fields when returning an attachment.
func (a Attachment) Bytes() []byte { return append([]byte(nil), a.data...) }

func (s *Store) SaveAttachment(header *multipart.FileHeader) (Attachment, error) {
	if header == nil || header.Size <= 0 || header.Size > MaxAttachmentSize {
		return Attachment{}, errors.New("attachment size is not allowed")
	}
	file, err := header.Open()
	if err != nil {
		return Attachment{}, err
	}
	defer file.Close()
	limited := io.LimitReader(file, MaxAttachmentSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return Attachment{}, err
	}
	if int64(len(data)) > MaxAttachmentSize {
		return Attachment{}, errors.New("attachment size is not allowed")
	}
	mime := http.DetectContentType(data)
	if !allowedMIME(mime) {
		return Attachment{}, errors.New("attachment type is not allowed")
	}
	id := randomID()
	dir := filepath.Join(s.stateDir, "attachments")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return Attachment{}, err
	}
	path := filepath.Join(dir, id)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return Attachment{}, err
	}
	return Attachment{ID: id, Name: filepath.Base(header.Filename), MIME: mime, Size: int64(len(data)), Path: path}, nil
}

func (s *Store) LoadAttachments(ids []string) ([]Attachment, error) {
	result := make([]Attachment, 0, len(ids))
	for _, id := range ids {
		if id == "" || filepath.Base(id) != id || strings.ContainsAny(id, `/\\`) {
			return nil, ErrInvalidSessionID
		}
		path := filepath.Join(s.stateDir, "attachments", id)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > MaxAttachmentSize {
			return nil, errors.New("attachment size is not allowed")
		}
		result = append(result, Attachment{ID: id, Name: id, MIME: http.DetectContentType(data), Size: int64(len(data)), Path: path, data: data})
	}
	return result, nil
}

func allowedMIME(mime string) bool {
	switch mime {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "application/pdf", "text/plain":
		return true
	default:
		return false
	}
}
func randomID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "attachment-fallback"
	}
	return hex.EncodeToString(b)
}

// data is intentionally private so handlers cannot accidentally return file bytes.
// The gateway adapter receives it through the explicit conversion below.
