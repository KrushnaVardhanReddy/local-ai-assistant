#!/bin/bash
cat << 'PATCH_EOF' | patch prompts/tasks_v2.md
--- prompts/tasks_v2.md
+++ prompts/tasks_v2.md
@@ -107,3 +107,5 @@


 | **P57-T9** | Cache History UI & Vertical Transcripts | ✅ Merged | Sequential | [PR 200] | Updates ConvPanel for vertical transcripts and AnswerPanel for cache history browsing. |
+| **P57-T10** | Microphone Mute Toggle | ✅ Merged | Sequential | [#P57-T10] | Update backend to toggle PortAudio/PulseAudio capture state when user clicks Mic Off/Live in ConvPanel header. |
+
PATCH_EOF
