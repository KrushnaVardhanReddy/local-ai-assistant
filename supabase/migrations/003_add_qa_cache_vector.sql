-- Enable the pgvector extension to work with embedding vectors
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE qa_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    embedding VECTOR(768),
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Create an index for faster similarity search
CREATE INDEX ON qa_cache USING hnsw (embedding vector_cosine_ops);

-- Enable RLS
ALTER TABLE qa_cache ENABLE ROW LEVEL SECURITY;

-- Create Policies
CREATE POLICY "Users can view own cache" ON qa_cache
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own cache" ON qa_cache
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own cache" ON qa_cache
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own cache" ON qa_cache
    FOR DELETE USING (auth.uid() = user_id);
