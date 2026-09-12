<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { Command } from "@tauri-apps/plugin-shell";

  let terminalOutput: string[] = $state([]);
  let isRunning = $state(false);
  let processChild: any = null;
  let terminalRef: HTMLDivElement;

  $effect(() => {
    if (terminalRef && terminalOutput.length) {
      terminalRef.scrollTop = terminalRef.scrollHeight;
    }
  });

  async function startBackend() {
    if (isRunning) return;

    if (!(window as any).__TAURI_INTERNALS__) {
      terminalOutput = [...terminalOutput, "Running in browser mode: Tauri IPC unavailable. Backend spawn skipped."];
      return;
    }

    try {
      let cmd;
      if (import.meta.env.DEV) {
        cmd = Command.create("python", ["backend/app.py"]);
      } else {
        cmd = Command.sidecar("backend");
      }

      cmd.on('close', data => {
        terminalOutput = [...terminalOutput, `Process exited with code ${data.code}`];
        isRunning = false;
        processChild = null;
      });

      cmd.on('error', error => {
        terminalOutput = [...terminalOutput, `Error: "${error}"`];
        isRunning = false;
        processChild = null;
      });

      cmd.stdout.on('data', line => {
        terminalOutput = [...terminalOutput, line];
      });

      cmd.stderr.on('data', line => {
        terminalOutput = [...terminalOutput, `[STDERR] ${line}`];
      });

      processChild = await cmd.spawn();
      isRunning = true;
      terminalOutput = [...terminalOutput, "Backend started successfully..."];
    } catch (e: any) {
      terminalOutput = [...terminalOutput, `Failed to start backend: ${e.message || String(e)}`];
    }
  }

  async function stopBackend() {
    if (!isRunning || !processChild) return;

    if (!(window as any).__TAURI_INTERNALS__) {
      return;
    }

    try {
      await processChild.kill();
      terminalOutput = [...terminalOutput, "Kill signal sent to backend."];
    } catch (e: any) {
      terminalOutput = [...terminalOutput, `Failed to kill backend: ${e.message || String(e)}`];
    }
  }

  onDestroy(async () => {
    await stopBackend();
  });
</script>

<div class="stealth-terminal-container">
  <div class="controls">
    <button class="btn-primary" onclick={startBackend} disabled={isRunning} data-testid="start-backend-btn">
      Start Backend
    </button>
    <button class="btn-secondary" onclick={stopBackend} disabled={!isRunning} data-testid="stop-backend-btn">
      Stop Backend
    </button>
  </div>

  <div class="terminal-output" bind:this={terminalRef} data-testid="stealth-terminal-output">
    {#each terminalOutput as line}
      <div class="line">{line}</div>
    {/each}
  </div>
</div>

<style>
  .stealth-terminal-container {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    width: 100%;
  }

  .controls {
    display: flex;
    gap: 0.5rem;
  }

  .btn-primary, .btn-secondary {
    padding: 0.5rem 1rem;
    border-radius: 4px;
    font-size: 0.85rem;
    cursor: pointer;
    border: none;
    transition: background 0.2s;
    font-weight: 600;
  }

  .btn-primary {
    background: #007bff;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: #0056b3;
  }

  .btn-secondary {
    background: #444;
    color: #fff;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #555;
  }

  .btn-primary:disabled, .btn-secondary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .terminal-output {
    background-color: #000;
    color: #0f0;
    font-family: 'Courier New', Courier, monospace;
    font-size: 0.8rem;
    padding: 0.75rem;
    border-radius: 6px;
    height: 150px;
    overflow-y: auto;
    border: 1px solid #333;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .line {
    margin-bottom: 2px;
  }
</style>
