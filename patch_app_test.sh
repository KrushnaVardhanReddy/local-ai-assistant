#!/bin/bash
cat << 'PATCH_EOF' | patch wails-app/app_test.go
--- wails-app/app_test.go
+++ wails-app/app_test.go
@@ -107,3 +107,17 @@
		t.Error("Expected error when context is nil")
	}
 }
+
+func TestApp_ToggleMic(t *testing.T) {
+	app := NewApp()
+
+	// Wait for any async initialization to avoid races, though NewApp doesn't start capture by itself
+
+	// Initial toggle should return true (start capture) or false (fail to start due to CI).
+	// It relies on SetAudioDevice which might fail in CI and print log but doesn't panic.
+	// Actually, SetAudioDevice returns an error if it fails, and the capturing state might not change.
+	app.ToggleMic()
+
+	// Another toggle
+	app.ToggleMic()
+}
PATCH_EOF
