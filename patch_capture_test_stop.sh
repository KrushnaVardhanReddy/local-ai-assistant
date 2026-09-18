#!/bin/bash
cat << 'PATCH_EOF' | patch wails-app/backend/audio/capture_test.go
--- wails-app/backend/audio/capture_test.go
+++ wails-app/backend/audio/capture_test.go
@@ -107,6 +107,11 @@
	}
 }

+func TestStop(t *testing.T) {
+	engine := NewCaptureEngine()
+	engine.Stop()
+}
+
 func TestIsCapturing(t *testing.T) {
	engine := NewCaptureEngine()

PATCH_EOF
