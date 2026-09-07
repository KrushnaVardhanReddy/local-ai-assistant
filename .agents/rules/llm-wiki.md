---
description: Rules for maintaining the LLM Wiki in this project
---

# LLM Wiki Maintenance Rules

This project uses an LLM Wiki located in the `wiki/` directory to store and compound knowledge. All AI agents (Antigravity, OpenCode, Jules) must adhere to these rules when interacting with or maintaining the wiki.

## Directory Structure
- `wiki/raw/`: Immutable source documents (e.g. PDFs, external articles). Agents must NEVER modify files here.
- `wiki/`: LLM-maintained markdown knowledge base. Agents SHOULD modify files here.
- `wiki/index.md`: Catalog of all files in the wiki.
- `wiki/log.md`: Chronological, append-only log of changes.

## 1. Ingesting New Knowledge
When asked to ingest a document from `wiki/raw/` or from a user message:
1. Read the source carefully.
2. Create or update relevant markdown pages in the `wiki/` folder with the new information.
3. Update `wiki/index.md` to include any newly created pages.
4. Append an entry to `wiki/log.md` with the format `## [YYYY-MM-DD] ingest | <Brief description>`.

## 2. Querying and Synthesizing
When the user asks questions that require synthesizing project knowledge:
1. Start by reading `wiki/index.md` to locate relevant pages.
2. Read the relevant pages.
3. Answer the user's question.
4. **CRITICAL**: If your answer contains valuable new synthesis, comparisons, or analysis, file it back into the wiki as a new markdown page, update `index.md`, and append to `log.md`. Do not let good answers disappear into chat history.

## 3. Maintenance (Linting)
Periodically, or when asked to "lint" the wiki:
1. Check for contradictions or stale claims.
2. Check for missing cross-references between pages.
3. Check `wiki/index.md` to ensure it accurately reflects all files in `wiki/`.
4. Make the necessary updates and log the pass in `wiki/log.md`.

## Principles
- You are the maintainer. The human is the curator. 
- You must do the grunt work: cross-referencing, updating indices, logging changes.
- Never lose knowledge. Always persist it to the wiki.
