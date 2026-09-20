import sys

content = open("wails-app/app.go").read()

content = content.replace("""func (a *App) SummarizeSession(requests []engine.SummaryRequest) error {
	return a.engine.SummarizeSession(requests)
}""", """func (a *App) SummarizeSession(requests []engine.SummaryRequest) error {
	return a.engine.SummarizeSession(requests)
}

func (a *App) SetProxyToken(token string) {
	llm.SetProxyToken(token)
}""")

open("wails-app/app.go", "w").write(content)
