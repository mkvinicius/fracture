package memory

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/fracture/fracture/engine"
	"github.com/google/uuid"
)

// DocumentType classifies the kind of content stored in a RAG document.
type DocumentType string

const (
	DocSimulationSummary DocumentType = "simulation_summary"
	DocFracturePoint     DocumentType = "fracture_point"
	DocCompanyContext    DocumentType = "company_context"
	DocDomainSignal      DocumentType = "domain_signal"
)

// vectorDim is the fixed dimensionality of all TF-IDF vectors.
// Hash-bucketing maps any vocabulary to this fixed size.
const vectorDim = 256

// RAGDocument is a single indexed unit in the local RAG store.
type RAGDocument struct {
	ID        string
	CompanyID string
	Type      DocumentType
	Content   string
	Metadata  string // JSON
	TFIDF     []byte // serialized float64[vectorDim] little-endian
	CreatedAt time.Time
}

// DomainSignal carries domain research data for RAG indexing.
// Defined here (not in deepsearch) to avoid a circular import.
type DomainSignal struct {
	Domain    string
	Summary   string
	Signals   []string
	Sentiment float64
}

// RAGStore is a SQLite-backed store for local TF-IDF retrieval.
type RAGStore struct {
	db *sql.DB
}

// NewRAGStore returns a RAGStore backed by the given SQLite connection.
func NewRAGStore(db *sql.DB) *RAGStore {
	return &RAGStore{db: db}
}

// ─── public API ──────────────────────────────────────────────────────────────

// Index stores a document, computing its TF-IDF vector before insertion.
func (r *RAGStore) Index(companyID string, doc RAGDocument) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	if doc.CompanyID == "" {
		doc.CompanyID = companyID
	}
	if doc.Metadata == "" {
		doc.Metadata = "{}"
	}

	vec := computeTFIDF(tokenize(doc.Content))
	blob := serializeTFIDF(vec)

	_, err := r.db.Exec(`
		INSERT INTO rag_documents (id, company_id, doc_type, content, metadata, tfidf, created_at)
		VALUES (?, ?, ?, ?, ?, ?, unixepoch())
		ON CONFLICT(id) DO UPDATE SET
			content    = excluded.content,
			metadata   = excluded.metadata,
			tfidf      = excluded.tfidf,
			created_at = excluded.created_at
	`, doc.ID, doc.CompanyID, string(doc.Type), doc.Content, doc.Metadata, blob)
	return err
}

