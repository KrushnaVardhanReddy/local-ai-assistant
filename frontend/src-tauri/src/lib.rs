use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Mutex;
use tauri::{Emitter, Manager};
use keyring::Entry;
use xcap::Monitor;
use base64::{Engine as _, engine::general_purpose::STANDARD};
use std::io::Cursor;
use image::{DynamicImage, RgbaImage};

static STEALTH_ENABLED: AtomicBool = AtomicBool::new(false);
static IN_MEMORY_TOKEN: Mutex<Option<String>> = Mutex::new(None);
static IN_MEMORY_MACHINE_ID: Mutex<Option<String>> = Mutex::new(None);

#[cfg(target_os = "macos")]
fn set_screen_share_safe(window: &tauri::WebviewWindow, enabled: bool) {
    use cocoa::appkit::NSWindow;
    use objc::runtime::Object;
    // SAFETY: We get the underlying NSWindow handle safely from Tauri,
    // cast it to an objc Object pointer, and use Objective-C runtime message
    // sending which is well-defined for setSharingType.
    unsafe {
        if let Ok(ns_win) = window.ns_window() {
            let ns_win = ns_win as *mut Object;
            let val: u64 = if enabled { 0 } else { 1 }; // NSWindowSharingNone = 0, NSWindowSharingReadOnly = 1
            let _: () = objc::msg_send![ns_win, setSharingType: val];
        }
    }
}

#[cfg(target_os = "windows")]
fn set_screen_share_safe(window: &tauri::WebviewWindow, enabled: bool) {
    use windows::Win32::Foundation::HWND;
    use windows::Win32::UI::WindowsAndMessaging::{SetWindowDisplayAffinity, WDA_EXCLUDEFROMCAPTURE, WDA_MONITOR, WDA_NONE};

    if let Ok(hwnd) = window.hwnd() {
        let hwnd = HWND(hwnd.0);
        let affinity = if enabled {
            WDA_EXCLUDEFROMCAPTURE
        } else {
            WDA_NONE
        };

        // SAFETY: The HWND provided by Tauri is valid and owned by the application.
        // SetWindowDisplayAffinity is safe to call on windows belonging to the current process.
        unsafe {
            let res = SetWindowDisplayAffinity(hwnd, affinity);
            if res.is_err() && enabled {
                // Fallback for older Windows 10 versions
                let _ = SetWindowDisplayAffinity(hwnd, WDA_MONITOR);
            }
        }
    }
}

#[cfg(any(target_os = "linux", target_os = "android", target_os = "ios"))]
fn set_screen_share_safe(_window: &tauri::WebviewWindow, _enabled: bool) {
    println!("Warning: set_screen_share_safe is a no-op on Linux, as compositors handle this natively.");
}

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
    set_screen_share_safe(&window, enable);
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

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_global_shortcut::Builder::new().build())
        .setup(|app| {
            hide_from_dock();

            if let Some(win) = app.get_webview_window("main") {
                // Initialize in stealth mode off.
                set_screen_share_safe(&win, false);
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
                    _ => {}
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

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            greet,
            toggle_stealth,
            capture_screen,
            save_token,
            load_token,
            delete_token,
            get_machine_id
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
