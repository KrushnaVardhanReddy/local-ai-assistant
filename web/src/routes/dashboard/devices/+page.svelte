<script lang="ts">
  import { supabase } from '$lib/supabase';
  import { onMount } from 'svelte';

  let devices: any[] = [];
  let loading = true;
  let planLimit = 1;

  onMount(async () => {
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) return;

    // Fetch devices
    const { data } = await supabase
      .from('user_devices')
      .select('*')
      .eq('user_id', user.id)
      .order('last_seen_at', { ascending: false });
    devices = data || [];

    // Fetch plan limit
    const { data: profile } = await supabase
      .from('profiles')
      .select('plan')
      .eq('id', user.id)
      .single();
    const plan = profile?.plan || 'demo';
    const limits: Record<string, number> = { demo: 1, payg: 1, monthly: 2, founding: 3 };
    planLimit = limits[plan] || 1;

    loading = false;
  });

  async function removeDevice(deviceId: string) {
    await supabase.from('user_devices').delete().eq('id', deviceId);
    devices = devices.filter(d => d.id !== deviceId);
  }

  function relativeTime(dateStr: string) {
    const diff = Date.now() - new Date(dateStr).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 60) return `${mins}m ago`;
    const hrs = Math.floor(mins / 60);
    if (hrs < 24) return `${hrs}h ago`;
    return `${Math.floor(hrs / 24)}d ago`;
  }
</script>

<svelte:head><title>Registered Devices — Parakeet</title></svelte:head>

<div class="devices-page">
  <h1>Registered Devices</h1>
  <p class="subtitle">{devices.length} / {planLimit} devices registered</p>

  {#if loading}
    <div class="loading">Loading...</div>
  {:else if devices.length === 0}
    <div class="empty">No devices registered yet. Open the app to register this device.</div>
  {:else}
    <div class="devices-list">
      {#each devices as device}
        <div class="device-card">
          <div class="device-info">
            <span class="device-icon">💻</span>
            <div>
              <div class="device-name">{device.device_label}</div>
              <div class="device-meta">Last seen: {relativeTime(device.last_seen_at)} · ID: {device.machine_id.slice(0, 8)}...</div>
            </div>
          </div>
          <button class="remove-btn" on:click={() => removeDevice(device.id)}>Remove</button>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .devices-page { max-width: 640px; margin: 0 auto; padding: 40px 24px; }
  h1 { font-size: 24px; font-weight: 700; color: #e2e8f0; margin: 0 0 8px; }
  .subtitle { color: rgba(255,255,255,0.4); font-size: 14px; margin: 0 0 32px; }
  .device-card {
    display: flex; align-items: center; justify-content: space-between;
    background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.08);
    border-radius: 12px; padding: 16px 20px; margin-bottom: 12px;
  }
  .device-info { display: flex; align-items: center; gap: 14px; }
  .device-icon { font-size: 24px; }
  .device-name { font-size: 15px; font-weight: 600; color: #e2e8f0; }
  .device-meta { font-size: 12px; color: rgba(255,255,255,0.4); margin-top: 2px; }
  .remove-btn {
    background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3);
    color: #f87171; border-radius: 8px; padding: 6px 16px; cursor: pointer; font-size: 13px;
  }
  .remove-btn:hover { background: rgba(239, 68, 68, 0.2); }
</style>