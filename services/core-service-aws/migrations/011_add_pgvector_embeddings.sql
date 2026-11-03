-- Migration 011: Install pgvector and create AI embeddings table
-- Enables semantic search across all CRM data using Google Gemini embeddings

-- Step 1: Enable pgvector extension (must be installed on PostgreSQL server)
-- On AWS RDS: enable via Parameter Group or run as superuser
-- On self-hosted: run: sudo apt install postgresql-16-pgvector (adjust version)
CREATE EXTENSION IF NOT EXISTS vector;

-- Step 2: AI embeddings table — stores vectorized representations of CRM records
CREATE TABLE IF NOT EXISTS ai_embeddings (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_type TEXT NOT NULL,          -- 'user', 'deal', 'contact', 'audit_log', 'note', 'lead'
    source_id   TEXT NOT NULL,          -- UUID of the original record
    content     TEXT NOT NULL,          -- Human-readable text that was embedded (for display)
    embedding   vector(768),            -- Gemini text-embedding-004 produces 768-dim vectors
    metadata    JSONB DEFAULT '{}',     -- Extra filterable data (status, date, etc.)
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Step 3: Unique constraint — one embedding per source record
CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_embeddings_source
    ON ai_embeddings(source_type, source_id);

-- Step 4: IVFFlat index for fast approximate nearest-neighbour search
-- lists=100 is good for up to ~1M rows; cosine distance works best for Gemini embeddings
CREATE INDEX IF NOT EXISTS idx_ai_embeddings_vector
    ON ai_embeddings USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Step 5: Metadata index for filtered searches
CREATE INDEX IF NOT EXISTS idx_ai_embeddings_source_type ON ai_embeddings(source_type);
CREATE INDEX IF NOT EXISTS idx_ai_embeddings_created    ON ai_embeddings(created_at DESC);
