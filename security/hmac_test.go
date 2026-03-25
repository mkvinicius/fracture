package security

import (
	"testing"
)

func TestSignAndVerify(t *testing.T) {
	signer, err := NewSigner([]byte("test-secret-key-32-bytes-padded!!"))
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}

	data := "simulation|round-1|agent-abc"
	sig := signer.Sign(data)

	if sig == "" {
		t.Fatal("Sign returned empty signature")
	}
	if !signer.Verify(data, sig) {
		t.Error("Verify returned false for a valid signature")
	}
}

func TestSignDeterministic(t *testing.T) {
	signer, _ := NewSigner([]byte("fixed-secret-key-for-test-32byte"))
	data := "same-input-every-time"
	sig1 := signer.Sign(data)
	sig2 := signer.Sign(data)
	if sig1 != sig2 {
		t.Errorf("Sign is non-deterministic: %q != %q", sig1, sig2)
	}
}

func TestTamperDetection(t *testing.T) {
	signer, err := NewSigner([]byte("test-secret-key-32-bytes-padded!!"))
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}

	data := "original content"
	sig := signer.Sign(data)

	// Tamper with the data
	if signer.Verify("tampered content", sig) {
		t.Error("Verify accepted a tampered payload — HMAC check is broken")
	}

	// Tamper with the signature
	tampered := sig[:len(sig)-4] + "0000"
	if signer.Verify(data, tampered) {
		t.Error("Verify accepted a tampered signature")
	}
}

func TestSignPromptAndVerify(t *testing.T) {
	signer, _ := NewSigner([]byte("test-secret-key-32-bytes-padded!!"))

	sp := signer.SignPrompt("agent-42", 7, "the market will disrupt")
	if !signer.VerifyPrompt(sp) {
		t.Error("VerifyPrompt returned false for a valid SignedPrompt")
	}

	// Tamper with round number
	sp.Round = 99
	if signer.VerifyPrompt(sp) {
		t.Error("VerifyPrompt accepted a tampered Round field")
	}
}

func TestNewSignerWithNilSecret(t *testing.T) {
	// Nil secret should generate a random key — two signers sign differently
	s1, err := NewSigner(nil)
	if err != nil {
		t.Fatalf("NewSigner(nil): %v", err)
	}
	s2, err := NewSigner(nil)
	if err != nil {
		t.Fatalf("NewSigner(nil): %v", err)
	}

	data := "hello"
	if s1.Sign(data) == s2.Sign(data) {
		t.Error("two signers with random keys produced identical signatures — entropy failure")
	}
}
