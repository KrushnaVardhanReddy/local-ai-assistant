import re

# Update tasks.md
with open('prompts/tasks.md', 'r') as f:
    t = f.read()
t = t.replace('| P26-T4 | `backend/app.py`, `frontend/src/lib/SessionReport.svelte` | **Auto-CRM Sync:** Add a "Push to Salesforce/HubSpot" button in the Session Report panel to log meeting notes directly to the CRM. | ⬜ | — |',
              '| P26-T4 | `backend/app.py`, `frontend/src/lib/SessionReport.svelte` | **Auto-CRM Sync:** Add a "Push to Salesforce/HubSpot" button in the Session Report panel to log meeting notes directly to the CRM. | ⬜ | — |\n| P26-T5 | `web/tests/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright and Pytest tests for Enterprise B2B features. | ⬜ | — |')
t = t.replace('| P27-T3 | `frontend/src/lib/Assistant.svelte` | **Voice Mode UI:** Add a visual indicator (like a glowing orb) when the assistant is actively listening/speaking, bypassing the stealth chat UI. | ⬜ | — |',
              '| P27-T3 | `frontend/src/lib/Assistant.svelte` | **Voice Mode UI:** Add a visual indicator (like a glowing orb) when the assistant is actively listening/speaking, bypassing the stealth chat UI. | ⬜ | — |\n| P27-T4 | `frontend/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright and Pytest tests for Voice Mode. | ⬜ | — |')
t = t.replace('| P28-T4 | `backend/mcp/` | **Model Context Protocol (MCP):** Add an MCP client to the Tool Registry to dynamically discover and use tools from external enterprise MCP servers (e.g., GitHub, DBs). | ⬜ | — |',
              '| P28-T4 | `backend/mcp/` | **Model Context Protocol (MCP):** Add an MCP client to the Tool Registry to dynamically discover and use tools from external enterprise MCP servers (e.g., GitHub, DBs). | ⬜ | — |\n| P28-T5 | `tests/e2e/` | **E2E Tests:** Pytest tests for Agentic Tool Registry and Loop. | ⬜ | — |')
t = t.replace('| P29-T3 | `frontend/src/lib/resume-styles.css` | **Templates & Export:** 5-10 selectable CSS themes and a `window.print()` PDF export button. | ⬜ | — |',
              '| P29-T3 | `frontend/src/lib/resume-styles.css` | **Templates & Export:** 5-10 selectable CSS themes and a `window.print()` PDF export button. | ⬜ | — |\n| P29-T4 | `frontend/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright and Pytest tests for Resume Builder. | ⬜ | — |')
with open('prompts/tasks.md', 'w') as f:
    f.write(t)

# Update parallel-plan.md
with open('prompts/parallel-plan.md', 'r') as f:
    p = f.read()
p = p.replace('| **10** | P26-T1 · P26-T2 · P26-T3 | ✅ T1 & T2 parallel. T3 runs alone (touches app.py) | main up to date |',
              '| **10** | P26-T1 · P26-T2 · P26-T3 | ✅ T1 & T2 parallel. T3 runs alone (touches app.py) | main up to date |\n| **10b** | P26-T5 | ❌ Sequential | P26-T4 merged |')
p = p.replace('| **11c** | P27-T3 | ✅ Parallel with 11b | P27-T1 merged |',
              '| **11c** | P27-T3 | ✅ Parallel with 11b | P27-T1 merged |\n| **11d** | P27-T4 | ❌ Sequential | P27-T2 & P27-T3 merged |')
p = p.replace('| **12d** | P28-T4 | ❌ Sequential | P28-T3 merged |',
              '| **12d** | P28-T4 | ❌ Sequential | P28-T3 merged |\n| **12e** | P28-T5 | ❌ Sequential | P28-T4 merged |')
p = p.replace('| **13b** | P29-T3 | ❌ Sequential | P29-T2 merged |',
              '| **13b** | P29-T3 | ❌ Sequential | P29-T2 merged |\n| **13c** | P29-T4 | ❌ Sequential | P29-T3 merged |')
with open('prompts/parallel-plan.md', 'w') as f:
    f.write(p)

# Update EXECUTION_GUIDE.md
with open('prompts/tasks/EXECUTION_GUIDE.md', 'r') as f:
    e = f.read()
e = e.replace('python3 scripts/jules_submit.py --task P26-T3\n```',
              'python3 scripts/jules_submit.py --task P26-T3\n# After P26-T3 and P26-T4 merged:\npython3 scripts/jules_submit.py --task P26-T5\n```')
e = e.replace('| 🟡 3rd | `phase_27_voice_mode/P27_T3_voice_mode_ui` | Glowing orb UI | `frontend/src/lib/Assistant.svelte` only |',
              '| 🟡 3rd | `phase_27_voice_mode/P27_T3_voice_mode_ui` | Glowing orb UI | `frontend/src/lib/Assistant.svelte` only |\n| 🟢 4th | `phase_27_voice_mode/P27_T4_e2e_tests` | E2E Tests | `frontend/e2e/`, `tests/e2e/` |')
e = e.replace('python3 scripts/jules_submit.py --task P27-T3\n```',
              'python3 scripts/jules_submit.py --task P27-T3\n# After P27-T3 merged:\npython3 scripts/jules_submit.py --task P27-T4\n```')
e = e.replace('| 🔴 4th | `phase_28_agentic/P28_T4_mcp_support` | MCP Client Integration | `backend/mcp/client.py` (NEW), `backend/tools/registry.py`, `backend/config.py`, `frontend/src/lib/Settings.svelte` |',
              '| 🔴 4th | `phase_28_agentic/P28_T4_mcp_support` | MCP Client Integration | `backend/mcp/client.py` (NEW), `backend/tools/registry.py`, `backend/config.py`, `frontend/src/lib/Settings.svelte` |\n| 🟢 5th | `phase_28_agentic/P28_T5_e2e_tests` | E2E Tests | `tests/e2e/` |')
e = e.replace('python3 scripts/jules_submit.py --task P28-T4\n```',
              'python3 scripts/jules_submit.py --task P28-T4\n# After P28-T4 merged:\npython3 scripts/jules_submit.py --task P28-T5\n```')
e = e.replace('| 🟡 3rd | `phase_29_resume_builder/P29_T3_pdf_export` | 5-10 CSS Templates & PDF Export | `frontend/src/lib/resume-styles.css` (NEW), `frontend/src/lib/ResumeBuilder.svelte` |',
              '| 🟡 3rd | `phase_29_resume_builder/P29_T3_pdf_export` | 5-10 CSS Templates & PDF Export | `frontend/src/lib/resume-styles.css` (NEW), `frontend/src/lib/ResumeBuilder.svelte` |\n| 🟢 4th | `phase_29_resume_builder/P29_T4_e2e_tests` | E2E Tests | `frontend/e2e/`, `tests/e2e/` |')
e = e.replace('python3 scripts/jules_submit.py --task P29-T3\n```',
              'python3 scripts/jules_submit.py --task P29-T3\n# After P29-T3 merged:\npython3 scripts/jules_submit.py --task P29-T4\n```')
with open('prompts/tasks/EXECUTION_GUIDE.md', 'w') as f:
    f.write(e)
