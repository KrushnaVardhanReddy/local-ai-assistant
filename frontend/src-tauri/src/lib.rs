use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Mutex;
use tauri::{Emitter, Manager};
use keyring::Entry;
use xcap::Monitor;
use base64::{Engine as _, engine::general_purpose::STANDARD};
use std::io::Cursor;
use image::{DynamicImage, RgbaImage};

fn read_local_settings() -> serde_json::Value {
    use std::fs;
    // Look for settings.json next to the executable
    if let Ok(exe_path) = std::env::current_exe() {
        if let Some(dir) = exe_path.parent() {
            let settings_path = dir.join("settings.json");
            if settings_path.exists() {
                if let Ok(content) = fs::read_to_string(&settings_path) {
                    if let Ok(json) = serde_json::from_str::<serde_json::Value>(&content) {
                        return json;
                    }
                }
            }
        }
    }
    serde_json::Value::Object(serde_json::Map::new())
}

static STEALTH_ENABLED: AtomicBool = AtomicBool::new(false);
static IN_MEMORY_TOKEN: Mutex<Option<String>> = Mutex::new(None);
static IN_MEMORY_MACHINE_ID: Mutex<Option<String>> = Mutex::new(None);

#[cfg(target_os = "macos")]
fn hide_from_dock() {
    use cocoa::appkit::{NSApp, NSApplicationActivationPolicy::NSApplicationActivationPolicyAccessory};
    // SAFETY: NSApp() returns a valid shared application instance, and calling
    // setActivationPolicy_ on it is safe and changes the app presentation state.
    unsafe {
        NSApp().setActivationPolicy_(NSApplicationActivationPolicyAccessory);
    }
}

#[cfg(not(target_os = "macos"))]
fn hide_from_dock() {
    // Handled by skipTaskbar on other platforms
}

// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

#[tauri::command]
fn toggle_stealth(window: tauri::WebviewWindow, enable: bool) {
    STEALTH_ENABLED.store(enable, Ordering::SeqCst);
    // Use Tauri's built-in content protection, which correctly applies
    // SetWindowDisplayAffinity to both the parent shell HWND and the inner
    // WebView2 child HWND via wry — fixing the Zoom screen-share visibility bug.
    let _ = window.set_content_protected(enable);
}

#[tauri::command]
fn quit_app(app: tauri::AppHandle) {
    app.exit(0);
}

#[tauri::command]
fn capture_screen() -> Result<String, String> {
    let monitors = Monitor::all().map_err(|e| e.to_string())?;

    // We only capture the primary display (the first one)
    if let Some(monitor) = monitors.first() {
        let image: RgbaImage = monitor.capture_image().map_err(|e| e.to_string())?;
        let dynamic_image = DynamicImage::ImageRgba8(image);

        let mut buffer = Cursor::new(Vec::new());
        dynamic_image.write_to(&mut buffer, image::ImageFormat::Jpeg).map_err(|e| e.to_string())?;

        let base64_string = STANDARD.encode(buffer.into_inner());
        Ok(format!("data:image/jpeg;base64,{}", base64_string))
    } else {
        Err("No monitors found".into())
    }
}

fn use_in_memory_keychain() -> bool {
    std::env::var("SKIP_KEYCHAIN").is_ok()
}

#[tauri::command]
fn save_token(token: String) -> Result<(), String> {
    if use_in_memory_keychain() {
        let mut mem_token = IN_MEMORY_TOKEN.lock().unwrap();
        *mem_token = Some(token);
        return Ok(());
    }

    let entry = Entry::new("local-ai-assistant", "auth-token").map_err(|e| e.to_string())?;
    entry.set_password(&token).map_err(|e| e.to_string())?;
    Ok(())
}

#[tauri::command]
fn load_token() -> Result<String, String> {
    if use_in_memory_keychain() {
        let mem_token = IN_MEMORY_TOKEN.lock().unwrap();
        if let Some(token) = &*mem_token {
            return Ok(token.clone());
        } else {
            return Err("No token found".to_string());
        }
    }

    let entry = Entry::new("local-ai-assistant", "auth-token").map_err(|e| e.to_string())?;
    entry.get_password().map_err(|e| e.to_string())
}

#[tauri::command]
fn delete_token() -> Result<(), String> {
    if use_in_memory_keychain() {
        let mut mem_token = IN_MEMORY_TOKEN.lock().unwrap();
        *mem_token = None;
        return Ok(());
    }

    let entry = Entry::new("local-ai-assistant", "auth-token").map_err(|e| e.to_string())?;
    entry.delete_password().map_err(|e| e.to_string())
}

#[tauri::command]
fn get_machine_id() -> Result<String, String> {
    if use_in_memory_keychain() {
        let mut mem_id = IN_MEMORY_MACHINE_ID.lock().unwrap();
        if let Some(id) = &*mem_id {
            return Ok(id.clone());
        } else {
            let new_id = uuid::Uuid::new_v4().to_string();
            *mem_id = Some(new_id.clone());
            return Ok(new_id);
        }
    }

    let entry = Entry::new("local-ai-assistant", "machine-id").map_err(|e| e.to_string())?;
    if let Ok(id) = entry.get_password() {
        Ok(id)
    } else {
        let new_id = uuid::Uuid::new_v4().to_string();
        let _ = entry.set_password(&new_id);
        Ok(new_id)
    }
}

