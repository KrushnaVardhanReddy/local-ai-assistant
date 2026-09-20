import sys

content = open("wails-app/backend/llm/scorecard_test.go").read()

content = content.replace('os.Setenv("OPENAI_API_KEY", "test-key")', 'SetProxyToken("")\n\tos.Setenv("OPENAI_API_KEY", "test-key")')

open("wails-app/backend/llm/scorecard_test.go", "w").write(content)
