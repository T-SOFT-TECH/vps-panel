import { api } from './client';
import type { User } from '$lib/types';

export interface LoginRequest {
	email: string;
	password: string;
}

export interface RegisterRequest {
	email: string;
	password: string;
	name: string;
}

export interface AuthResponse {
	token: string;
	refresh_token: string;
	user: User;
}

export const authAPI = {
	async login(credentials: LoginRequest): Promise<AuthResponse> {
		return api.post<AuthResponse>('/auth/login', credentials, { requiresAuth: false });
	},

	async register(data: RegisterRequest): Promise<AuthResponse> {
		return api.post<AuthResponse>('/auth/register', data, { requiresAuth: false });
	},

	async getCurrentUser(): Promise<User> {
		return api.get<User>('/users/me');
	},

	async updateProfile(data: Partial<User>): Promise<User> {
		return api.put<User>('/users/me', data);
	}
};
