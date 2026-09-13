1. **Create RPC Migration for pgvector search:**
   - Create `supabase/migrations/004_add_qa_cache_rpc.sql` containing a PostgreSQL function `match_qa_cache` that performs a cosine similarity vector search on the `qa_cache` table.
   - The function will accept `query_embedding`, `match_threshold`, `match_count`, and crucially, `p_user_id` to strictly enforce multi-tenant isolation.

2. **Extract `userId` in the Worker:**
   - In `cloud-worker/src/index.ts`, define a `let userId = '';` inside the `fetch` function block.
   - When authorizing in the Supabase API validation step, capture `userId = data[0].user_id`.

3. **Handle `/api/cache/prewarm` (POST) in Worker:**
   - Parse the JSON payload which will contain a list of Q&A pairs along with their pre-computed embeddings.
   - Map the payload into an array of objects structured for Supabase: `{user_id: userId, question: q, answer: a, embedding: e}`.
   - Insert the resulting records into Supabase via `POST /rest/v1/qa_cache` using `fetch`.

4. **Handle `/api/ask` (POST) in Worker:**
   - Add a route block for `POST /api/ask`.
   - Parse the request payload, which will contain the pre-computed embedding.
   - Call the Supabase REST API `POST /rest/v1/rpc/match_qa_cache` with the embedding and `p_user_id = userId`.
   - Return `{ cached_answer: match.answer }` if a valid match is found, or an appropriate cache miss response.

5. **Handle `/api/cache` (DELETE) in Worker:**
   - Parse the JSON payload to extract an array of `ids`.
   - Call the Supabase REST API using `DELETE /rest/v1/qa_cache?id=in.(${ids.join(',')})&user_id=eq.${userId}` to clear only the specific vectors matching the provided IDs for the authenticated user.

6. **Verify Source Code:**
   - Read the modified files using `cat` and `ls` to ensure all implementations are correct and the syntax is valid.

7. **Pre Commit Steps:**
   - Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.

8. **Run Tests:**
   - Execute all relevant tests to confirm the integration works and hasn't introduced regressions.
