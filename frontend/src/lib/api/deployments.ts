import { api } from './client';
import type { Deployment, BuildLog } from '$lib/types';

export const deploymentsAPI = {
	async getAll(projectId: number): Promise<{ deployments: Deployment[]; total: number }> {
		return api.get(`/projects/${projectId}/deployments`);
	},

	async getById(projectId: number, deploymentId: number): Promise<Deployment> {
		return api.get(`/projects/${projectId}/deployments/${deploymentId}`);
	},

	async create(projectId: number): Promise<Deployment> {
		return api.post(`/projects/${projectId}/deployments`);
	},

	async cancel(projectId: number, deploymentId: number): Promise<Deployment> {
		return api.post(`/projects/${projectId}/deployments/${deploymentId}/cancel`);
	},

	async getLogs(projectId: number, deploymentId: number): Promise<{ logs: BuildLog[]; total: number }> {
		return api.get(`/projects/${projectId}/deployments/${deploymentId}/logs`);
	}
};
