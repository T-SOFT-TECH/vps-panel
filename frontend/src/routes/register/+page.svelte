<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { authStore } from '$lib/stores/auth.svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Card from '$lib/components/Card.svelte';
	import Alert from '$lib/components/Alert.svelte';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let loading = $state(false);
	let error = $state('');

	async function handleRegister(e: Event) {
		e.preventDefault();

		if (!name || !email || !password || !confirmPassword) {
			error = 'Please fill in all fields';
			return;
		}

		if (password.length < 8) {
			error = 'Password must be at least 8 characters';
			return;
		}

		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}

		loading = true;
		error = '';

		try {
			await authStore.register(email, password, name);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Registration failed';
		} finally {
			loading = false;
		}
	}

	function handleKeyPress(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			handleRegister(e);
		}
	}
</script>

<svelte:head>
	<title>Register - VPS Panel</title>
</svelte:head>

<div class="min-h-screen bg-zinc-950 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-md w-full space-y-8">
		<div class="text-center" in:fly={{ y: -20, duration: 500 }}>
			<h1 class="text-4xl font-bold text-green-500 mb-2">VPS Panel</h1>
			<h2 class="text-2xl font-semibold text-zinc-100">Create your account</h2>
			<p class="mt-2 text-sm text-zinc-400">
				Already have an account?
				<a href="/login" class="font-medium text-green-500 hover:text-green-400">
					Sign in here
				</a>
			</p>
		</div>

		<div in:fly={{ y: 20, duration: 500, delay: 100 }}>
			<Card>
			<form onsubmit={handleRegister} class="space-y-6">
				{#if error}
					<Alert variant="error" dismissible ondismiss={() => error = ''}>
						{error}
					</Alert>
				{/if}

				<Input
					label="Full name"
					type="text"
					bind:value={name}
					placeholder="John Doe"
					required
					disabled={loading}
					onkeypress={handleKeyPress}
				/>

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
					placeholder="At least 8 characters"
					required
					disabled={loading}
					onkeypress={handleKeyPress}
				/>

				<Input
					label="Confirm password"
					type="password"
					bind:value={confirmPassword}
					placeholder="Re-enter your password"
					required
					disabled={loading}
					onkeypress={handleKeyPress}
				/>

				<Button type="submit" {loading} disabled={loading} class="w-full">
					{loading ? 'Creating account...' : 'Create account'}
				</Button>

				<p class="text-xs text-zinc-400 text-center">
					By creating an account, you agree to our
					<a href="/terms" class="text-green-500 hover:text-green-400">Terms of Service</a>
					and
					<a href="/privacy" class="text-green-500 hover:text-green-400">Privacy Policy</a>
				</p>
			</form>
			</Card>
		</div>
	</div>
</div>
