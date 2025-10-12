<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores/auth.svelte';
	import { authAPI } from '$lib/api/auth';
	import Button from '$lib/components/Button.svelte';
	import Input from '$lib/components/Input.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let loading = $state(false);
	let error = $state('');
	let passwordStrength = $state(0);
	let registrationEnabled = $state(true);
	let checkingStatus = $state(true);

	onMount(async () => {
		try {
			const status = await authAPI.checkRegistrationStatus();
			registrationEnabled = status.enabled;
			if (!status.enabled) {
				error = status.message;
			}
		} catch (err) {
			console.error('Failed to check registration status:', err);
		} finally {
			checkingStatus = false;
		}
	});

	// Calculate password strength
	$effect(() => {
		if (!password) {
			passwordStrength = 0;
			return;
		}

		let strength = 0;
		if (password.length >= 8) strength++;
		if (password.length >= 12) strength++;
		if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength++;
		if (/\d/.test(password)) strength++;
		if (/[^a-zA-Z\d]/.test(password)) strength++;

		passwordStrength = Math.min(strength, 4);
	});

	function getStrengthColor(strength: number) {
		switch (strength) {
			case 0:
			case 1:
				return 'bg-red-500';
			case 2:
				return 'bg-yellow-500';
			case 3:
				return 'bg-blue-500';
			case 4:
				return 'bg-primary-800';
			default:
				return '';
		}
	}

	function getStrengthText(strength: number) {
		switch (strength) {
			case 0:
			case 1:
				return 'Weak';
			case 2:
				return 'Fair';
			case 3:
				return 'Good';
			case 4:
				return 'Strong';
			default:
				return '';
		}
	}

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
</script>

<svelte:head>
	<title>Create Account - VPS Panel</title>
</svelte:head>

