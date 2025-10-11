<script lang="ts">
	import { fade, fly, scale } from 'svelte/transition';
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

<div class="min-h-screen relative overflow-hidden flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8
            bg-gradient-to-br from-zinc-950 via-zinc-900 to-zinc-950
            dark:from-zinc-950 dark:via-zinc-900 dark:to-zinc-950
            light:from-zinc-50 light:via-white light:to-zinc-100">

	<!-- Animated Background Elements -->
	<div class="absolute inset-0 overflow-hidden">
		<!-- Large gradient orbs -->
		<div class="absolute top-0 -left-4 w-96 h-96 bg-primary-500/10 rounded-full mix-blend-multiply filter blur-3xl opacity-70 animate-blob"></div>
		<div class="absolute top-0 -right-4 w-96 h-96 bg-emerald-500/10 rounded-full mix-blend-multiply filter blur-3xl opacity-70 animate-blob animation-delay-2000"></div>
		<div class="absolute -bottom-8 left-20 w-96 h-96 bg-primary-600/10 rounded-full mix-blend-multiply filter blur-3xl opacity-70 animate-blob animation-delay-4000"></div>
	</div>

	<!-- Theme Toggle - Top Right -->
	<div class="absolute top-6 right-6 z-20" in:scale={{ delay: 300 }}>
		<ThemeToggle />
	</div>

	<!-- Login Card -->
	<div class="relative z-10 max-w-md w-full space-y-8">
		<!-- Logo & Title -->
		<div class="text-center" in:fly={{ y: -30, duration: 600 }}>
			<!-- Logo with gradient -->
			<div class="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 shadow-2xl mb-6 transform hover:scale-110 transition-transform duration-200">
				<svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
				</svg>
			</div>

			<h1 class="text-5xl font-bold bg-gradient-to-r from-primary-400 via-primary-500 to-emerald-400 bg-clip-text text-transparent mb-3">
				VPS Panel
			</h1>
			<p class="text-lg text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
				Welcome back! Sign in to continue
			</p>
		</div>

		<!-- Login Form Card -->
		<div in:fly={{ y: 30, duration: 600, delay: 200 }}>
			<div class="modern-card p-8">
				<form onsubmit={handleLogin} class="space-y-6">
					{#if error}
						<div in:scale={{ duration: 200 }}>
							<Alert variant="error" dismissible ondismiss={() => error = ''}>
								{error}
							</Alert>
						</div>
					{/if}

					<div class="space-y-5">
						<div>
							<label for="email" class="block text-sm font-medium text-zinc-300 dark:text-zinc-300 light:text-zinc-700 mb-2">
								Email Address
							</label>
							<Input
								id="email"
								type="email"
								bind:value={email}
								placeholder="you@example.com"
								required
								disabled={loading}
								onkeypress={handleKeyPress}
								class="modern-input w-full"
							/>
						</div>

						<div>
							<label for="password" class="block text-sm font-medium text-zinc-300 dark:text-zinc-300 light:text-zinc-700 mb-2">
								Password
							</label>
							<Input
								id="password"
								type="password"
								bind:value={password}
								placeholder="Enter your password"
								required
								disabled={loading}
								onkeypress={handleKeyPress}
								class="modern-input w-full"
							/>
						</div>
					</div>

					<div class="flex items-center justify-between text-sm">
						<div class="flex items-center">
							<input
								type="checkbox"
								id="remember"
								class="w-4 h-4 rounded border-zinc-600 dark:border-zinc-600 light:border-zinc-300
								       bg-zinc-800 dark:bg-zinc-800 light:bg-zinc-100
								       text-primary-500 focus:ring-2 focus:ring-primary-500 focus:ring-offset-2
								       focus:ring-offset-zinc-900 dark:focus:ring-offset-zinc-900 light:focus:ring-offset-white
								       transition-all"
							/>
							<label for="remember" class="ml-2 text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
								Remember me
							</label>
						</div>
						<a href="/forgot-password" class="font-medium text-primary-400 hover:text-primary-300 transition-colors">
							Forgot password?
						</a>
					</div>

					<Button type="submit" {loading} disabled={loading} class="btn-primary w-full">
						{#if loading}
							<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white inline" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
							Signing in...
						{:else}
							Sign in
						{/if}
					</Button>
				</form>

				<!-- Register Link -->
				<div class="mt-6 pt-6 border-t border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200">
					<p class="text-center text-sm text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
						Don't have an account?
						<a href="/register" class="font-semibold text-primary-400 hover:text-primary-300 dark:text-primary-400 light:text-primary-600 transition-colors ml-1">
							Create one now →
						</a>
					</p>
				</div>
			</div>
		</div>

		<!-- Security Badge -->
		<div class="text-center" in:fade={{ delay: 400 }}>
			<p class="text-xs text-zinc-500 dark:text-zinc-500 light:text-zinc-400 flex items-center justify-center gap-2">
				<svg class="w-4 h-4 text-primary-500" fill="currentColor" viewBox="0 0 20 20">
					<path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
				</svg>
				Secured with enterprise-grade encryption
			</p>
		</div>
	</div>
</div>

<style>
	@keyframes blob {
		0% {
			transform: translate(0px, 0px) scale(1);
		}
		33% {
			transform: translate(30px, -50px) scale(1.1);
		}
		66% {
			transform: translate(-20px, 20px) scale(0.9);
		}
		100% {
			transform: translate(0px, 0px) scale(1);
		}
	}

	.animate-blob {
		animation: blob 7s infinite;
	}

	.animation-delay-2000 {
		animation-delay: 2s;
	}

	.animation-delay-4000 {
		animation-delay: 4s;
	}
</style>
