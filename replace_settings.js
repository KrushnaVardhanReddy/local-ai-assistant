const fs = require('fs');

let content = fs.readFileSync('wails-app/frontend/src/lib/Settings.svelte', 'utf-8');

// Replace imports
content = content.replace(/import { invoke } from "@tauri-apps\/api\/core";\n/g, '');

// Replace invoke
content = content.replace(/await invoke\("toggle_stealth", { enable: !isDevModeChecked }\);/g, 'await (window as any).go.main.App.ToggleStealth({ enable: !isDevModeChecked });');

fs.writeFileSync('wails-app/frontend/src/lib/Settings.svelte', content);
