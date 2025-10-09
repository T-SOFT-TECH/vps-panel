// API Client for VPS Panel Backend
import { browser } from '$app/environment';
import { goto } from '$app/navigation';

// Use relative URL - works with Caddy reverse proxy in production
// and Vite proxy in development
const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

interface FetchOptions extends RequestInit {
	requiresAuth?: boolean;
}

class APIClient {
	private baseURL: string;

	constructor(baseURL: string) {
		this.baseURL = baseURL;
	}

	private getToken(): string | null {
		if (!browser) return null;
		return localStorage.getItem('token');
	}

	private async request<T>(endpoint: string, options: FetchOptions = {}): Promise<T> {
		const { requiresAuth = true, headers = {}, ...fetchOptions } = options;

		const requestHeaders: HeadersInit = {
			'Content-Type': 'application/json',
			...headers
		};

		if (requiresAuth) {
			const token = this.getToken();
			if (token) {
				requestHeaders['Authorization'] = `Bearer ${token}`;
			}
		}

		const url = `${this.baseURL}${endpoint}`;

		try {
			const response = await fetch(url, {
				...fetchOptions,
				headers: requestHeaders,
				credentials: 'include' // Required for cookies to work cross-origin
			});

			if (response.status === 401) {
				// Unauthorized - clear token and redirect to login
				if (browser) {
					localStorage.removeItem('token');
					localStorage.removeItem('user');
					goto('/login');
				}
				throw new Error('Unauthorized');
			}

			if (!response.ok) {
				const error = await response.json().catch(() => ({ error: 'Request failed' }));
				throw new Error(error.error || `HTTP ${response.status}`);
			}

			// Handle 204 No Content
			if (response.status === 204) {
				return {} as T;
			}

			return await response.json();
		} catch (error) {
			console.error('API Error:', error);
			throw error;
		}
	}

	// HTTP Methods
	async get<T>(endpoint: string, options?: FetchOptions): Promise<T> {
		return this.request<T>(endpoint, { ...options, method: 'GET' });
	}

	async post<T>(endpoint: string, data?: unknown, options?: FetchOptions): Promise<T> {
		return this.request<T>(endpoint, {
			...options,
			method: 'POST',
			body: data ? JSON.stringify(data) : undefined
		});
	}

	async put<T>(endpoint: string, data?: unknown, options?: FetchOptions): Promise<T> {
		return this.request<T>(endpoint, {
			...options,
			method: 'PUT',
			body: data ? JSON.stringify(data) : undefined
		});
	}

	async delete<T>(endpoint: string, options?: FetchOptions): Promise<T> {
		return this.request<T>(endpoint, { ...options, method: 'DELETE' });
	}
}

export const api = new APIClient(API_BASE_URL);
