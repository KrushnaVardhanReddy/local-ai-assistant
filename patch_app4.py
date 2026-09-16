import sys
with open('wails-app/app.go', 'r') as f:
    content = f.read()

content = content.replace('"wails-app/adapters/cacheadapter"', '"wails-app/adapters/cache"')
content = content.replace('"wails-app/adapters/eventsadapter"', '"wails-app/adapters/events"')
content = content.replace('"wails-app/adapters/llmadapter"', '"wails-app/adapters/llm"')

content = content.replace('llmadapter.NewOpenAIAdapter()', 'llm.NewOpenAIAdapter()')
content = content.replace('cacheadapter.NewSQLiteVecAdapter(db)', 'cache.NewSQLiteVecAdapter(db)')
content = content.replace('eventsadapter.NewWailsEventAdapter(ctx)', 'events.NewWailsEventAdapter(ctx)')

# Ensure we don't have duplicated import aliases like llm "wails-app/backend/llm" and "wails-app/adapters/llm"
# Check if "wails-app/backend/llm" is used.
with open('wails-app/app.go', 'w') as f:
    f.write(content)
