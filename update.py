import sys

content = open("wails-app/backend/llm/openai.go").read()

content = content.replace("""type ImageURLContent struct {
	URL string `json:"url"`
}""", """var DemoProxyToken string

func SetProxyToken(token string) {
	DemoProxyToken = token
}

type ImageURLContent struct {
	URL string `json:"url"`
}""")

content = content.replace("""	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}
	baseURL := getEnvOrDefault("LLM_BASE_URL", "https://api.openai.com/v1")""", """	var baseURL, apiKey string
	if DemoProxyToken != "" {
		baseURL = getEnvOrDefault("SUPABASE_EDGE_URL", "http://127.0.0.1:54321/functions/v1/llm-proxy")
		apiKey = DemoProxyToken
	} else {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY environment variable is not set")
		}
		baseURL = getEnvOrDefault("LLM_BASE_URL", "https://api.openai.com/v1")
	}""")

open("wails-app/backend/llm/openai.go", "w").write(content)
