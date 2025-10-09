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

	async deleteDomain(projectId: number, domainId: number): Promise<void> {
		return api.delete(`/projects/${projectId}/domains/${domainId}`);
	}
};
