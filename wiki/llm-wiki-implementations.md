# LLM Wiki Implementations

There are several existing implementations and frameworks for building LLM-maintained wikis. 

## Top Options

- **engram**: A lightweight, file-based system designed to be integrated directly into a project repository. This is ideal for developers and engineers who want version control and simplicity.
- **obsidian-llm-wiki**: A native solution for Obsidian users. It leverages Obsidian's graph view and visual browsing capabilities to navigate the AI-maintained knowledge graph.
- **Tome**: Focused on polished, well-formatted output presentation. Best used when the visual output and sharing of the knowledge base is the primary goal.
- **nashsu/llm_wiki**: A cross-platform desktop application that automatically converts local documents into an interlinked knowledge base.

## Relevance to Project Parakeet
For this project, we have effectively adopted the **engram** approach by integrating the wiki directly into the `Local_AI_Assistant` git repository, maintaining it as a set of flat markdown files with a raw source folder. This ensures the wiki stays lightweight and developer-friendly.
