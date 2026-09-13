const fs = require('fs');

let content = fs.readFileSync('wails-app/frontend/src/lib/Assistant.svelte', 'utf-8');

// Replace imports
content = content.replace(/import { listen } from "@tauri-apps\/api\/event";/g, 'import { EventsOn, EventsOff, WindowHide } from "../../wailsjs/runtime/runtime";');
content = content.replace(/import { invoke } from "@tauri-apps\/api\/core";/g, '');
content = content.replace(/import { getCurrentWindow } from "@tauri-apps\/api\/window";/g, '');

// Replace unlistens
// Note: In Wails EventsOn doesn't return an unlisten function. We just call EventsOff("event-name") on destroy.
// Let's replace the listeners first.
content = content.replace(/const unlistenScrollDown = listen\("scroll-down", \(\) => {/g, 'EventsOn("scroll-down", () => {');
content = content.replace(/const unlistenScrollUp = listen\("scroll-up", \(\) => {/g, 'EventsOn("scroll-up", () => {');
content = content.replace(/const unlisten = listen\("trigger-vision", async \(\) => {/g, 'EventsOn("trigger-vision", async () => {');
content = content.replace(/const unlistenClickthrough = listen\("toggle-clickthrough", \(\) => {/g, 'EventsOn("toggle-clickthrough", () => {');
content = content.replace(/const unlistenTranscript1 = listen\("hotkey_transcript_1", \(\) => {/g, 'EventsOn("hotkey_transcript_1", () => {');
content = content.replace(/const unlistenTranscript2 = listen\("hotkey_transcript_2", \(\) => {/g, 'EventsOn("hotkey_transcript_2", () => {');
content = content.replace(/const unlistenTranscript3 = listen\("hotkey_transcript_3", \(\) => {/g, 'EventsOn("hotkey_transcript_3", () => {');
content = content.replace(/const unlistenTranscript4 = listen\("hotkey_transcript_4", \(\) => {/g, 'EventsOn("hotkey_transcript_4", () => {');
content = content.replace(/const unlistenTranscript5 = listen\("hotkey_transcript_5", \(\) => {/g, 'EventsOn("hotkey_transcript_5", () => {');
content = content.replace(/const unlistenTranscript6 = listen\("hotkey_transcript_6", \(\) => {/g, 'EventsOn("hotkey_transcript_6", () => {');

// Replace destroy unlistens
content = content.replace(/unlisten\.then\(f => f\(\)\);/g, 'EventsOff("trigger-vision");');
content = content.replace(/unlistenScrollDown\.then\(f => f\(\)\);/g, 'EventsOff("scroll-down");');
content = content.replace(/unlistenScrollUp\.then\(f => f\(\)\);/g, 'EventsOff("scroll-up");');
content = content.replace(/unlistenClickthrough\.then\(f => f\(\)\);/g, 'EventsOff("toggle-clickthrough");');
content = content.replace(/unlistenTranscript1\.then\(f => f\(\)\);/g, 'EventsOff("hotkey_transcript_1");');
content = content.replace(/unlistenTranscript2\.then\(f => f\(\)\);/g, 'EventsOff("hotkey_transcript_2");');
content = content.replace(/unlistenTranscript3\.then\(f => f\(\)\);/g, 'EventsOff("hotkey_transcript_3");');
content = content.replace(/unlistenTranscript4\.then\(f => f\(\)\);/g, 'EventsOff("hotkey_transcript_4");');
content = content.replace(/unlistenTranscript5\.then\(f => f\(\)\);/g, 'EventsOff("hotkey_transcript_5");');
content = content.replace(/unlistenTranscript6\.then\(f => f\(\)\);/g, 'EventsOff("hotkey_transcript_6");');

// Replace window controls
content = content.replace(/await getCurrentWindow\(\)\.startDragging\(\);/g, '// Dragging handled by CSS in Wails');
content = content.replace(/await getCurrentWindow\(\)\.hide\(\);/g, 'WindowHide();');

// Replace invoke
content = content.replace(/await invoke\("quit_app"\);/g, 'await (window as any).go.main.App.QuitApp();');
content = content.replace(/await invoke<string>\("capture_screen"\);/g, 'await (window as any).go.main.App.CaptureScreen();');
content = content.replace(/await invoke\('set_clickthrough', { enable: clickthrough }\);/g, 'await (window as any).go.main.App.SetClickthrough({ enable: clickthrough });');
content = content.replace(/invoke<string>\("capture_screen"\)/g, '(window as any).go.main.App.CaptureScreen()');

// Replace onmousedown startDrag with Wails CSS attribute directly in the Svelte code.
// Actually, setting style="--wails-draggable:drag;" is better, but maybe we can just keep the empty startDrag func, it's harmless and does nothing now.
// Let's modify the onmousedown div styles directly using regex if possible.
content = content.replace(/onmousedown=\{startDrag\}/g, 'style="--wails-draggable:drag"');

fs.writeFileSync('wails-app/frontend/src/lib/Assistant.svelte', content);
