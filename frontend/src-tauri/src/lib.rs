use std::sync::atomic::{AtomicBool, Ordering};
use tauri::Manager;

static STEALTH_ENABLED: AtomicBool = AtomicBool::new(false);

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
fn toggle_stealth(window: tauri::WebviewWindow, enabled: bool) {
    STEALTH_ENABLED.store(enabled, Ordering::SeqCst);
    set_screen_share_safe(&window, enabled);
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

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![greet, toggle_stealth])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
