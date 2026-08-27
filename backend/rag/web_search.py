from duckduckgo_search import DDGS

def search_web(query: str, max_results: int = 3) -> str:
    """
    Synchronously fetches search results from DuckDuckGo.
    Returns a formatted string of the top results.
    """
    try:
        ddgs = DDGS()
        results = ddgs.text(query, max_results=max_results)

        if not results:
            return ""

        formatted_results = ["--- Internet Search Results ---"]
        for res in results:
            href = res.get("href", "")
            title = res.get("title", "")
            body = res.get("body", "")
            formatted_results.append(f"[Source: {href}] {title}\n{body}")

        return "\n".join(formatted_results)
    except Exception as e:
        import sys
        print(f"Web search error: {e}", file=sys.stderr)
        return ""
