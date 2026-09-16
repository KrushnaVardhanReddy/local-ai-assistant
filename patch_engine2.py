import sys
with open('wails-app/core/engine/engine.go', 'r') as f:
    content = f.read()

content += '\n// GetCache returns the cache port instance.\nfunc (e *StealthEngine) GetCache() driven.CachePort {\n\treturn e.cache\n}\n'

with open('wails-app/core/engine/engine.go', 'w') as f:
    f.write(content)

# Update app.go to cast GetCache to *backend.VectorDB
with open('wails-app/app.go', 'r') as f:
    app_content = f.read()

app_content = app_content.replace('''	if a.engine.GetCache() == nil {
		return []backend.CacheItem{}
	}
	items, err := a.engine.GetCache().GetAllItems()''', '''	cacheAdapter, ok := a.engine.GetCache().(interface{ GetAllItems() ([]backend.CacheItem, error) })
	if !ok {
		return []backend.CacheItem{}
	}
	items, err := cacheAdapter.GetAllItems()''')

app_content = app_content.replace('''	if a.engine.GetCache() == nil {
		return nil
	}
	for _, id := range ids {
		_ = a.engine.GetCache().DeleteItem(id)
	}''', '''	cacheAdapter, ok := a.engine.GetCache().(interface{ DeleteItem(string) error })
	if !ok {
		return nil
	}
	for _, id := range ids {
		_ = cacheAdapter.DeleteItem(id)
	}''')

app_content = app_content.replace('''	if a.engine.GetCache() == nil {
		return nil
	}
	return a.engine.GetCache().ClearAll()''', '''	cacheAdapter, ok := a.engine.GetCache().(interface{ ClearAll() error })
	if !ok {
		return nil
	}
	return cacheAdapter.ClearAll()''')

with open('wails-app/app.go', 'w') as f:
    f.write(app_content)
