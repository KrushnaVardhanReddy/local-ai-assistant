const fs = require('fs');

let content = fs.readFileSync('wails-app/frontend/src/lib/auth.svelte.ts', 'utf-8');

// Replace imports
content = content.replace(/import { invoke } from '@tauri-apps\/api\/core';\n/g, '');

// Replace invoke
content = content.replace(/await invoke\("load_token"\);/g, 'await (window as any).go.main.App.LoadToken();');
content = content.replace(/await invoke\("delete_token"\);/g, 'await (window as any).go.main.App.DeleteToken();');
content = content.replace(/await invoke\("save_token", { token }\);/g, 'await (window as any).go.main.App.SaveToken({ token });');

fs.writeFileSync('wails-app/frontend/src/lib/auth.svelte.ts', content);
