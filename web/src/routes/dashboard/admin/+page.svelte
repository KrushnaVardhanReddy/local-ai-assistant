<script lang="ts">
    let { data } = $props();
    let inviteModalOpen = $state(false);
    let inviteEmail = $state('');
    let inviting = $state(false);
    let errorMsg = $state('');

    let members = $state(data.members || []);
    let org = $derived(data.org || {});

    $effect(() => {
        members = data.members || [];
    });

    // Filter out revoked from active count
    let activeMembers = $derived(members.filter((m: any) => m.plan !== 'revoked'));
    let seatsUsed = $derived(activeMembers.length);
    let seatsPurchased = $derived(org.seats_purchased || 10);
    let progressPercent = $derived(Math.min((seatsUsed / seatsPurchased) * 100, 100));

    async function inviteUser() {
        if (!inviteEmail) return;
        inviting = true;
        errorMsg = '';

        try {
            const res = await fetch('/api/enterprise/invite', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email: inviteEmail })
            });
            const result = await res.json();

            if (!res.ok) {
                errorMsg = result.error || 'Failed to send invite';
            } else {
                inviteModalOpen = false;
                inviteEmail = '';
                // Optimistically add to list
                members = [...members, { email: inviteEmail, plan: 'enterprise', active_session_at: null, id: 'temp-' + Date.now() }];
            }
        } catch (e: any) {
            errorMsg = e.message;
        } finally {
            inviting = false;
        }
    }

    async function revokeUser(userId: string) {
        if (!confirm('Are you sure you want to revoke this user? They will instantly lose access.')) return;

        try {
            const res = await fetch('/api/enterprise/revoke', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ userId })
            });

            if (res.ok) {
                members = members.map((m: any) => m.id === userId ? { ...m, plan: 'revoked' } : m);
            } else {
                const result = await res.json();
                alert(result.error || 'Failed to revoke user');
            }
        } catch (e: any) {
            alert(e.message);
        }
    }

    function timeAgo(dateString: string | null) {
        if (!dateString) return '—';
        const date = new Date(dateString);
        const diff = Date.now() - date.getTime();
        const minutes = Math.floor(diff / 60000);
        if (minutes < 60) return `${minutes}m ago`;
        const hours = Math.floor(minutes / 60);
        if (hours < 24) return `${hours}h ago`;
        return `${Math.floor(hours / 24)}d ago`;
    }

    function closeInviteModal() {
        inviteModalOpen = false;
    }
</script>

<svelte:head>
  <title>Admin Dashboard - BarnOwl</title>
</svelte:head>

