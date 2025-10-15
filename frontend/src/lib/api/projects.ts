import { api } from './client';
import type { Project, CreateProjectRequest, Environment, Domain, DetectionResult } from '$lib/types';

export const projectsAPI = {
	async getAll(): Promise<{ projects: Project[]; total: number }> {
		return api.get('/projects');
	},

	async getById(id: number): Promise<Project> {
		return api.get(`/projects/${id}`);
	},

	async create(data: CreateProjectRequest): Promise<Project> {
		return api.post('/projects', data);
	},

	async update(id: number, data: Partial<CreateProjectRequest>): Promise<Project> {
		return api.put(`/projects/${id}`, data);
	},

	async delete(id: number): Promise<void> {
		return api.delete(`/projects/${id}`);
	},

	async detectFramework(
		gitUrl: string,
		gitBranch: string = 'main',
		gitUsername?: string,
		gitToken?: string,
		rootDirectory?: string
	): Promise<DetectionResult> {
		return api.post('/projects/detect', {
			git_url: gitUrl,
			git_branch: gitBranch,
			git_username: gitUsername,
			git_token: gitToken,
			root_directory: rootDirectory
		});
	},

	async listBranches(gitUrl: string, gitUsername?: string, gitToken?: string): Promise<{ branches: string[] }> {
		return api.post('/projects/branches', {
			git_url: gitUrl,
			git_username: gitUsername,
			git_token: gitToken
		});
	},

	async listDirectories(
		gitUrl: string,
		gitBranch: string = 'main',
		gitUsername?: string,
		gitToken?: string
	): Promise<{ directories: string[] }> {
		return api.post('/projects/directories', {
			git_url: gitUrl,
			git_branch: gitBranch,
			git_username: gitUsername,
			git_token: gitToken
		});
	},

	// Environment variables
	async getEnvironments(projectId: number): Promise<{ environments: Environment[] }> {
		return api.get(`/projects/${projectId}/environments`);
	},

	async addEnvironment(
		projectId: number,
		data: { key: string; value: string; is_secret: boolean }
	): Promise<Environment> {
		return api.post(`/projects/${projectId}/environments`, data);
	},

	async updateEnvironment(
		projectId: number,
		envId: number,
		data: { value: string }
	): Promise<Environment> {
		return api.put(`/projects/${projectId}/environments/${envId}`, data);
	},

	async deleteEnvironment(projectId: number, envId: number): Promise<void> {
		return api.delete(`/projects/${projectId}/environments/${envId}`);
	},

	// Domains
	async getDomains(projectId: number): Promise<{ domains: Domain[] }> {
		return api.get(`/projects/${projectId}/domains`);
	},

	async addDomain(
		projectId: number,
		data: { domain: string; ssl_enabled: boolean }
	): Promise<Domain> {
		return api.post(`/projects/${projectId}/domains`, data);
	},

	async updateDomain(
		projectId: number,
		domainId: number,
		data: { domain?: string; is_active?: boolean; ssl_enabled?: boolean }
	): Promise<Domain> {
		return api.put(`/projects/${projectId}/domains/${domainId}`, data);
	},

	async deleteDomain(projectId: number, domainId: number): Promise<void> {
		return api.delete(`/projects/${projectId}/domains/${domainId}`);
	},

	// Webhooks
	async getWebhookInfo(projectId: number): Promise<{
		enabled: boolean;
		webhook?: {
			secret: string;
			urls: {
				github: string;
				gitlab: string;
				gitea: string;
			};
			branch: string;
		};
	}> {
		return api.get(`/projects/${projectId}/webhook`);
	},

	async enableWebhook(projectId: number): Promise<{
		auto_created: boolean;
		manual_setup_required?: boolean;
		webhook?: {
			secret: string;
			urls: {
				github: string;
				gitlab: string;
				gitea: string;
			};
			branch: string;
		};
		message?: string;
	}> {
		return api.post(`/projects/${projectId}/webhook/enable`, {});
	},

	async disableWebhook(projectId: number): Promise<{
		auto_deleted: boolean;
		message?: string;
	}> {
		return api.post(`/projects/${projectId}/webhook/disable`, {});
	},

	// PocketBase updates
	async checkPocketBaseUpdate(projectId: number): Promise<{
		current_version: string;
		latest_version: string;
		update_available: boolean;
		project_id: number;
		project_name: string;
	}> {
		return api.get(`/projects/${projectId}/pocketbase/check-update`);
	},

	async updatePocketBase(projectId: number): Promise<{
		message: string;
		deployment_id: number;
		current_version: string;
		target_version: string;
	}> {
		return api.post(`/projects/${projectId}/pocketbase/update`, {});
	}
};
