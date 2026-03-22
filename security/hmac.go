package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Signer signs and verifies simulation prompts using HMAC-SHA256.
type Signer struct {
	secret []byte
}

// NewSigner creates a Signer. If secret is nil, a random key is generated.
func NewSigner(secret []byte) (*Signer, error) {
	if secret == nil {
		secret = make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, fmt.Errorf("generate secret: %w", err)
		}
	}
	return &Signer{secret: secret}, nil
}

// Sign returns the HMAC-SHA256 hex signature of the given data.
func (s *Signer) Sign(data string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks that the signature matches the data.
func (s *Signer) Verify(data, signature string) bool {
	expected := s.Sign(data)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// SignedPrompt wraps a prompt with its HMAC signature.
type SignedPrompt struct {
	AgentID   string `json:"agent_id"`
	Round     int    `json:"round"`
	Content   string `json:"content"`
	Signature string `json:"signature"`
}

// SignPrompt creates a SignedPrompt for the given agent and content.
func (s *Signer) SignPrompt(agentID string, round int, content string) SignedPrompt {
	data := fmt.Sprintf("%s|%d|%s", agentID, round, content)
	return SignedPrompt{
		AgentID:   agentID,
		Round:     round,
		Content:   content,
		Signature: s.Sign(data),
	}
}

// VerifyPrompt checks that a SignedPrompt has not been tampered with.
func (s *Signer) VerifyPrompt(sp SignedPrompt) bool {
	data := fmt.Sprintf("%s|%d|%s", sp.AgentID, sp.Round, sp.Content)
	return s.Verify(data, sp.Signature)
}

// AuditLogger writes immutable, chained audit entries to SQLite.
type AuditLogger struct {
	db      *sql.DB
	signer  *Signer
	prevSig string // chain: each entry signs the previous signature
}

// NewAuditLogger creates an AuditLogger backed by the given DB.
func NewAuditLogger(db *sql.DB, signer *Signer) *AuditLogger {
	return &AuditLogger{db: db, signer: signer}
}

// Log writes an immutable audit entry.
func (a *AuditLogger) Log(eventType, entityID string, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		payloadJSON = []byte("{}")
	}

	// Chain: sign (eventType + entityID + payload + prevSig)
	data := fmt.Sprintf("%s|%s|%s|%s|%d",
		eventType, entityID, string(payloadJSON), a.prevSig, time.Now().UnixNano())
	sig := a.signer.Sign(data)
	a.prevSig = sig

	_, err = a.db.Exec(`
			INSERT INTO audit_log (event_type, entity_id, payload, hmac_sig, created_at)
			VALUES (?, ?, ?, ?, unixepoch())
		`, eventType, entityID, string(payloadJSON), sig)
	return err
}
