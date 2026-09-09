create extension if not exists vector;

create table if not exists public.team_documents (
  id uuid primary key default gen_random_uuid(),
  org_id uuid references public.enterprise_orgs(id) on delete cascade,
  content text not null,
  embedding vector(1536),
  metadata jsonb,
  created_at timestamptz default now()
);
create index on public.team_documents using ivfflat (embedding vector_cosine_ops);

create table if not exists public.team_documents_meta (
  id uuid primary key default gen_random_uuid(),
  org_id uuid references public.enterprise_orgs(id) on delete cascade,
  filename text,
  uploaded_by uuid references auth.users(id),
  size_bytes int,
  uploaded_at timestamptz default now()
);

create or replace function match_team_documents (
  query_embedding vector(1536),
  match_threshold float,
  match_count int,
  p_org_id uuid
)
returns table (
  id uuid,
  content text,
  metadata jsonb,
  similarity float
)
language sql stable
as $$
  select
    team_documents.id,
    team_documents.content,
    team_documents.metadata,
    1 - (team_documents.embedding <=> query_embedding) as similarity
  from team_documents
  where org_id = p_org_id
    and 1 - (team_documents.embedding <=> query_embedding) > match_threshold
  order by team_documents.embedding <=> query_embedding
  limit match_count;
$$;
