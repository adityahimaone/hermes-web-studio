package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
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
	contentType := canonicalMIME(http.DetectContentType(data))
	if !allowedMIME(contentType) {
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
	attachment := Attachment{ID: id, Name: filepath.Base(header.Filename), MIME: contentType, Size: int64(len(data)), Path: path}
	metadata, _ := json.Marshal(map[string]string{"name": attachment.Name, "mime": attachment.MIME})
	if err := os.WriteFile(path+".json", metadata, 0600); err != nil {
		_ = os.Remove(path)
		return Attachment{}, err
	}
	return attachment, nil
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
		name, contentType := id, canonicalMIME(http.DetectContentType(data))
		if metadata, readErr := os.ReadFile(path + ".json"); readErr == nil {
			var fields struct {
				Name string `json:"name"`
				MIME string `json:"mime"`
			}
			if json.Unmarshal(metadata, &fields) == nil {
				if fields.Name != "" {
					name = filepath.Base(fields.Name)
				}
				if allowedMIME(fields.MIME) {
					contentType = canonicalMIME(fields.MIME)
				}
			}
		}
		result = append(result, Attachment{ID: id, Name: name, MIME: contentType, Size: int64(len(data)), Path: path, data: data})
	}
	return result, nil
}

func canonicalMIME(value string) string {
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return value
	}
	return mediaType
}

func allowedMIME(mime string) bool {
	return strings.HasPrefix(mime, "image/png") || strings.HasPrefix(mime, "image/jpeg") || strings.HasPrefix(mime, "image/gif") || strings.HasPrefix(mime, "image/webp") || strings.HasPrefix(mime, "application/pdf") || strings.HasPrefix(mime, "text/plain")
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
