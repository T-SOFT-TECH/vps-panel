<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { authStore } from '$lib/stores/auth.svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Card from '$lib/components/Card.svelte';
	import Alert from '$lib/components/Alert.svelte';

	let email = $state('');
	let password = $state('');
	let loading = $state(false);
	let error = $state('');

	async function handleLogin(e: Event) {
		e.preventDefault();

		if (!email || !password) {
			error = 'Please fill in all fields';
			return;
		}

		loading = true;
		error = '';

		try {
			await authStore.login(email, password);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login failed';
		} finally {
			loading = false;
		}
	}

	function handleKeyPress(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			handleLogin(e);
		}
	}
</script>

<svelte:head>
	<title>Login - VPS Panel</title>
</svelte:head>

<div class="min-h-screen bg-zinc-950 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-md w-full space-y-8">
		<div class="text-center" in:fly={{ y: -20, duration: 500 }}>
			<h1 class="text-4xl font-bold text-green-500 mb-2">VPS Panel</h1>
			<h2 class="text-2xl font-semibold text-zinc-100">Sign in to your account</h2>
			<p class="mt-2 text-sm text-zinc-400">
				Don't have an account?
				<a href="/register" class="font-medium text-green-500 hover:text-green-400">
					Register here
				</a>
			</p>
		</div>

		<div in:fly={{ y: 20, duration: 500, delay: 100 }}>
			<Card>
			<form onsubmit={handleLogin} class="space-y-6">
				{#if error}
					<Alert variant="error" dismissible ondismiss={() => error = ''}>
						{error}
					</Alert>
				{/if}

				<Input
					label="Email address"
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					required
					disabled={loading}
					onkeypress={handleKeyPress}
				/>

				<Input
					label="Password"
					type="password"
					bind:value={password}
					placeholder="Enter your password"
					required
					disabled={loading}
					onkeypress={handleKeyPress}
				/>

				<div class="flex items-center justify-between">
					<div class="text-sm">
						<a href="/forgot-password" class="font-medium text-green-500 hover:text-green-400">
							Forgot your password?
						</a>
					</div>
				</div>

				<Button type="submit" {loading} disabled={loading} class="w-full">
					{loading ? 'Signing in...' : 'Sign in'}
				</Button>
			</form>
			</Card>
		</div>
	</div>
</div>
