import { api } from './client';
import type { GitHubRepository, GiteaRepository } from '$lib/types';

export const oauthAPI = {
	// GitHub OAuth
	async getGitHubAuthURL(): Promise<{ url: string }> {
		return api.get('/auth/oauth/github/init');
	},

	async disconnectGitHub(): Promise<void> {
		return api.get('/auth/oauth/github/disconnect');
	},

	async listGitHubRepositories(): Promise<{ repositories: GitHubRepository[] }> {
		return api.get('/auth/oauth/github/repositories');
	},

	// Gitea OAuth
	async getGiteaAuthURL(giteaURL: string): Promise<{ url: string }> {
		return api.post('/auth/oauth/gitea/init', { gitea_url: giteaURL });
	},

	async disconnectGitea(): Promise<void> {
		return api.get('/auth/oauth/gitea/disconnect');
	},

	async listGiteaRepositories(): Promise<{ repositories: GiteaRepository[] }> {
		return api.get('/auth/oauth/gitea/repositories');
	}
};
