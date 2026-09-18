#!/bin/bash
cat << 'PATCH_EOF' | patch wails-app/backend/audio/capture_test.go
--- wails-app/backend/audio/capture_test.go
+++ wails-app/backend/audio/capture_test.go
@@ -107,6 +107,26 @@
	}
 }

+func TestIsCapturing(t *testing.T) {
+	engine := NewCaptureEngine()
+
+	// Uninitialized
+	if engine.IsCapturing() {
+		t.Error("Expected IsCapturing to be false on uninitialized engine")
+	}
+
+	if err := engine.Initialize(); err != nil {
+		t.Skipf("Skipping test due to portaudio initialization failure: %v", err)
+	}
+	defer engine.Terminate()
+
+	// Initialized but not capturing
+	if engine.IsCapturing() {
+		t.Error("Expected IsCapturing to be false when not capturing")
+	}
+	// We don't test the true state here as starting a real capture device in CI could fail and is covered by StartCapture tests anyway.
+}
+
 func generateAudioChunk(rmsTarget float64, numSamples int) []byte {
	// Simple square wave to get target RMS
	val := float64(0)
PATCH_EOF
