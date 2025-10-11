<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

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
</script>

<svelte:head>
	<title>Login - VPS Panel</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 transition-colors duration-300" style="background-color: rgb(var(--bg-primary));">
	<!-- Theme Toggle - Top Right -->
	<div class="absolute top-6 right-6 z-20">
		<ThemeToggle />
	</div>

	<!-- Login Card -->
	<div class="w-full max-w-md space-y-8">
		<!-- Logo & Title -->
		<div class="text-center">
			<img
				src="/img/My Logo.png"
				alt="TSOFT Technologies"
				class="mx-auto h-20 mb-6"
			/>
			<h1 class="text-3xl font-bold mb-2" style="color: rgb(var(--text-primary));">
				Welcome Back
			</h1>
			<p class="text-base" style="color: rgb(var(--text-secondary));">
				Sign in to your VPS Panel
			</p>
		</div>

		<!-- Login Form Card -->
		<div class="elevated-card p-8">
			<form onsubmit={handleLogin} class="space-y-6">
				{#if error}
					<Alert variant="error" dismissible ondismiss={() => error = ''}>
						{error}
					</Alert>
				{/if}

				<div>
					<label
						for="email"
						class="block text-sm font-medium mb-2"
						style="color: rgb(var(--text-primary));"
					>
						Email Address
					</label>
					<Input
						id="email"
						type="email"
						bind:value={email}
						placeholder="you@example.com"
						required
						disabled={loading}
						class="modern-input"
					/>
				</div>

				<div>
					<label
						for="password"
						class="block text-sm font-medium mb-2"
						style="color: rgb(var(--text-primary));"
					>
						Password
					</label>
					<Input
						id="password"
						type="password"
						bind:value={password}
						placeholder="Enter your password"
						required
						disabled={loading}
						class="modern-input"
					/>
				</div>

				<div class="flex items-center justify-between text-sm">
					<label class="flex items-center cursor-pointer">
						<input
							type="checkbox"
							class="w-4 h-4 rounded border-2 text-primary-800 focus:ring-primary-600 focus:ring-offset-2 transition-all"
							style="border-color: rgb(var(--border-primary));"
						/>
						<span class="ml-2" style="color: rgb(var(--text-secondary));">
							Remember me
						</span>
					</label>
					<a href="/forgot-password" class="font-medium text-primary-800 hover:text-primary-700 transition-colors">
						Forgot password?
					</a>
				</div>

				<Button type="submit" {loading} disabled={loading} class="btn-primary w-full">
					{#if loading}
						<span class="spinner mr-2"></span>
						Signing in...
					{:else}
						Sign in
					{/if}
				</Button>
			</form>

			<!-- Register Link -->
			<div class="mt-6 pt-6 border-t" style="border-color: rgb(var(--border-primary));">
				<p class="text-center text-sm" style="color: rgb(var(--text-secondary));">
					Don't have an account?
					<a href="/register" class="font-semibold text-primary-800 hover:text-primary-700 transition-colors ml-1">
						Create one now →
					</a>
				</p>
			</div>
		</div>

		<!-- Security Badge -->
		<div class="text-center">
			<p class="text-xs flex items-center justify-center gap-2" style="color: rgb(var(--text-tertiary));">
				<svg class="w-4 h-4 text-primary-800" fill="currentColor" viewBox="0 0 20 20">
					<path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
				</svg>
				Secured with enterprise-grade encryption
			</p>
		</div>
	</div>
</div>
