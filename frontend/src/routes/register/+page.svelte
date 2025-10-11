<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
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
				return 'bg-primary-500';
			default:
				return 'bg-zinc-700';
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

	function handleKeyPress(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			handleRegister(e);
		}
	}
</script>

<svelte:head>
	<title>Create Account - VPS Panel</title>
</svelte:head>

<div
	class="min-h-screen relative overflow-hidden flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8
            bg-gradient-to-br from-zinc-950 via-zinc-900 to-zinc-950
            dark:from-zinc-950 dark:via-zinc-900 dark:to-zinc-950
            light:from-zinc-50 light:via-white light:to-zinc-100"
>
	<!-- Animated Background Elements -->
	<div class="absolute inset-0 overflow-hidden">
		<div
			class="absolute top-0 -left-4 w-96 h-96 bg-primary-500/10 rounded-full mix-blend-multiply filter blur-3xl opacity-70 animate-blob"
		></div>
		<div
			class="absolute top-0 -right-4 w-96 h-96 bg-emerald-500/10 rounded-full mix-blend-multiply filter blur-3xl opacity-70 animate-blob animation-delay-2000"
		></div>
		<div
			class="absolute -bottom-8 left-20 w-96 h-96 bg-primary-600/10 rounded-full mix-blend-multiply filter blur-3xl opacity-70 animate-blob animation-delay-4000"
		></div>
	</div>

	<!-- Theme Toggle - Top Right -->
	<div class="absolute top-6 right-6 z-20" in:scale={{ delay: 300 }}>
		<ThemeToggle />
	</div>

	<!-- Register Card -->
	<div class="relative z-10 max-w-md w-full space-y-8">
		<!-- Logo & Title -->
		<div class="text-center" in:fly={{ y: -30, duration: 600 }}>
			<!-- Logo with gradient -->
			<div
				class="inline-flex items-center justify-center w-20 h-20 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 shadow-2xl mb-6 transform hover:scale-110 transition-transform duration-200"
			>
				<svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"
					/>
				</svg>
			</div>

			<h1
				class="text-5xl font-bold bg-gradient-to-r from-primary-400 via-primary-500 to-emerald-400 bg-clip-text text-transparent mb-3"
			>
				VPS Panel
			</h1>
			<p class="text-lg text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
				Create your account to get started
			</p>
		</div>

		<!-- Register Form Card -->
		<div in:fly={{ y: 30, duration: 600, delay: 200 }}>
			<div class="modern-card p-8">
				{#if checkingStatus}
					<div class="text-center py-8">
						<div class="relative w-12 h-12 mx-auto mb-4">
							<div class="absolute inset-0 rounded-full border-4 border-zinc-800 dark:border-zinc-800 light:border-zinc-200"></div>
							<div class="absolute inset-0 rounded-full border-4 border-primary-500 border-t-transparent animate-spin"></div>
						</div>
						<p class="text-zinc-400 dark:text-zinc-400 light:text-zinc-600 font-medium">Checking registration status...</p>
					</div>
				{:else if !registrationEnabled}
					<div class="text-center py-8">
						<div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-zinc-800/50 dark:bg-zinc-800/50 light:bg-zinc-100 border border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200 mb-4">
							<svg class="w-8 h-8 text-zinc-500 dark:text-zinc-500 light:text-zinc-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
							</svg>
						</div>
						<h3 class="text-lg font-bold text-zinc-100 dark:text-zinc-100 light:text-zinc-900 mb-2">
							Registration Disabled
						</h3>
						<p class="text-base text-zinc-400 dark:text-zinc-400 light:text-zinc-600 mb-6">
							{error}
						</p>
						<a href="/login" class="text-sm font-medium text-primary-400 hover:text-primary-300 dark:text-primary-400 light:text-primary-600 transition-colors">
							Go to Login →
						</a>
					</div>
				{:else}
					<form onsubmit={handleRegister} class="space-y-5">
						{#if error}
							<div in:scale={{ duration: 200 }}>
								<Alert variant="error" dismissible ondismiss={() => (error = '')}>
									{error}
								</Alert>
							</div>
						{/if}

					<div>
						<label
							for="name"
							class="block text-sm font-medium text-zinc-300 dark:text-zinc-300 light:text-zinc-700 mb-2"
						>
							Full Name
						</label>
						<Input
							id="name"
							type="text"
							bind:value={name}
							placeholder="John Doe"
							required
							disabled={loading}
							onkeypress={handleKeyPress}
							class="modern-input w-full"
						/>
					</div>

					<div>
						<label
							for="email"
							class="block text-sm font-medium text-zinc-300 dark:text-zinc-300 light:text-zinc-700 mb-2"
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
							onkeypress={handleKeyPress}
							class="modern-input w-full"
						/>
					</div>

					<div>
						<label
							for="password"
							class="block text-sm font-medium text-zinc-300 dark:text-zinc-300 light:text-zinc-700 mb-2"
						>
							Password
						</label>
						<Input
							id="password"
							type="password"
							bind:value={password}
							placeholder="At least 8 characters"
							required
							disabled={loading}
							onkeypress={handleKeyPress}
							class="modern-input w-full"
						/>

						<!-- Password Strength Indicator -->
						{#if password}
							<div class="mt-2" in:scale={{ duration: 200 }}>
								<div class="flex items-center gap-1 mb-1">
									{#each Array(4) as _, i}
										<div
											class="h-1 flex-1 rounded-full transition-all duration-300 {i <
											passwordStrength
												? getStrengthColor(passwordStrength)
												: 'bg-zinc-700 dark:bg-zinc-700 light:bg-zinc-300'}"
										></div>
									{/each}
								</div>
								<p
									class="text-xs {passwordStrength >= 3
										? 'text-primary-400'
										: 'text-zinc-400 dark:text-zinc-400 light:text-zinc-600'}"
								>
									Password strength: {getStrengthText(passwordStrength)}
								</p>
							</div>
						{/if}
					</div>

					<div>
						<label
							for="confirmPassword"
							class="block text-sm font-medium text-zinc-300 dark:text-zinc-300 light:text-zinc-700 mb-2"
						>
							Confirm Password
						</label>
						<Input
							id="confirmPassword"
							type="password"
							bind:value={confirmPassword}
							placeholder="Re-enter your password"
							required
							disabled={loading}
							onkeypress={handleKeyPress}
							class="modern-input w-full"
						/>
						{#if confirmPassword && password !== confirmPassword}
							<p class="mt-1 text-xs text-red-400" in:scale={{ duration: 200 }}>
								Passwords do not match
							</p>
						{/if}
					</div>

					<div class="pt-2">
						<Button type="submit" {loading} disabled={loading} class="btn-primary w-full">
							{#if loading}
								<svg
									class="animate-spin -ml-1 mr-3 h-5 w-5 text-white inline"
									fill="none"
									viewBox="0 0 24 24"
								>
									<circle
										class="opacity-25"
										cx="12"
										cy="12"
										r="10"
										stroke="currentColor"
										stroke-width="4"
									></circle>
									<path
										class="opacity-75"
										fill="currentColor"
										d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
									></path>
								</svg>
								Creating your account...
							{:else}
								Create Account
							{/if}
						</Button>
					</div>

					<p class="text-xs text-zinc-500 dark:text-zinc-500 light:text-zinc-400 text-center">
						By creating an account, you agree to our
						<a
							href="/terms"
							class="text-primary-400 hover:text-primary-300 dark:text-primary-400 light:text-primary-600 transition-colors"
							>Terms</a
						>
						and
						<a
							href="/privacy"
							class="text-primary-400 hover:text-primary-300 dark:text-primary-400 light:text-primary-600 transition-colors"
							>Privacy Policy</a
						>
					</p>
					</form>

					<!-- Login Link -->
					<div
						class="mt-6 pt-6 border-t border-zinc-700/50 dark:border-zinc-700/50 light:border-zinc-200"
					>
						<p class="text-center text-sm text-zinc-400 dark:text-zinc-400 light:text-zinc-600">
							Already have an account?
							<a
								href="/login"
								class="font-semibold text-primary-400 hover:text-primary-300 dark:text-primary-400 light:text-primary-600 transition-colors ml-1"
							>
								Sign in →
							</a>
						</p>
					</div>
				{/if}
			</div>
		</div>

		<!-- Features -->
		<div class="grid grid-cols-3 gap-4 mt-8" in:fade={{ delay: 400 }}>
			<div class="text-center">
				<div
					class="inline-flex items-center justify-center w-10 h-10 rounded-lg bg-primary-500/10 border border-primary-500/20 mb-2"
				>
					<svg class="w-5 h-5 text-primary-400" fill="currentColor" viewBox="0 0 20 20">
						<path
							fill-rule="evenodd"
							d="M2.166 4.999A11.954 11.954 0 0010 1.944 11.954 11.954 0 0017.834 5c.11.65.166 1.32.166 2.001 0 5.225-3.34 9.67-8 11.317C5.34 16.67 2 12.225 2 7c0-.682.057-1.35.166-2.001zm11.541 3.708a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
							clip-rule="evenodd"
						/>
					</svg>
				</div>
				<p class="text-xs text-zinc-500 dark:text-zinc-500 light:text-zinc-400">Secure</p>
			</div>
			<div class="text-center">
				<div
					class="inline-flex items-center justify-center w-10 h-10 rounded-lg bg-primary-500/10 border border-primary-500/20 mb-2"
				>
					<svg class="w-5 h-5 text-primary-400" fill="currentColor" viewBox="0 0 20 20">
						<path d="M13 7H7v6h6V7z" />
						<path
							fill-rule="evenodd"
							d="M7 2a1 1 0 012 0v1h2V2a1 1 0 112 0v1h2a2 2 0 012 2v2h1a1 1 0 110 2h-1v2h1a1 1 0 110 2h-1v2a2 2 0 01-2 2h-2v1a1 1 0 11-2 0v-1H9v1a1 1 0 11-2 0v-1H5a2 2 0 01-2-2v-2H2a1 1 0 110-2h1V9H2a1 1 0 010-2h1V5a2 2 0 012-2h2V2zM5 5h10v10H5V5z"
							clip-rule="evenodd"
						/>
					</svg>
				</div>
				<p class="text-xs text-zinc-500 dark:text-zinc-500 light:text-zinc-400">Fast</p>
			</div>
			<div class="text-center">
				<div
					class="inline-flex items-center justify-center w-10 h-10 rounded-lg bg-primary-500/10 border border-primary-500/20 mb-2"
				>
					<svg class="w-5 h-5 text-primary-400" fill="currentColor" viewBox="0 0 20 20">
						<path
							d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"
						/>
						<path
							fill-rule="evenodd"
							d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z"
							clip-rule="evenodd"
						/>
					</svg>
				</div>
				<p class="text-xs text-zinc-500 dark:text-zinc-500 light:text-zinc-400">Simple</p>
			</div>
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
