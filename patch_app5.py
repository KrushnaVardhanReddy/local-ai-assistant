import sys
with open('wails-app/app.go', 'r') as f:
    content = f.read()

# Change the import of wails-app/adapters/llm to have an alias
content = content.replace('\t"wails-app/adapters/llm"\n', '\tllmadapter "wails-app/adapters/llm"\n')
content = content.replace('llm.NewOpenAIAdapter()', 'llmadapter.NewOpenAIAdapter()')
content = content.replace('\t"wails-app/adapters/cache"\n', '\tcacheadapter "wails-app/adapters/cache"\n')
content = content.replace('cache.NewSQLiteVecAdapter(db)', 'cacheadapter.NewSQLiteVecAdapter(db)')
content = content.replace('\t"wails-app/adapters/events"\n', '\teventsadapter "wails-app/adapters/events"\n')
content = content.replace('events.NewWailsEventAdapter(ctx)', 'eventsadapter.NewWailsEventAdapter(ctx)')

with open('wails-app/app.go', 'w') as f:
    f.write(content)
