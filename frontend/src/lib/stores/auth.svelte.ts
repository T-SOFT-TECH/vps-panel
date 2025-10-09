// Auth store using Svelte 5 runes
import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { authAPI } from '$lib/api/auth';
import type { User } from '$lib/types';

class AuthStore {
	user = $state<User | null>(null);
	token = $state<string | null>(null);
	loading = $state(true);

	constructor() {
		if (browser) {
			this.init();
		}
	}

	init() {
		const storedToken = localStorage.getItem('token');
		const storedUser = localStorage.getItem('user');

		if (storedToken && storedUser) {
			this.token = storedToken;
			this.user = JSON.parse(storedUser);
		}

		this.loading = false;
	}

	async login(email: string, password: string) {
		try {
			const response = await authAPI.login({ email, password });
			this.setAuth(response.token, response.user);
			goto('/dashboard');
			return response;
		} catch (error) {
			console.error('Login failed:', error);
			throw error;
		}
	}

	async register(email: string, password: string, name: string) {
		try {
			const response = await authAPI.register({ email, password, name });
			this.setAuth(response.token, response.user);
			goto('/dashboard');
			return response;
		} catch (error) {
			console.error('Registration failed:', error);
			throw error;
		}
	}

	async fetchCurrentUser() {
		try {
			const user = await authAPI.getCurrentUser();
			this.user = user;
			if (browser) {
				localStorage.setItem('user', JSON.stringify(user));
			}
			return user;
		} catch (error) {
			console.error('Failed to fetch user:', error);
			this.logout();
			throw error;
		}
	}

	logout() {
		this.user = null;
		this.token = null;
		if (browser) {
			localStorage.removeItem('token');
			localStorage.removeItem('user');
			goto('/login');
		}
	}

	private setAuth(token: string, user: User) {
		this.token = token;
		this.user = user;
		if (browser) {
			localStorage.setItem('token', token);
			localStorage.setItem('user', JSON.stringify(user));
		}
	}

	get isAuthenticated() {
		return !!this.token && !!this.user;
	}
}

export const authStore = new AuthStore();
