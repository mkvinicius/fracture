package skills

import (
	"encoding/binary"
	"math"
	"strings"
)

// embedDim is the fixed vector dimensionality for all node embeddings.
// Smaller than memory/rag.go's vectorDim (256) — nodes are short labels.
const embedDim = 128

// computeNodeVec produces a normalized TF-IDF embedding for a node's label/text.
// Uses FNV-1a hash bucketing to map tokens to fixed-size buckets — same strategy
// as memory/rag.go so embeddings are comparable across packages.
func computeNodeVec(text string) []float64 {
	tokens := tokenizeText(text)
	vec := make([]float64, embedDim)
	for tok := range tokens {
		idx := int(uint(fnvStr(tok)) % embedDim)
		vec[idx] += 1.0
	}
	return l2Normalize(vec)
}

// cosineSim returns the cosine similarity between two equal-length float64 vectors.
// If both are pre-normalised (as from computeNodeVec) this equals the dot product.
func cosineSim(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// l2Normalize divides every element of vec by its L2 norm (in-place).
// Returns vec unchanged if the norm is zero.
func l2Normalize(vec []float64) []float64 {
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	if norm == 0 {
		return vec
	}
	norm = math.Sqrt(norm)
	for i := range vec {
		vec[i] /= norm
	}
	return vec
}

// tokenizeText lowercases s and splits it into a set of alphanumeric tokens.
func tokenizeText(s string) map[string]struct{} {
	lower := strings.ToLower(s)
	tokens := make(map[string]struct{})
	var word strings.Builder
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			word.WriteRune(r)
		} else if word.Len() > 0 {
			tokens[word.String()] = struct{}{}
			word.Reset()
		}
	}
	if word.Len() > 0 {
		tokens[word.String()] = struct{}{}
	}
	return tokens
}

// serializeVec encodes a float64 slice as little-endian bytes.
// Returns nil for an empty slice.
func serializeVec(vec []float64) []byte {
	if len(vec) == 0 {
		return nil
	}
	buf := make([]byte, len(vec)*8)
	for i, v := range vec {
		binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(v))
	}
	return buf
}

// deserializeVec decodes little-endian bytes back to a float64 slice.
// Returns nil if b is empty or not a multiple of 8 bytes.
func deserializeVec(b []byte) []float64 {
	if len(b) == 0 || len(b)%8 != 0 {
		return nil
	}
	vec := make([]float64, len(b)/8)
	for i := range vec {
		vec[i] = math.Float64frombits(binary.LittleEndian.Uint64(b[i*8:]))
	}
	return vec
}

// fnvStr hashes s with FNV-1a (32-bit) and returns a signed integer.
// Shared by both node embedding (bucket selection) and edge ID generation.
func fnvStr(s string) int {
	const (
		offset32 uint32 = 2166136261
		prime32  uint32 = 16777619
	)
	h := offset32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime32
	}
	return int(h)
}
