-- Migration 008: RAG local document store
-- Adds rag_documents table for TF-IDF based local semantic search.
-- Used by memory.RAGStore to index simulation summaries, fracture points,
-- company context, and domain signals for retrieval in future simulations.

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
