import { api } from './client';
import type {
	GitProvider,
	CreateProviderRequest,
	UpdateProviderRequest,
	GitHubRepository,
	GiteaRepository
} from '$lib/types';

export const gitProvidersAPI = {
	// Provider Management
	async getAll(): Promise<{ providers: GitProvider[] }> {
		return api.get('/git-providers');
	},

	async getById(id: number): Promise<GitProvider> {
		return api.get(`/git-providers/${id}`);
	},

	async create(data: CreateProviderRequest): Promise<GitProvider> {
		return api.post('/git-providers', data);
	},

	async update(id: number, data: UpdateProviderRequest): Promise<GitProvider> {
		return api.put(`/git-providers/${id}`, data);
	},

	async delete(id: number): Promise<void> {
		return api.delete(`/git-providers/${id}`);
	},

	async disconnect(id: number): Promise<{ message: string }> {
		return api.post(`/git-providers/${id}/disconnect`);
	},

	// OAuth Flow
	async initiateOAuth(providerId: number): Promise<{ url: string }> {
		const provider = await this.getById(providerId);

		if (provider.type === 'github') {
			return api.get(`/auth/oauth/github/init?provider_id=${providerId}`);
		} else if (provider.type === 'gitea') {
			return api.get(`/auth/oauth/gitea/init?provider_id=${providerId}`);
		} else {
			throw new Error('Provider type not supported yet');
		}
	},

	// Repository Listing
	async listRepositories(
		providerId: number
	): Promise<{ repositories: GitHubRepository[] | GiteaRepository[] }> {
		return api.get(`/git-providers/${providerId}/repositories`);
	}
};