#[tauri::command]
fn get_app_display_name() -> String {
    let settings = read_local_settings();
    settings
        .get("APP_DISPLAY_NAME")
        .and_then(|v| v.as_str())
        .unwrap_or("AudioService")
        .to_string()
}

#[tauri::command]
fn set_clickthrough(window: tauri::WebviewWindow, enable: bool) -> Result<(), String> {
    window.set_ignore_cursor_events(enable).map_err(|e| e.to_string())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_global_shortcut::Builder::new().build())
        .setup(|app| {
            hide_from_dock();

            let settings = read_local_settings();
            let app_name = settings
                .get("APP_DISPLAY_NAME")
                .and_then(|v| v.as_str())
                .unwrap_or("AudioService")
                .to_string();

            if let Some(win) = app.get_webview_window("main") {
                let _ = win.set_title(&app_name);
                // contentProtected: true in tauri.conf.json already applies protection
                // declaratively before this point, but we set it explicitly here too
                // as a belt-and-suspenders guarantee before the window is shown.
                let _ = win.set_content_protected(true);
                // Now safe to show — the window is already invisible to capture tools.
                let _ = win.show();
            }

            use tauri_plugin_global_shortcut::GlobalShortcutExt;

            app.global_shortcut().on_shortcut("Ctrl+Shift+Space", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    if let Some(win) = app.get_webview_window("main") {
                        if let Ok(visible) = win.is_visible() {
                            if visible {
                                let _ = win.hide();
                            } else {
                                let _ = win.show();
                                let _ = win.set_focus();
                            }
                        }
                    }
                }
            }).expect("failed to register global shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+S", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("trigger-vision", ());
                }
            }).expect("failed to register Ctrl+Shift+S shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+P", |app, _shortcut, event| {
                match event.state() {
                    tauri_plugin_global_shortcut::ShortcutState::Pressed => {
                        let _ = app.emit("ptt-start", ());
                    }
                    tauri_plugin_global_shortcut::ShortcutState::Released => {
                        let _ = app.emit("ptt-stop", ());
                    }
                }
            }).expect("failed to register Ctrl+Shift+P shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+Down", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("scroll-down", ());
                }
            }).expect("failed to register Ctrl+Shift+Down shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+Up", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("scroll-up", ());
                }
            }).expect("failed to register Ctrl+Shift+Up shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+X", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("panic-clear", ());
                }
            }).expect("failed to register Ctrl+Shift+X shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+M", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("toggle-clickthrough", ());
                }
            }).expect("failed to register Ctrl+Shift+M shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+1", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("hotkey_transcript_1", ());
                }
            }).expect("failed to register Ctrl+Shift+1 shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+2", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("hotkey_transcript_2", ());
                }
            }).expect("failed to register Ctrl+Shift+2 shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+3", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("hotkey_transcript_3", ());
                }
            }).expect("failed to register Ctrl+Shift+3 shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+4", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("hotkey_transcript_4", ());
                }
            }).expect("failed to register Ctrl+Shift+4 shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+5", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("hotkey_transcript_5", ());
                }
            }).expect("failed to register Ctrl+Shift+5 shortcut");

            app.global_shortcut().on_shortcut("Ctrl+Shift+6", |app, _shortcut, event| {
                if event.state() == tauri_plugin_global_shortcut::ShortcutState::Pressed {
                    let _ = app.emit("hotkey_transcript_6", ());
                }
            }).expect("failed to register Ctrl+Shift+6 shortcut");

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            greet,
            toggle_stealth,
            quit_app,
            capture_screen,
            save_token,
            load_token,
            delete_token,
            get_machine_id,
            get_app_display_name,
            set_clickthrough
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::env;

    #[test]
    fn test_in_memory_keychain_logic() {
        // Set env var to force in-memory keychain for testing
        env::set_var("SKIP_KEYCHAIN", "1");
        assert!(use_in_memory_keychain());

        // Reset state
        {
            let mut mem_token = IN_MEMORY_TOKEN.lock().unwrap();
            *mem_token = None;
            let mut mem_id = IN_MEMORY_MACHINE_ID.lock().unwrap();
            *mem_id = None;
        }

        // Test save & load token
        assert!(load_token().is_err());
        assert!(save_token("test-token-123".to_string()).is_ok());
        assert_eq!(load_token().unwrap(), "test-token-123");

        // Test delete token
        assert!(delete_token().is_ok());
        assert!(load_token().is_err());

        // Test get_machine_id
        let id1 = get_machine_id().unwrap();
        assert!(!id1.is_empty());

        let id2 = get_machine_id().unwrap();
        assert_eq!(id1, id2); // Should return the same ID
    }

    #[test]
    fn test_capture_screen_is_result() {
        // We just ensure it compiles and either succeeds or fails properly
        // rather than failing outright because we might not have a display in CI
        let result = capture_screen();
        match result {
            Ok(base64_str) => {
                assert!(base64_str.starts_with("data:image/jpeg;base64,"));
            }
            Err(e) => {
                // In headless environments without X11/Wayland, it might fail to find monitors.
                // We just verify it returns a clean string error.
                assert!(!e.is_empty());
            }
        }
    }
}
