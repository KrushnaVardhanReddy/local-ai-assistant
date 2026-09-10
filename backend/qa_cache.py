import hashlib
import time
import sys
import chromadb
from config import config
from local_intelligence import get_local_intelligence

COLLECTION_NAME = "qa_cache"
SIMILARITY_THRESHOLD = 0.92
_client = None
_collection = None

def _get_collection():
    global _client, _collection
    if _collection is None:
        _client = chromadb.PersistentClient(path=config.CHROMA_DIR)
        _collection = _client.get_or_create_collection(
            name=COLLECTION_NAME,
            metadata={"hnsw:space": "cosine"}
        )
    return _collection

def lookup(question: str) -> str | None:
    """
    Returns cached answer string if a similar question was seen before.
    Returns None on cache miss.
    """
    t0 = time.time()
    try:
        col = _get_collection()
        results = col.query(query_texts=[question], n_results=1, include=["documents", "metadatas", "distances"])
        if results["ids"] and results["ids"][0]:
            distance = results["distances"][0][0]
            similarity = 1.0 - distance
            if similarity >= SIMILARITY_THRESHOLD:
                # Answer is now stored in metadata, not documents
                answer = results["metadatas"][0][0]["answer"]
                print(f"[QACache] {int((time.time() - t0) * 1000)}ms Hit (similarity={similarity:.3f})", file=sys.stderr)
                return answer
        print(f"[QACache] {int((time.time() - t0) * 1000)}ms Miss", file=sys.stderr)
    except Exception as e:
        print(f"[QACache] lookup error: {e}", file=sys.stderr)
    return None

def store(question: str, answer: str) -> None:
    """Stores a Q&A pair in the cache for future lookups."""
    try:
        col = _get_collection()

        doc_id = hashlib.md5(question.encode()).hexdigest()
        col.upsert(
            ids=[doc_id],
            documents=[question],
            metadatas=[{"answer": answer, "timestamp": str(int(time.time()))}]
        )
        print(f"[QACache] Stored new Q&A pair (total: {col.count()})", file=sys.stderr)
    except Exception as e:
        print(f"[QACache] store error: {e}", file=sys.stderr)


def store_bulk(qa_pairs: list[dict]) -> int:
    """
    Bulk stores a list of Q&A pairs.
    Each pair must have 'question' and 'answer' keys.
    Returns the number of pairs successfully stored.
    """
    li = get_local_intelligence()
    ids = []
    embeddings = []
    documents = []
    metadatas = []


    for pair in qa_pairs:
        q = pair.get("question", "")
        a = pair.get("answer", "")
        if not q or not a:
            continue
        vector = li.encode(q)
        if not vector:
            continue

        doc_id = hashlib.md5(q.encode()).hexdigest()
        ids.append(doc_id)
        embeddings.append(vector)
        documents.append(a)
        metadatas.append({"question": q[:200], "timestamp": str(int(time.time()))})

    if not ids:
        return 0

    try:
        col = _get_collection()
        col.upsert(
            ids=ids,
            embeddings=embeddings,
            documents=documents,
            metadatas=metadatas
        )
        print(f"[QACache] Bulk stored {len(ids)} Q&A pairs (total: {col.count()})", file=sys.stderr)
        return len(ids)
    except Exception as e:
        print(f"[QACache] bulk store error: {e}", file=sys.stderr)
        return 0

def clear() -> int:

    """Deletes all entries from the qa_cache collection. Returns count deleted."""
    try:
        col = _get_collection()
        count = col.count()
        all_ids = col.get(include=[])["ids"]
        if all_ids:
            col.delete(ids=all_ids)
        print(f"[QACache] Cleared {count} entries.", file=sys.stderr)
        return count
    except Exception as e:
        print(f"[QACache] clear error: {e}", file=sys.stderr)
        return 0

def stats() -> dict:
    """Returns cache statistics."""
    try:
        col = _get_collection()
        count = col.count()
        estimated_tokens_saved = count * 256  # rough avg tokens per answer
        return {"cached_pairs": count, "estimated_tokens_saved": estimated_tokens_saved}
    except Exception as e:
        return {"cached_pairs": 0, "estimated_tokens_saved": 0, "error": str(e)}
