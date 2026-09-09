<script lang="ts">
	import { goto } from '$app/navigation';
	import { supabase } from '$lib/supabase';

	let activeTab = $state<'signin' | 'signup'>('signin');
	let email = $state('');
	let password = $state('');
	let errorMsg = $state('');
	let loading = $state(false);

	let showSso = $state(false);
	let ssoEmail = $state('');
	let ssoLoading = $state(false);
	let ssoError = $state('');

	async function handleSubmit(event: Event) {
		event.preventDefault();
		errorMsg = '';
		loading = true;

		try {
			if (activeTab === 'signin') {
				const { error } = await supabase.auth.signInWithPassword({ email, password });
				if (error) throw error;
				goto('/dashboard');
			} else {
				const { error } = await supabase.auth.signUp({ email, password });
				if (error) throw error;
				// If email confirmations are enabled, they might need to check email.
				// If not, they are signed in immediately.
				goto('/dashboard');
			}
		} catch (error: any) {
			errorMsg = error.message;
		} finally {
			loading = false;
		}
	}

	async function handleGoogleOAuth() {
		errorMsg = '';
		loading = true;
		try {
			const { error } = await supabase.auth.signInWithOAuth({
				provider: 'google',
				options: {
					redirectTo: `${window.location.origin}/auth/callback`
				}
			});
			if (error) throw error;
		} catch (error: any) {
			errorMsg = error.message;
			loading = false;
		}
	}

	async function handleSsoSubmit(event: Event) {
		event.preventDefault();
		ssoError = '';
		ssoLoading = true;

		try {
			if (!ssoEmail || !ssoEmail.includes('@')) {
				throw new Error('Please enter a valid work email.');
			}
			const domain = ssoEmail.split('@')[1];

			const res = await fetch('/api/enterprise/sso-init', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ domain })
			});

			const data = await res.json();
			if (!res.ok) {
				throw new Error(data.error || 'Failed to initialize SSO.');
			}

			if (data.url) {
				window.location.href = data.url;
			} else {
				throw new Error('Invalid SSO response.');
			}
		} catch (error: any) {
			ssoError = error.message;
		} finally {
			ssoLoading = false;
		}
	}
</script>

<div class="login-container">
	<div class="login-box">
		<h2>Welcome to Local AI Assistant</h2>

		<div class="tabs">
			<button
				class="tab"
				class:active={activeTab === 'signin'}
				onclick={() => {
					activeTab = 'signin';
					errorMsg = '';
				}}
			>
				Sign In
			</button>
			<button
				class="tab"
				class:active={activeTab === 'signup'}
				onclick={() => {
					activeTab = 'signup';
					errorMsg = '';
				}}
			>
				Sign Up
			</button>
		</div>

		<form onsubmit={handleSubmit} class="form">
			<div class="form-group">
				<label for="email">Email</label>
				<input type="email" id="email" bind:value={email} required />
			</div>

			<div class="form-group">
				<label for="password">Password</label>
				<input type="password" id="password" bind:value={password} required />
			</div>

			<button type="submit" class="btn-primary" disabled={loading}>
				{loading
					? 'Please wait...'
					: activeTab === 'signin'
						? 'Sign In'
						: 'Sign Up'}
			</button>
		</form>

		<div class="divider">
			<span>OR</span>
		</div>

		<button onclick={handleGoogleOAuth} class="btn-google" disabled={loading}>
			Continue with Google
		</button>

		{#if errorMsg}
			<div class="error">{errorMsg}</div>
		{/if}

		<div class="divider">
			<span>Enterprise</span>
		</div>

		{#if !showSso}
			<button class="btn-google" onclick={() => showSso = true} disabled={loading}>
				Sign in with SSO
			</button>
		{:else}
			<form onsubmit={handleSsoSubmit} class="form sso-form">
				<div class="form-group">
					<label for="sso-email">Work email</label>
					<input type="email" id="sso-email" bind:value={ssoEmail} placeholder="you@acme.com" required />
				</div>
				<button type="submit" class="btn-primary" disabled={ssoLoading}>
					{ssoLoading ? 'Redirecting...' : 'Continue with SSO'}
				</button>
				{#if ssoError}
					<div class="error">{ssoError}</div>
				{/if}
			</form>
		{/if}
	</div>
</div>

<style>
	.login-container {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 100vh;
		background-color: #f9fafb;
		padding: 1rem;
	}

	.login-box {
		background: white;
		padding: 2rem;
		border-radius: 8px;
		box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
		width: 100%;
		max-width: 400px;
	}

	h2 {
		text-align: center;
		color: #111827;
		margin-top: 0;
		margin-bottom: 1.5rem;
	}

	.tabs {
		display: flex;
		margin-bottom: 1.5rem;
		border-bottom: 2px solid #e5e7eb;
	}

	.tab {
		flex: 1;
		background: none;
		border: none;
		padding: 0.75rem;
		font-weight: 500;
		color: #6b7280;
		cursor: pointer;
		position: relative;
		font-size: 1rem;
	}

	.tab:hover {
		color: #374151;
	}

	.tab.active {
		color: #2563eb;
	}

	.tab.active::after {
		content: '';
		position: absolute;
		bottom: -2px;
		left: 0;
		right: 0;
		height: 2px;
		background-color: #2563eb;
	}

	.form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	input {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 1rem;
	}

	input:focus {
		outline: none;
		border-color: #2563eb;
		box-shadow: 0 0 0 1px #2563eb;
	}

	.btn-primary {
		background-color: #2563eb;
		color: white;
		padding: 0.75rem;
		border: none;
		border-radius: 4px;
		font-weight: 500;
		font-size: 1rem;
		cursor: pointer;
		margin-top: 0.5rem;
	}

	.btn-primary:hover {
		background-color: #1d4ed8;
	}

	.btn-primary:disabled {
		background-color: #93c5fd;
		cursor: not-allowed;
	}

	.divider {
		display: flex;
		align-items: center;
		text-align: center;
		margin: 1.5rem 0;
		color: #6b7280;
		font-size: 0.875rem;
	}

	.divider::before,
	.divider::after {
		content: '';
		flex: 1;
		border-bottom: 1px solid #e5e7eb;
	}

	.divider span {
		padding: 0 0.5rem;
	}

	.btn-google {
		background-color: white;
		color: #374151;
		padding: 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-weight: 500;
		font-size: 1rem;
		cursor: pointer;
		width: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-google:hover {
		background-color: #f9fafb;
	}

	.btn-google:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}

	.error {
		margin-top: 1rem;
		padding: 0.75rem;
		background-color: #fee2e2;
		color: #b91c1c;
		border-radius: 4px;
		font-size: 0.875rem;
		text-align: center;
	}

	.sso-form {
		margin-top: 1rem;
		padding: 1rem;
		background-color: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 4px;
	}
</style>