// Search returns up to n documents most semantically similar to query,
// scoped to the given company, ranked by cosine similarity of TF-IDF vectors.
func (r *RAGStore) Search(companyID, query string, n int) ([]RAGDocument, error) {
	queryVec := computeTFIDF(tokenize(query))

	rows, err := r.db.Query(`
		SELECT id, company_id, doc_type, content, metadata, tfidf, created_at
		FROM rag_documents
		WHERE company_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, companyID, n*10) // over-fetch then rank
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scored struct {
		doc   RAGDocument
		score float64
	}
	var candidates []scored

	for rows.Next() {
		var (
			doc      RAGDocument
			tfidfBlob []byte
			createdAt int64
		)
		if err := rows.Scan(
			&doc.ID, &doc.CompanyID, &doc.Type, &doc.Content,
			&doc.Metadata, &tfidfBlob, &createdAt,
		); err != nil {
			continue
		}
		doc.CreatedAt = time.Unix(createdAt, 0)
		doc.TFIDF = tfidfBlob

		docVec := deserializeTFIDF(tfidfBlob)
		score := cosineSimilarity(queryVec, docVec)
		candidates = append(candidates, scored{doc, score})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rag search: %w", err)
	}

	// Sort by descending similarity
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	result := make([]RAGDocument, 0, n)
	for i, c := range candidates {
		if i >= n {
			break
		}
		result = append(result, c.doc)
	}
	return result, nil
}

// IndexSimulation indexes the key artifacts of a completed simulation so that
// future simulations for the same company can retrieve them via Search.
func (r *RAGStore) IndexSimulation(
	companyID string,
	report engine.FullReport,
	domainSignals []DomainSignal,
) error {
	// 1. Simulation summary document
	summaryMeta, _ := json.Marshal(map[string]interface{}{
		"simulation_id": report.SimulationID,
		"question":      report.Question,
		"tokens":        report.TotalTokens,
		"duration_ms":   report.DurationMs,
	})
	if err := r.Index(companyID, RAGDocument{
		ID:       "sim-summary-" + report.SimulationID,
		Type:     DocSimulationSummary,
		Content:  report.Question + "\n" + report.ProbableFuture.Narrative,
		Metadata: string(summaryMeta),
	}); err != nil {
		return fmt.Errorf("index simulation summary: %w", err)
	}

	// 2. Fracture point documents (one per event)
	for i, fe := range report.FractureEvents {
		content := fmt.Sprintf(
			"Fracture proposed by %s in round %d: %s → %s",
			fe.ProposedBy, fe.Round,
			fe.Proposal.OriginalRuleID,
			fe.Proposal.NewDescription,
		)
		meta, _ := json.Marshal(map[string]interface{}{
			"simulation_id": report.SimulationID,
			"round":         fe.Round,
			"accepted":      fe.Accepted,
			"confidence":    fe.Confidence,
			"proposed_by":   fe.ProposedBy,
		})
		if err := r.Index(companyID, RAGDocument{
			ID:       fmt.Sprintf("fracture-%s-%d", report.SimulationID, i),
			Type:     DocFracturePoint,
			Content:  content,
			Metadata: string(meta),
		}); err != nil {
			// Non-fatal: continue indexing remaining events
			continue
		}
	}

	// 3. Rupture scenarios as company context
	for i, rs := range report.RuptureScenarios {
		content := fmt.Sprintf(
			"Ruptura em %s (prob %.0f%%): %s. Quem quebra: %s. Como: %s",
			rs.RuleDescription, rs.Probability*100,
			rs.ImpactOnCompany, rs.WhoBreaks, rs.HowItHappens,
		)
		meta, _ := json.Marshal(map[string]interface{}{
			"simulation_id": report.SimulationID,
			"rule_id":       rs.RuleID,
			"probability":   rs.Probability,
		})
		_ = r.Index(companyID, RAGDocument{
			ID:       fmt.Sprintf("rupture-%s-%d", report.SimulationID, i),
			Type:     DocCompanyContext,
			Content:  content,
			Metadata: string(meta),
		})
	}

	// 4. Domain signal documents
	for _, ds := range domainSignals {
		if ds.Summary == "" {
			continue
		}
		content := ds.Domain + ": " + ds.Summary
		if len(ds.Signals) > 0 {
			content += "\nSinais: " + joinStrings(ds.Signals, "; ")
		}
		meta, _ := json.Marshal(map[string]interface{}{
			"simulation_id": report.SimulationID,
			"domain":        ds.Domain,
			"sentiment":     ds.Sentiment,
		})
		_ = r.Index(companyID, RAGDocument{
			ID:       fmt.Sprintf("domain-%s-%s", report.SimulationID, ds.Domain),
			Type:     DocDomainSignal,
			Content:  content,
			Metadata: string(meta),
		})
	}

	return nil
}

// ─── TF-IDF implementation ────────────────────────────────────────────────────

// computeTFIDF builds a normalized TF vector of fixed size vectorDim.
// Each token is mapped to a bucket via FNV-style hash bucketing.
// The resulting vector is L2-normalized so that dot product == cosine similarity.
func computeTFIDF(tokens map[string]struct{}) []float64 {
	vec := make([]float64, vectorDim)
	for tok := range tokens {
		bucket := fnvBucket(tok, vectorDim)
		vec[bucket] += 1.0
	}
	// L2 normalize
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec
}

// cosineSimilarity computes the cosine similarity between two float64 vectors.
// If vectors are pre-normalized (as from computeTFIDF), this equals the dot product.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// serializeTFIDF converts a float64 slice to a []byte using little-endian encoding.
func serializeTFIDF(vec []float64) []byte {
	buf := make([]byte, len(vec)*8)
	for i, v := range vec {
		bits := math.Float64bits(v)
		binary.LittleEndian.PutUint64(buf[i*8:], bits)
	}
	return buf
}

// deserializeTFIDF converts a little-endian []byte back to a float64 slice.
func deserializeTFIDF(b []byte) []float64 {
	if len(b)%8 != 0 {
		return make([]float64, vectorDim)
	}
	vec := make([]float64, len(b)/8)
	for i := range vec {
		bits := binary.LittleEndian.Uint64(b[i*8:])
		vec[i] = math.Float64frombits(bits)
	}
	return vec
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// fnvBucket maps a string token to a bucket index in [0, size) using FNV-1a.
func fnvBucket(s string, size int) int {
	const (
		offset32 uint32 = 2166136261
		prime32  uint32 = 16777619
	)
	h := offset32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime32
	}
	return int(h) % size
}

// joinStrings joins a string slice with a separator.
func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
