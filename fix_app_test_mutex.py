import sys

content = open("wails-app/backend/llm/openai.go").read()

if "sync.Mutex" not in content:
    content = content.replace("var DemoProxyToken string\n\nfunc SetProxyToken(token string) {\n\tDemoProxyToken = token\n}", """import_placeholder""")

    if 'import (' in content:
        content = content.replace('import (', 'import (\n\t"sync"\n')

    content = content.replace("import_placeholder", """var (
	DemoProxyToken string
	proxyTokenMutex sync.RWMutex
)

func SetProxyToken(token string) {
	proxyTokenMutex.Lock()
	defer proxyTokenMutex.Unlock()
	DemoProxyToken = token
}

func GetProxyToken() string {
	proxyTokenMutex.RLock()
	defer proxyTokenMutex.RUnlock()
	return DemoProxyToken
}""")

    content = content.replace("DemoProxyToken != \"\"", "GetProxyToken() != \"\"")
    content = content.replace("apiKey = DemoProxyToken", "apiKey = GetProxyToken()")

open("wails-app/backend/llm/openai.go", "w").write(content)

content = open("wails-app/backend/llm/openai_test.go").read()
content = content.replace("DemoProxyToken !=", "GetProxyToken() !=")
content = content.replace("got '%s'\", DemoProxyToken", "got '%s'\", GetProxyToken()")
open("wails-app/backend/llm/openai_test.go", "w").write(content)
