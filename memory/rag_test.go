package memory

import (
	"database/sql"
	"math"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func openRAGTestDB(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "rag-test-*.db")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.Close()

	db, err := sql.Open("sqlite3", f.Name()+"?_foreign_keys=off")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS rag_documents (
			id          TEXT PRIMARY KEY,
			company_id  TEXT NOT NULL,
			doc_type    TEXT NOT NULL,
			content     TEXT NOT NULL,
			metadata    TEXT NOT NULL DEFAULT '{}',
			tfidf       BLOB,
			created_at  INTEGER NOT NULL DEFAULT (unixepoch())
		);
		CREATE INDEX IF NOT EXISTS idx_rag_company ON rag_documents(company_id, doc_type);
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func TestRAGIndexAndSearch(t *testing.T) {
	db := openRAGTestDB(t)
	r := NewRAGStore(db)

	const company = "acme-corp"

	docs := []RAGDocument{
		{ID: "doc-1", Type: DocSimulationSummary, Content: "market disruption competitive pricing pressure"},
		{ID: "doc-2", Type: DocFracturePoint, Content: "technology regulation artificial intelligence adoption"},
		{ID: "doc-3", Type: DocCompanyContext, Content: "supply chain geopolitical risk currency fluctuation"},
	}
	for _, d := range docs {
		if err := r.Index(company, d); err != nil {
			t.Fatalf("Index %s: %v", d.ID, err)
		}
	}

	// Search with a query similar to doc-1
	results, err := r.Search(company, "market pricing competitive disruption", 2)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one search result")
	}
	// The top result should be doc-1 (highest term overlap)
	if results[0].ID != "doc-1" {
		t.Errorf("expected top result to be doc-1, got %s", results[0].ID)
	}
}

func TestRAGSearchRespectsCompanyScope(t *testing.T) {
	db := openRAGTestDB(t)
	r := NewRAGStore(db)

	r.Index("company-a", RAGDocument{ID: "a-doc", Type: DocSimulationSummary, Content: "finance investment capital"})
	r.Index("company-b", RAGDocument{ID: "b-doc", Type: DocSimulationSummary, Content: "finance investment capital"})

	results, err := r.Search("company-a", "finance investment", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for _, res := range results {
		if res.CompanyID != "company-a" {
			t.Errorf("got document from wrong company: %s", res.CompanyID)
		}
	}
}

func TestRAGIndexUpsert(t *testing.T) {
	db := openRAGTestDB(t)
	r := NewRAGStore(db)

	doc := RAGDocument{ID: "upsert-doc", Type: DocFracturePoint, Content: "original content"}
	if err := r.Index("co", doc); err != nil {
		t.Fatalf("Index: %v", err)
	}
	doc.Content = "updated content with more words"
	if err := r.Index("co", doc); err != nil {
		t.Fatalf("Index upsert: %v", err)
	}

	results, err := r.Search("co", "updated content", 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 || results[0].Content != "updated content with more words" {
		t.Errorf("expected upserted content, got %v", results)
	}
}

func TestComputeTFIDFProducesNormalizedVector(t *testing.T) {
	tokens := tokenize("market disruption competitive pricing pressure technology")
	vec := computeTFIDF(tokens)

	if len(vec) != vectorDim {
		t.Errorf("expected vector dim %d, got %d", vectorDim, len(vec))
	}

	// Compute L2 norm — should be ~1.0 after normalization
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if math.Abs(norm-1.0) > 0.001 {
		t.Errorf("expected L2 norm≈1.0, got %.6f", norm)
	}
}

func TestCosineSimilaritySameVector(t *testing.T) {
	tokens := tokenize("fracture disruption market")
	vec := computeTFIDF(tokens)
	sim := cosineSimilarity(vec, vec)
	if math.Abs(sim-1.0) > 0.001 {
		t.Errorf("cosine similarity of vector with itself should be 1.0, got %.6f", sim)
	}
}

func TestSerializeDeserializeRoundtrip(t *testing.T) {
	tokens := tokenize("regulation finance geopolitics technology")
	original := computeTFIDF(tokens)
	blob := serializeTFIDF(original)
	recovered := deserializeTFIDF(blob)

	if len(recovered) != len(original) {
		t.Fatalf("length mismatch: original %d, recovered %d", len(original), len(recovered))
	}
	for i, v := range original {
		if math.Abs(v-recovered[i]) > 1e-12 {
			t.Errorf("index %d: original %.10f != recovered %.10f", i, v, recovered[i])
		}
	}
}

func TestRAGSearchEmptyStore(t *testing.T) {
	db := openRAGTestDB(t)
	r := NewRAGStore(db)

	results, err := r.Search("nobody", "any query", 5)
	if err != nil {
		t.Fatalf("Search on empty store: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results from empty store, got %d", len(results))
	}
}

func TestRAGDocumentCreatedAtPopulated(t *testing.T) {
	db := openRAGTestDB(t)
	r := NewRAGStore(db)

	before := time.Now().Add(-time.Second)
	r.Index("co", RAGDocument{ID: "ts-doc", Type: DocDomainSignal, Content: "timestamp test"})

	results, _ := r.Search("co", "timestamp test", 1)
	if len(results) == 0 {
		t.Fatal("expected 1 result")
	}
	if results[0].CreatedAt.Before(before) {
		t.Errorf("CreatedAt %v is before test start %v", results[0].CreatedAt, before)
	}
}
