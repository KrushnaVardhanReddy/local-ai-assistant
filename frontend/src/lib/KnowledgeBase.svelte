<script lang="ts">
  import { onMount } from "svelte";

  let ragEnabled = $state(false);
  let webSearchEnabled = $state(false);
  let documents = $state<string[]>([]);
  let isDragging = $state(false);

  type UploadItem = {
    id: string;
    file: File;
    progress: number;
    status: "uploading" | "success" | "error";
    errorMsg?: string;
  };

  let uploadQueue = $state<UploadItem[]>([]);

  const API_BASE = "http://127.0.0.1:8765";

  async function fetchStatus() {
    try {
      const res = await fetch(`${API_BASE}/rag/status`);
      if (res.ok) {
        const data = await res.json();
        ragEnabled = data.enabled;
      }
    } catch (e) {
      console.error("Failed to fetch RAG status", e);
    }

    try {
      const res = await fetch(`${API_BASE}/web_search/status`);
      if (res.ok) {
        const data = await res.json();
        webSearchEnabled = data.enabled;
      }
    } catch (e) {
      console.error("Failed to fetch web search status", e);
    }
  }

  async function toggleRag() {
    try {
      const res = await fetch(`${API_BASE}/rag/toggle`, { method: "POST" });
      if (res.ok) {
        const data = await res.json();
        ragEnabled = data.enabled;
      }
    } catch (e) {
      console.error("Failed to toggle RAG", e);
    }
  }

  async function toggleWebSearch() {
    try {
      const res = await fetch(`${API_BASE}/web_search/toggle`, { method: "POST" });
      if (res.ok) {
        const data = await res.json();
        webSearchEnabled = data.enabled;
      }
    } catch (e) {
      console.error("Failed to toggle web search", e);
    }
  }

  async function fetchDocuments() {
    try {
      const res = await fetch(`${API_BASE}/rag/documents`);
      if (res.ok) {
        const data = await res.json();
        documents = data.documents;
      }
    } catch (e) {
      console.error("Failed to fetch documents", e);
    }
  }

  async function deleteDocument(filename: string) {
    try {
      const res = await fetch(`${API_BASE}/rag/document/${encodeURIComponent(filename)}`, { method: "DELETE" });
      if (res.ok) {
        await fetchDocuments();
      }
    } catch (e) {
      console.error("Failed to delete document", e);
    }
  }

  onMount(() => {
    fetchStatus();
    fetchDocuments();
  });

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    isDragging = true;
  }

  function handleDragLeave(e: DragEvent) {
    e.preventDefault();
    isDragging = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragging = false;
    if (e.dataTransfer?.files) {
      handleFiles(Array.from(e.dataTransfer.files));
    }
  }

  function handleFileInput(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files) {
      handleFiles(Array.from(target.files));
    }
    // reset input so the same file can be selected again if needed
    target.value = "";
  }

  function handleFiles(files: File[]) {
    const allowed = [".pdf", ".txt", ".md"];
    for (const file of files) {
      const ext = file.name.slice(file.name.lastIndexOf(".")).toLowerCase();
      if (!allowed.includes(ext)) {
        alert(`File type not allowed: ${file.name}. Only .pdf, .txt, .md are accepted.`);
        continue;
      }

      const id = crypto.randomUUID();
      const item: UploadItem = {
        id,
        file,
        progress: 0,
        status: "uploading",
      };

      uploadQueue = [...uploadQueue, item];
      uploadFile(item);
    }
  }

  function updateQueueItem(id: string, updates: Partial<UploadItem>) {
    uploadQueue = uploadQueue.map(item => item.id === id ? { ...item, ...updates } : item);
  }

  function uploadFile(item: UploadItem) {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${API_BASE}/rag/upload`, true);

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) {
        const progress = Math.round((e.loaded / e.total) * 100);
        updateQueueItem(item.id, { progress });
      }
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        updateQueueItem(item.id, { status: "success", progress: 100 });
        fetchDocuments();
        setTimeout(() => {
          uploadQueue = uploadQueue.filter(i => i.id !== item.id);
        }, 2000);
      } else {
        updateQueueItem(item.id, { status: "error", errorMsg: "Upload failed" });
      }
    };

    xhr.onerror = () => {
      updateQueueItem(item.id, { status: "error", errorMsg: "Network error" });
    };

    const formData = new FormData();
    formData.append("file", item.file);

    xhr.send(formData);
  }
</script>

<div class="knowledge-base">
  <div class="header">
    <h2>Knowledge Base</h2>
    <label class="toggle">
      <input type="checkbox" checked={ragEnabled} onchange={toggleRag} />
      <span class="slider"></span>
      Use knowledge base for responses
    </label>
    <label class="toggle">
      <input type="checkbox" checked={webSearchEnabled} onchange={toggleWebSearch} />
      <span class="slider"></span>
      Use Internet Search for responses
    </label>
  </div>

  <div
    class="drop-zone {isDragging ? 'dragging' : ''}"
    ondragover={handleDragOver}
    ondragleave={handleDragLeave}
    ondrop={handleDrop}
    role="button"
    tabindex="0"
    onclick={() => document.getElementById('fileUpload')?.click()}
    onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') document.getElementById('fileUpload')?.click(); }}
  >
    <p>Drop PDF, TXT, or MD files here, or click to browse</p>
    <input
      id="fileUpload"
      type="file"
      multiple
      accept=".pdf,.txt,.md"
      onchange={handleFileInput}
      style="display: none;"
    />
  </div>

  {#if uploadQueue.length > 0}
    <div class="upload-queue">
      {#each uploadQueue as item (item.id)}
        <div class="upload-item">
          <div class="upload-info">
            <span class="filename">{item.file.name}</span>
            <span class="status-icon">
              {#if item.status === 'uploading'}
                ⏳
              {:else if item.status === 'success'}
                ✅
              {:else if item.status === 'error'}
                ❌
              {/if}
            </span>
          </div>
          <div class="progress-bar-container">
            <div class="progress-bar {item.status}" style="width: {item.progress}%"></div>
          </div>
          {#if item.errorMsg}
            <div class="error-msg">{item.errorMsg}</div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  <div class="documents-list">
    <h3>Ingested Documents</h3>
    {#if documents.length === 0}
      <p class="empty-state">No documents ingested yet. Upload your first file above.</p>
    {:else}
      <ul>
        {#each documents as doc}
          <li>
            <span class="doc-name">📄 {doc}</span>
            <button class="delete-btn" onclick={() => deleteDocument(doc)}>Delete</button>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>

<style>
  .knowledge-base {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    padding: 1.5rem;
    color: #e2e8f0;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    width: 100%;
  }

  .header {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .header h2 {
    margin: 0;
    font-size: 1.25rem;
    color: #e2e8f0;
  }

  .toggle {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    cursor: pointer;
    font-size: 0.95rem;
    color: #cbd5e1;
  }

  .toggle input {
    display: none;
  }

  .slider {
    position: relative;
    width: 36px;
    height: 20px;
    background-color: #475569;
    border-radius: 20px;
    transition: 0.3s;
  }

  .slider::before {
    content: "";
    position: absolute;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background-color: white;
    top: 2px;
    left: 2px;
    transition: 0.3s;
  }

  .toggle input:checked + .slider {
    background-color: #7c3aed;
  }

  .toggle input:checked + .slider::before {
    transform: translateX(16px);
  }

  .drop-zone {
    border: 2px dashed rgba(255, 255, 255, 0.2);
    border-radius: 8px;
    height: 120px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s ease;
    text-align: center;
    padding: 1rem;
  }

  .drop-zone:hover, .drop-zone.dragging {
    border-color: #7c3aed;
    background: rgba(124, 58, 237, 0.05);
    box-shadow: 0 0 12px rgba(124, 58, 237, 0.3);
  }

  .drop-zone p {
    color: #94a3b8;
    margin: 0;
    pointer-events: none;
  }

  .upload-queue {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .upload-item {
    background: rgba(0, 0, 0, 0.2);
    padding: 0.75rem;
    border-radius: 6px;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .upload-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.9rem;
  }

  .filename {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .progress-bar-container {
    height: 6px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-bar {
    height: 100%;
    background: #7c3aed;
    transition: width 0.2s ease;
  }

  .progress-bar.success {
    background: #22c55e;
  }

  .progress-bar.error {
    background: #ef4444;
  }

  .error-msg {
    color: #ef4444;
    font-size: 0.8rem;
  }

  .documents-list h3 {
    margin: 0 0 1rem 0;
    font-size: 1.1rem;
    color: #e2e8f0;
  }

  .documents-list ul {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .documents-list li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: rgba(255, 255, 255, 0.02);
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .doc-name {
    font-size: 0.95rem;
  }

  .delete-btn {
    background: rgba(239, 68, 68, 0.1);
    color: #ef4444;
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: 4px;
    padding: 0.25rem 0.5rem;
    font-size: 0.8rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .delete-btn:hover {
    background: rgba(239, 68, 68, 0.2);
  }

  .empty-state {
    color: #64748b;
    font-style: italic;
    font-size: 0.95rem;
    margin: 0;
  }
</style>
