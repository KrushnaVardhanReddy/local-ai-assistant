import re

with open("frontend/src/lib/Settings.svelte", "r") as f:
    content = f.read()

# Add state variable
state_pattern = r'let resumeStatus = \$state\(""\);\n  let resumeFilename = \$state\(""\);'
state_replacement = r'''let resumeStatus = $state("");
  let resumeFilename = $state("");

  // System Status state
  let systemStatus = $state<any>(null);'''

content = re.sub(state_pattern, state_replacement, content)

# Add onMount fetch logic
onmount_pattern = r'currentLLMProvider = healthData.llm_provider;\n      }'
onmount_replacement = r'''currentLLMProvider = healthData.llm_provider;
      }

      const statusRes = await fetch(`${apiUrl}/api/status`);
      if (statusRes.ok) {
        systemStatus = await statusRes.json();
      }'''

content = content.replace(
    "currentLLMProvider = healthData.llm_provider;\n      }",
    "currentLLMProvider = healthData.llm_provider;\n      }\n      const statusRes = await fetch(`${apiUrl}/api/status`);\n      if (statusRes.ok) {\n        systemStatus = await statusRes.json();\n      }"
)

# Add HTML section
html_pattern = r'<h2>Settings<\/h2>\n  <\/div>'
html_replacement = r'''<h2>Settings</h2>
  </div>

  <div class="config-section">
    <div class="section-label">System Status</div>
    {#if systemStatus}
      <div class="status-grid">
        <div class="status-item">
          <span class="status-key">LLM Provider</span>
          <span class="status-value badge">{systemStatus.llm_provider}</span>
        </div>
        <div class="status-item">
          <span class="status-key">LLM Model</span>
          <span class="status-value badge">{systemStatus.llm_model}</span>
        </div>
        <div class="status-item">
          <span class="status-key">STT Provider</span>
          <span class="status-value badge">{systemStatus.stt_provider}</span>
        </div>
        <div class="status-item">
          <span class="status-key">STT Model</span>
          <span class="status-value badge">{systemStatus.stt_model}</span>
        </div>
        <div class="status-item">
          <span class="status-key">Local STT Engine</span>
          <span class="status-value badge">{systemStatus.local_stt_engine}</span>
        </div>
      </div>
    {:else}
      <div class="status-indicator">Loading status...</div>
    {/if}
  </div>

  <hr class="divider" />'''

content = content.replace("<h2>Settings</h2>\n  </div>", html_replacement)

# Add styles
style_pattern = r'\.status-indicator \{'
style_replacement = r'''.status-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 0.5rem;
    background: rgba(0, 0, 0, 0.2);
    padding: 1rem;
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .status-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.85rem;
  }

  .status-key {
    color: #aaa;
  }

  .status-value {
    color: #fff;
    font-weight: 500;
  }

  .status-indicator {'''

content = content.replace(".status-indicator {", style_replacement)


with open("frontend/src/lib/Settings.svelte", "w") as f:
    f.write(content)
