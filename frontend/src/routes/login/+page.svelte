<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import {onMount} from "svelte";
	import {authAPI} from "$lib/api/auth";

	let email = $state('');
	let password = $state('');
	let loading = $state(false);
	let error = $state('');
	let registrationEnabled = $state(true);
	let checkingStatus = $state(true);

	onMount(async () => {
		try {
			const status = await authAPI.checkRegistrationStatus();
			registrationEnabled = status.enabled;

		} catch (err) {
			console.error('Failed to check registration status:', err);
		} finally {
			checkingStatus = false;
		}
	});

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

<div class="min-h-screen relative overflow-hidden" style="background-color: rgb(var(--bg-primary));">
	<!-- Animated Background -->
	<div class="absolute inset-0 mesh-gradient"></div>
	<div class="absolute inset-0 dot-pattern opacity-30"></div>

	<!-- Floating Orbs -->
	<div class="absolute top-20 left-10 w-72 h-72 bg-primary-800 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob"></div>
	<div class="absolute top-40 right-10 w-72 h-72 bg-primary-600 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-2000"></div>
	<div class="absolute -bottom-8 left-1/2 w-72 h-72 bg-primary-700 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-4000"></div>

	<!-- Theme Toggle - Top Right -->
	<div class="absolute top-6 right-6 z-20">
		<ThemeToggle />
	</div>

	<!-- Login Container -->
	<div class="relative z-10 flex items-center justify-center min-h-screen py-12 px-4 sm:px-6 lg:px-8">
		<div class="w-full max-w-md space-y-8">
			<!-- Logo & Title -->
			<div class="text-center scale-up">
				<div class="inline-flex items-center justify-center rounded-2xl  mb-6 ">
					<img
						src="/img/My-logo.webp"
						alt="TSOFT Technologies"
						class="h-20 rounded-xl"
					/>
				</div>
				<h1 class="text-4xl font-bold mb-2 bg-gradient-to-r from-primary-600 to-primary-800 bg-clip-text text-transparent">
					Welcome Back
				</h1>
				<p class="text-lg" style="color: rgb(var(--text-secondary));">
					Sign in to continue to VPS Panel
				</p>
			</div>

			<!-- Login Form Card -->
			<div class="glass-pro rounded-2xl p-8 shadow-2xl fade-in" style="border-radius: 1.5rem;">
				<form onsubmit={handleLogin} class="space-y-6">
					{#if error}
						<div class="slide-in-up">
							<Alert variant="error" dismissible ondismiss={() => error = ''}>
								{error}
							</Alert>
						</div>
					{/if}

					<div class="space-y-5">
						<div class="slide-in-left stagger-1">
							<label
								for="email"
								class="block text-sm font-semibold mb-2"
								style="color: rgb(var(--text-primary));"
							>
								Email Address
							</label>
							<div class="relative group">
								<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
									<svg class="w-5 h-5 group-focus-within:text-primary-600 transition-colors" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.207" />
									</svg>
								</div>
								<Input
									id="email"
									type="email"
									bind:value={email}
									placeholder="you@example.com"
									required
									disabled={loading}
									class="modern-input  hover-lift"
								/>
							</div>
						</div>

						<div class="slide-in-left stagger-2">
							<label
								for="password"
								class="block text-sm font-semibold mb-2"
								style="color: rgb(var(--text-primary));"
							>
								Password
							</label>
							<div class="relative group">
								<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
									<svg class="w-5 h-5 group-focus-within:text-primary-600 transition-colors" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
									</svg>
								</div>
								<Input
									id="password"
									type="password"
									bind:value={password}
									placeholder="Enter your password"
									required
									disabled={loading}
									class="modern-input pl-12 hover-lift"
								/>
							</div>
						</div>
					</div>

					<div class="flex items-center justify-between text-sm slide-in-left stagger-3">
						<label class="flex items-center cursor-pointer group">
							<input
								type="checkbox"
								class="w-4 h-4 rounded border-2 text-primary-800 focus:ring-primary-600 focus:ring-offset-2 transition-all cursor-pointer"
								style="border-color: rgb(var(--border-primary));"
							/>
							<span class="ml-2 group-hover:text-primary-700 transition-colors" style="color: rgb(var(--text-secondary));">
								Remember me
							</span>
						</label>
						<a href="/forgot-password" class="font-semibold text-primary-800 hover:text-primary-600 transition-colors hover:underline">
							Forgot password?
						</a>
					</div>

					<div class="slide-in-up stagger-4">
						<Button
							type="submit"
							{loading}
							disabled={loading}
							class="btn-primary w-full text-base py-4 glow-green-hover"
						>
							{#if loading}
								<span class="spinner mr-2"></span>
								Signing in...
							{:else}
								<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
								</svg>
								Sign in to Dashboard
							{/if}
						</Button>
					</div>
				</form>

				{#if registrationEnabled}
				<!-- Divider -->
				<div class="relative my-8">
					<div class="absolute inset-0 flex items-center">
						<div class="w-full border-t" style="border-color: rgb(var(--border-primary));"></div>
					</div>
					<div class="relative flex justify-center text-sm">
						<span class="px-4 text-sm font-medium" style="background-color: rgb(var(--bg-secondary) / 0.7); color: rgb(var(--text-tertiary));">
							New to VPS Panel?
						</span>
					</div>
				</div>

				<!-- Register Link -->
				<div class="text-center fade-in">
					<a
						href="/register"
						class="inline-flex items-center justify-center px-6 py-3 rounded-xl font-semibold text-sm transition-all duration-200 hover:scale-105 hover-lift gradient-border"
						style="color: rgb(var(--text-primary));"
					>
						<svg class="w-5 h-5 mr-2 text-primary-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
						</svg>
						Create New Account
					</a>
				</div>
					{/if}
			</div>

			<!-- Security Badge -->
			<div class="text-center fade-in">
				<div class="inline-flex items-center gap-2 px-4 py-2 rounded-full glass-card">
					<svg class="w-5 h-5 text-primary-600" fill="currentColor" viewBox="0 0 20 20">
						<path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
					</svg>
					<span class="text-sm font-medium" style="color: rgb(var(--text-secondary));">
						Secured with enterprise-grade encryption
					</span>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	@keyframes blob {
		0%, 100% {
			transform: translate(0, 0) scale(1);
		}
		33% {
			transform: translate(30px, -50px) scale(1.1);
		}
		66% {
			transform: translate(-20px, 20px) scale(0.9);
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
