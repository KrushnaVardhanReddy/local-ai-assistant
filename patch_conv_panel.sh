#!/bin/bash
cat << 'PATCH_EOF' | patch wails-app/frontend/src/products/interview/ConvPanel.svelte
--- wails-app/frontend/src/products/interview/ConvPanel.svelte
+++ wails-app/frontend/src/products/interview/ConvPanel.svelte
@@ -64,11 +64,15 @@
     <div class="header-left">
       <span class="title">Live Session</span>
       {#if isListening && !isMockMode}
-        <div class="status-badge status-live pulse">● Live</div>
+        <button class="status-badge status-live pulse clickable" onclick={() => (window as any).go.main.App.ToggleMic()}>
+          <span class="material-symbols-outlined" style="font-size: 12px; margin-right: 2px;">mic</span> Live
+        </button>
       {:else if isMockMode}
         <div class="status-badge status-mock pulse">● Mock</div>
       {:else}
-        <div class="status-badge status-mic-off">● Mic Off</div>
+        <button class="status-badge status-mic-off clickable" onclick={() => (window as any).go.main.App.ToggleMic()}>
+          <span class="material-symbols-outlined" style="font-size: 12px; margin-right: 2px;">mic_off</span> Mic Off
+        </button>
       {/if}

       {#if isPTTHeld}
@@ -205,6 +209,10 @@
     gap: 4px;
     font-weight: 600;
     white-space: nowrap;
+    cursor: default;
+  }
+  .status-badge.clickable {
+    cursor: pointer;
   }

   .status-live {
PATCH_EOF
