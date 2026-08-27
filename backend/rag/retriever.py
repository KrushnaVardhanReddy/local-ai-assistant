from fastapi import APIRouter, Query
from typing import Optional

from rag.ingestor import _collection, _embedder, get_collection, get_embedder
from config import config

retriever_router = APIRouter()

def retrieve(query: str, top_k: Optional[int] = None) -> str:
    """
    Embeds the query, searches ChromaDB for top_k relevant chunks,
    deduplicates them, and returns a formatted context string.
    """
    if top_k is None:
        top_k = config.RAG_TOP_K

    # The prompt explicitly asked to import _collection and _embedder singletons.
    # However, ingestor.py initializes them lazily via get_collection() and get_embedder().
    # If they are None, we must initialize them here (which get_collection/get_embedder does),
    # but we will use the imported singletons as requested if they are available.

    collection = _collection if _collection is not None else get_collection()
    if collection is None:
        return ""

    count = collection.count()
    if count == 0:
        return ""

    embedder = _embedder if _embedder is not None else get_embedder()
    query_embedding = embedder.encode([query]).tolist()[0]

    n_results_to_fetch = min(top_k * 2, count)

    results = collection.query(
        query_embeddings=[query_embedding],
        n_results=n_results_to_fetch,
        include=["documents", "metadatas", "distances"]
    )

    if not results or not results["documents"] or not results["documents"][0]:
        return ""

    # ChromaDB returns a list of lists because you can pass multiple queries
    documents = results["documents"][0]
    metadatas = results["metadatas"][0]
    # distances = results["distances"][0] # distances are returned sorted by default

    kept = []
    seen_sources: dict[str, int] = {}

    for doc, meta in zip(documents, metadatas):
        source = meta.get("source_file", "unknown") if meta else "unknown"

        if seen_sources.get(source, 0) >= 2:
            continue

        seen_sources[source] = seen_sources.get(source, 0) + 1

        chunk_index = meta.get("chunk_index", "?") if meta else "?"
        kept.append((source, chunk_index, doc))

        if len(kept) == top_k:
            break

    if not kept:
        return ""

    context_blocks = ["--- Relevant context from your knowledge base ---"]
    for source, chunk_index, text in kept:
        context_blocks.append(f"[Source: {source}, chunk {chunk_index}]\n{text}")

    return "\n\n".join(context_blocks)


@retriever_router.get("/search")
async def search_endpoint(q: str = Query(..., alias="q"), k: Optional[int] = Query(None, alias="k")):
    """FastAPI endpoint to test retrieval."""
    result = retrieve(q, k)
    return {"context": result, "empty": result == ""}
