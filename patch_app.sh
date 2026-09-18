#!/bin/bash
cat << 'PATCH_EOF' | patch wails-app/app.go
--- wails-app/app.go
+++ wails-app/app.go
@@ -510,7 +510,10 @@
		activeDocName = activeDoc.Name
	}

-	isListening := a.audioCapture != nil
+	isListening := false
+	if a.audioCapture != nil {
+		isListening = a.audioCapture.IsCapturing()
+	}

	s := a.engine.GetState()

@@ -528,4 +531,18 @@
		"cachedPairs":             s.CachedPairs,
		"includeActiveDocContext": a.engine.GetIncludeActiveDocContext(),
	}
 }
+
+// ToggleMic toggles the microphone capturing state.
+// Returns true if the microphone was started, false if it was stopped.
+func (a *App) ToggleMic() bool {
+	if a.audioCapture == nil {
+		return false
+	}
+	if a.audioCapture.IsCapturing() {
+		a.audioCapture.StopCapture()
+		return false
+	} else {
+		a.SetAudioDevice(-1, false)
+		return true
+	}
+}
PATCH_EOF
