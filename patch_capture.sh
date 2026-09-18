#!/bin/bash
cat << 'PATCH_EOF' | patch wails-app/backend/audio/capture.go
--- wails-app/backend/audio/capture.go
+++ wails-app/backend/audio/capture.go
@@ -194,6 +194,14 @@
	return c.stopCaptureInternal()
 }

+// IsCapturing returns true if the engine is actively capturing audio.
+func (c *CaptureEngine) IsCapturing() bool {
+	c.mu.Lock()
+	defer c.mu.Unlock()
+
+	return c.isInitialized && c.device != nil
+}
+
 // Stop stops the current audio capture, implementing the AudioCapturePort interface.
 func (c *CaptureEngine) Stop() {
	_ = c.StopCapture()
PATCH_EOF
