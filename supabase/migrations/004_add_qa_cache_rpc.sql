-- Create a function to perform semantic similarity search on the qa_cache table
CREATE OR REPLACE FUNCTION match_qa_cache (
  query_embedding VECTOR(768),
  match_threshold FLOAT,
  match_count INT,
  p_user_id UUID
)
RETURNS TABLE (
  id UUID,
  question TEXT,
  answer TEXT,
  similarity FLOAT
)
LANGUAGE plpgsql
AS $$
BEGIN
  RETURN QUERY
  SELECT
    qa_cache.id,
    qa_cache.question,
    qa_cache.answer,
    1 - (qa_cache.embedding <=> query_embedding) AS similarity
  FROM qa_cache
  WHERE qa_cache.user_id = p_user_id
    AND 1 - (qa_cache.embedding <=> query_embedding) > match_threshold
  ORDER BY qa_cache.embedding <=> query_embedding
  LIMIT match_count;
END;
$$;