<main class="content">
  <header class="header">
    <h1>🏢 {org.domain || 'Organization'} — BarnOwl Enterprise</h1>
  </header>

  <div class="summary-card">
    <div class="summary-header">
      <h2>Seats: <span class="seat-count">{seatsUsed} / {seatsPurchased} used</span></h2>
      <button class="btn-primary" onclick={() => inviteModalOpen = true}>+ Invite</button>
    </div>

    <div class="progress-bar-container">
      <div class="progress-bar" style="width: {progressPercent}%"></div>
    </div>
  </div>

  <div class="table-container">
    <table class="members-table">
      <thead>
        <tr>
          <th>Email</th>
          <th>Last Active</th>
          <th>Status</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        {#each members as member (member.id)}
          <tr>
            <td>{member.email}</td>
            <td>{timeAgo(member.active_session_at)}</td>
            <td>
              {#if member.plan === 'revoked'}
                <span class="status-badge revoked">● Revoked</span>
              {:else if member.plan === 'enterprise'}
                <span class="status-badge active">● Active</span>
              {:else}
                <span class="status-badge pending">● {member.plan}</span>
              {/if}
            </td>
            <td>
              {#if member.plan !== 'revoked' && member.id !== data.user.id}
                <button class="btn-revoke" onclick={() => revokeUser(member.id)}>Revoke</button>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</main>

{#if inviteModalOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="modal-backdrop" onclick={closeInviteModal}>
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <h2>Invite Team Member</h2>
      <p>Send an invitation email to a colleague to join your organization.</p>

      {#if errorMsg}
        <div class="error">{errorMsg}</div>
      {/if}

      <div class="form-group">
        <label for="inviteEmail">Email address</label>
        <input type="email" id="inviteEmail" bind:value={inviteEmail} placeholder="colleague@yourcompany.com" />
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" onclick={closeInviteModal}>Cancel</button>
        <button class="btn-primary" onclick={inviteUser} disabled={inviting || !inviteEmail}>
          {inviting ? 'Sending...' : 'Send Invite'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .content {
    flex-grow: 1;
    padding: 2rem 4rem;
    max-width: 1000px;
    margin: 0 auto;
    width: 100%;
  }

  .header {
    margin-bottom: 2rem;
  }

  .header h1 {
    font-size: 1.875rem;
    margin: 0;
    color: var(--text);
  }

  .summary-card {
    background-color: var(--surface);
    border: var(--border);
    border-radius: 8px;
    padding: 2rem;
    margin-bottom: 2rem;
  }

  .summary-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .summary-header h2 {
    margin: 0;
    font-size: 1.25rem;
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .seat-count {
      color: var(--muted);
      font-weight: normal;
  }

  .progress-bar-container {
    height: 8px;
    background-color: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
    overflow: hidden;
  }

  .progress-bar {
    height: 100%;
    background-color: var(--accent);
    transition: width 0.3s ease;
  }

  .table-container {
    background-color: var(--surface);
    border: var(--border);
    border-radius: 8px;
    overflow: hidden;
  }

  .members-table {
    width: 100%;
    border-collapse: collapse;
    text-align: left;
  }

  .members-table th,
  .members-table td {
    padding: 1rem 1.5rem;
    border-bottom: var(--border);
  }

  .members-table th {
    background-color: rgba(255, 255, 255, 0.02);
    font-weight: 600;
    color: var(--text);
  }

  .members-table td {
    color: var(--muted);
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    padding: 0.25rem 0.5rem;
    border-radius: 99px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .status-badge.active {
    background-color: rgba(16, 185, 129, 0.1);
    color: #10b981;
  }

  .status-badge.pending {
    background-color: rgba(245, 158, 11, 0.1);
    color: #f59e0b;
  }

  .status-badge.revoked {
    background-color: rgba(239, 68, 68, 0.1);
    color: #ef4444;
  }

  .btn-primary {
    background-color: var(--accent);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.2s;
  }

  .btn-primary:hover:not(:disabled) {
    opacity: 0.9;
  }

  .btn-primary:disabled {
      opacity: 0.5;
      cursor: not-allowed;
  }

  .btn-secondary {
    background-color: transparent;
    color: var(--text);
    border: 1px solid rgba(255, 255, 255, 0.2);
    padding: 0.5rem 1rem;
    border-radius: 4px;
    font-weight: 600;
    cursor: pointer;
  }

  .btn-secondary:hover {
    background-color: rgba(255, 255, 255, 0.05);
  }

  .btn-revoke {
    background-color: transparent;
    color: #ef4444;
    border: 1px solid rgba(239, 68, 68, 0.5);
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-revoke:hover {
    background-color: rgba(239, 68, 68, 0.1);
  }

  /* Modal Styles */
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background-color: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
  }

  .modal {
    background-color: var(--surface);
    border: var(--border);
    border-radius: 8px;
    padding: 2rem;
    width: 100%;
    max-width: 400px;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  }

  .modal h2 {
    margin-top: 0;
    margin-bottom: 0.5rem;
  }

  .modal p {
    color: var(--muted);
    margin-bottom: 1.5rem;
    font-size: 0.875rem;
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.5rem;
    color: var(--text);
    font-size: 0.875rem;
    font-weight: 500;
  }

  .form-group input {
    width: 100%;
    padding: 0.75rem;
    background-color: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 4px;
    color: var(--text);
    box-sizing: border-box;
  }

  .form-group input:focus {
      outline: none;
      border-color: var(--accent);
  }

  .error {
    background-color: rgba(239, 68, 68, 0.1);
    color: #f87171;
    padding: 0.75rem;
    border-radius: 4px;
    margin-bottom: 1rem;
    font-size: 0.875rem;
  }

  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
  }
</style>