<div class="min-h-screen relative overflow-hidden" style="background-color: rgb(var(--bg-primary));">
	<!-- Animated Background -->
	<div class="absolute inset-0 mesh-gradient"></div>
	<div class="absolute inset-0 grid-pattern opacity-20"></div>

	<!-- Floating Orbs -->
	<div class="absolute top-20 right-10 w-72 h-72 bg-accent-600 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob"></div>
	<div class="absolute top-40 left-10 w-72 h-72 bg-primary-700 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-2000"></div>
	<div class="absolute -bottom-8 left-1/3 w-72 h-72 bg-primary-800 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-4000"></div>

	<!-- Theme Toggle - Top Right -->
	<div class="absolute top-6 right-6 z-20">
		<ThemeToggle />
	</div>

	<!-- Register Container -->
	<div class="relative z-10 flex items-center justify-center min-h-screen py-12 px-4 sm:px-6 lg:px-8">
		<div class="w-full max-w-lg space-y-8">
			<!-- Logo & Title -->
			<div class="text-center scale-up">
				<div class="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-brand shadow-xl glow-green mb-6 float">
					<svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
					</svg>
				</div>
				<h1 class="text-4xl font-bold mb-2 bg-gradient-to-r from-primary-600 to-primary-800 bg-clip-text text-transparent">
					Create Account
				</h1>
				<p class="text-lg" style="color: rgb(var(--text-secondary));">
					Start deploying in minutes
				</p>
			</div>

			<!-- Register Form Card -->
			<div class="glass-pro rounded-2xl p-8 shadow-2xl fade-in" style="border-radius: 1.5rem;">
				{#if checkingStatus}
					<div class="text-center py-12">
						<div class="relative w-16 h-16 mx-auto mb-6">
							<div class="absolute inset-0 rounded-full border-4" style="border-color: rgb(var(--border-primary));"></div>
							<div class="absolute inset-0 rounded-full border-4 border-primary-800 border-t-transparent animate-spin"></div>
						</div>
						<p class="text-lg font-medium" style="color: rgb(var(--text-secondary));">Checking registration status...</p>
					</div>
				{:else if !registrationEnabled}
					<div class="text-center py-12 fade-in">
						<div class="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-brand opacity-20 mb-6">
							<svg class="w-10 h-10 text-primary-800" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
							</svg>
						</div>
						<h3 class="text-2xl font-bold mb-3" style="color: rgb(var(--text-primary));">
							Registration Disabled
						</h3>
						<p class="text-base mb-8" style="color: rgb(var(--text-secondary));">
							{error}
						</p>
						<a href="/login" class="inline-flex items-center px-6 py-3 rounded-xl font-semibold text-sm transition-all duration-200 hover:scale-105 hover-lift bg-gradient-brand text-white glow-green-hover">
							<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
							</svg>
							Go to Login
						</a>
					</div>
				{:else}
					<form onsubmit={handleRegister} class="space-y-5">
						{#if error}
							<div class="slide-in-up">
								<Alert variant="error" dismissible ondismiss={() => error = ''}>
									{error}
								</Alert>
							</div>
						{/if}

						<div class="slide-in-left stagger-1">
							<label
								for="name"
								class="block text-sm font-semibold mb-2"
								style="color: rgb(var(--text-primary));"
							>
								Full Name
							</label>
							<div class="relative group">
								<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
									<svg class="w-5 h-5 group-focus-within:text-primary-600 transition-colors" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
									</svg>
								</div>
								<Input
									id="name"
									type="text"
									bind:value={name}
									placeholder="John Doe"
									required
									disabled={loading}
									class="modern-input pl-12 hover-lift"
								/>
							</div>
						</div>

						<div class="slide-in-left stagger-2">
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
									class="modern-input pl-12 hover-lift"
								/>
							</div>
						</div>

						<div class="slide-in-left stagger-3">
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
									placeholder="At least 8 characters"
									required
									disabled={loading}
									class="modern-input pl-12 hover-lift"
								/>
							</div>

							<!-- Password Strength Indicator -->
							{#if password}
								<div class="mt-3 space-y-2 fade-in">
									<div class="flex items-center gap-1.5">
										{#each Array(4) as _, i}
											<div
												class="h-2 flex-1 rounded-full transition-all duration-300 {i < passwordStrength
													? getStrengthColor(passwordStrength)
													: ''}"
												style="{i >= passwordStrength ? 'background-color: rgb(var(--border-primary))' : ''}"
											></div>
										{/each}
									</div>
									<div class="flex items-center justify-between">
										<p class="text-xs font-medium {passwordStrength >= 3 ? 'text-primary-800' : 'text-gray-500'}">
											Strength: {getStrengthText(passwordStrength)}
										</p>
										{#if passwordStrength >= 3}
											<svg class="w-4 h-4 text-primary-800" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
											</svg>
										{/if}
									</div>
								</div>
							{/if}
						</div>

						<div class="slide-in-left stagger-4">
							<label
								for="confirmPassword"
								class="block text-sm font-semibold mb-2"
								style="color: rgb(var(--text-primary));"
							>
								Confirm Password
							</label>
							<div class="relative group">
								<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
									<svg class="w-5 h-5 group-focus-within:text-primary-600 transition-colors" style="color: rgb(var(--text-tertiary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
									</svg>
								</div>
								<Input
									id="confirmPassword"
									type="password"
									bind:value={confirmPassword}
									placeholder="Re-enter your password"
									required
									disabled={loading}
									class="modern-input pl-12 hover-lift"
								/>
							</div>
							{#if confirmPassword && password !== confirmPassword}
								<div class="mt-2 flex items-center gap-2 text-xs text-red-500 font-medium fade-in">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
									Passwords do not match
								</div>
							{/if}
						</div>

						<div class="pt-2 slide-in-up stagger-4">
							<Button
								type="submit"
								{loading}
								disabled={loading}
								class="btn-primary w-full text-base py-4 glow-green-hover"
							>
								{#if loading}
									<span class="spinner mr-2"></span>
									Creating your account...
								{:else}
									<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
									</svg>
									Create Account
								{/if}
							</Button>
						</div>

						<p class="text-xs text-center pt-2" style="color: rgb(var(--text-tertiary));">
							By creating an account, you agree to our
							<a
								href="/terms"
								class="text-primary-700 hover:text-primary-600 transition-colors font-medium"
							>
								Terms
							</a>
							and
							<a
								href="/privacy"
								class="text-primary-700 hover:text-primary-600 transition-colors font-medium"
							>
								Privacy Policy
							</a>
						</p>
					</form>

					<!-- Divider -->
					<div class="relative my-8">
						<div class="absolute inset-0 flex items-center">
							<div class="w-full border-t" style="border-color: rgb(var(--border-primary));"></div>
						</div>
						<div class="relative flex justify-center text-sm">
							<span class="px-4 text-sm font-medium" style="background-color: rgb(var(--bg-secondary) / 0.7); color: rgb(var(--text-tertiary));">
								Already have an account?
							</span>
						</div>
					</div>

					<!-- Login Link -->
					<div class="text-center fade-in">
						<a
							href="/login"
							class="inline-flex items-center justify-center px-6 py-3 rounded-xl font-semibold text-sm transition-all duration-200 hover:scale-105 hover-lift gradient-border"
							style="color: rgb(var(--text-primary));"
						>
							<svg class="w-5 h-5 mr-2 text-primary-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
							</svg>
							Sign In Instead
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
