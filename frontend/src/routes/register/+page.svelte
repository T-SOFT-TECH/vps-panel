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

<div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 transition-colors duration-300" style="background-color: rgb(var(--bg-primary));">
	<!-- Theme Toggle - Top Right -->
	<div class="absolute top-6 right-6 z-20">
		<ThemeToggle />
	</div>

	<!-- Register Card -->
	<div class="w-full max-w-md space-y-8">
		<!-- Logo & Title -->
		<div class="text-center">
			<img
				src="/img/My-logo.webp"
				alt="TSOFT Technologies"
				class="mx-auto h-20 mb-6"
			/>
			<h1 class="text-3xl font-bold mb-2" style="color: rgb(var(--text-primary));">
				Create Your Account
			</h1>
			<p class="text-base" style="color: rgb(var(--text-secondary));">
				Get started with VPS Panel
			</p>
		</div>

		<!-- Register Form Card -->
		<div class="elevated-card p-8">
			{#if checkingStatus}
				<div class="text-center py-8">
					<div class="relative w-16 h-16 mx-auto mb-4">
						<div class="absolute inset-0 rounded-full border-4" style="border-color: rgb(var(--border-primary));"></div>
						<div class="absolute inset-0 rounded-full border-4 border-primary-800 border-t-transparent animate-spin"></div>
					</div>
					<p class="font-medium" style="color: rgb(var(--text-secondary));">Checking registration status...</p>
				</div>
			{:else if !registrationEnabled}
				<div class="text-center py-8">
					<div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl mb-4" style="background-color: rgb(var(--bg-secondary)); border: 1px solid rgb(var(--border-primary));">
						<svg class="w-8 h-8" style="color: rgb(var(--text-secondary));" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
						</svg>
					</div>
					<h3 class="text-lg font-bold mb-2" style="color: rgb(var(--text-primary));">
						Registration Disabled
					</h3>
					<p class="text-base mb-6" style="color: rgb(var(--text-secondary));">
						{error}
					</p>
					<a href="/login" class="text-sm font-semibold text-primary-800 hover:text-primary-700 transition-colors">
						Go to Login →
					</a>
				</div>
			{:else}
				<form onsubmit={handleRegister} class="space-y-6">
					{#if error}
						<Alert variant="error" dismissible ondismiss={() => error = ''}>
							{error}
						</Alert>
					{/if}

					<div>
						<label
							for="name"
							class="block text-sm font-medium mb-2"
							style="color: rgb(var(--text-primary));"
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
							class="modern-input"
						/>
					</div>

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
							placeholder="At least 8 characters"
							required
							disabled={loading}
							class="modern-input"
						/>

						<!-- Password Strength Indicator -->
						{#if password}
							<div class="mt-3">
								<div class="flex items-center gap-1 mb-2">
									{#each Array(4) as _, i}
										<div
											class="h-1.5 flex-1 rounded-full transition-all duration-300 {i < passwordStrength
												? getStrengthColor(passwordStrength)
												: ''}"
											style="{i >= passwordStrength ? 'background-color: rgb(var(--border-primary))' : ''}"
										></div>
									{/each}
								</div>
								<p
									class="text-xs {passwordStrength >= 3
										? 'text-primary-800 font-medium'
										: ''}"
									style="{passwordStrength < 3 ? 'color: rgb(var(--text-tertiary))' : ''}"
								>
									Password strength: {getStrengthText(passwordStrength)}
								</p>
							</div>
						{/if}
					</div>

					<div>
						<label
							for="confirmPassword"
							class="block text-sm font-medium mb-2"
							style="color: rgb(var(--text-primary));"
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
							class="modern-input"
						/>
						{#if confirmPassword && password !== confirmPassword}
							<p class="mt-2 text-xs text-red-500 font-medium">
								Passwords do not match
							</p>
						{/if}
					</div>

					<Button type="submit" {loading} disabled={loading} class="btn-primary w-full">
						{#if loading}
							<span class="spinner mr-2"></span>
							Creating your account...
						{:else}
							Create Account
						{/if}
					</Button>

					<p class="text-xs text-center" style="color: rgb(var(--text-tertiary));">
						By creating an account, you agree to our
						<a
							href="/terms"
							class="text-primary-800 hover:text-primary-700 transition-colors"
						>
							Terms
						</a>
						and
						<a
							href="/privacy"
							class="text-primary-800 hover:text-primary-700 transition-colors"
						>
							Privacy Policy
						</a>
					</p>
				</form>

				<!-- Login Link -->
				<div class="mt-6 pt-6 border-t" style="border-color: rgb(var(--border-primary));">
					<p class="text-center text-sm" style="color: rgb(var(--text-secondary));">
						Already have an account?
						<a href="/login" class="font-semibold text-primary-800 hover:text-primary-700 transition-colors ml-1">
							Sign in →
						</a>
					</p>
				</div>
			{/if}
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
