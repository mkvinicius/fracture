-- Add simulation mode column to simulation_jobs
ALTER TABLE simulation_jobs ADD COLUMN mode TEXT NOT NULL DEFAULT 'standard';

-- Add sentiment_score column to domain_contexts
ALTER TABLE domain_contexts ADD COLUMN sentiment_score REAL NOT NULL DEFAULT 0.0;
