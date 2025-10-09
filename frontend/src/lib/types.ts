// Type definitions for VPS Panel

export interface User {
	id: number;
	email: string;
	name: string;
	role: 'admin' | 'user';
	github_connected: boolean;
	github_username?: string;
	gitlab_connected: boolean;
	gitlab_username?: string;
	gitea_connected: boolean;
	gitea_username?: string;
	gitea_url?: string;
	created_at: string;
	updated_at: string;
}

export type FrameworkType =
	| 'sveltekit'
	| 'react'
	| 'vue'
	| 'angular'
	| 'nextjs'
	| 'nuxt';

export type BaaSType =
	| 'pocketbase'
	| 'supabase'
	| 'firebase'
	| 'appwrite'
	| '';

export type DeploymentStatus =
	| 'pending'
	| 'building'
	| 'deploying'
	| 'success'
	| 'failed'
	| 'cancelled';

export interface Project {
	id: number;
	name: string;
	description: string;
	user_id: number;
	git_url: string;
	git_branch: string;
	git_username?: string;
	root_directory?: string;
	framework: FrameworkType;
	baas_type: BaaSType;
	build_command: string;
	output_dir: string;
	install_command: string;
	node_version: string;
	frontend_port: number;
	backend_port: number;
	auto_deploy: boolean;
	deployment_path: string;
	status: 'pending' | 'deploying' | 'active' | 'failed';
	last_deployed?: string;
	created_at: string;
	updated_at: string;
	deployments?: Deployment[];
	environments?: Environment[];
	domains?: Domain[];
}

export interface CreateProjectRequest {
	name: string;
	description?: string;
	git_url: string;
	git_branch?: string;
	git_username?: string;
	git_token?: string;
	root_directory?: string;
	framework: FrameworkType;
	baas_type?: BaaSType;
	build_command?: string;
	output_dir?: string;
	install_command?: string;
	node_version?: string;
	frontend_port?: number;
	backend_port?: number;
	auto_deploy?: boolean;
}

export interface Deployment {
	id: number;
	project_id: number;
	commit_hash: string;
	commit_message: string;
	commit_author: string;
	branch: string;
	status: DeploymentStatus;
	started_at?: string;
	completed_at?: string;
	duration: number;
	error_message?: string;
	triggered_by: 'manual' | 'webhook' | 'api';
	triggered_by_id: number;
	created_at: string;
	updated_at: string;
	build_logs?: BuildLog[];
}

export interface Environment {
	id: number;
	project_id: number;
	key: string;
	value: string;
	is_secret: boolean;
	created_at: string;
	updated_at: string;
}

export interface Domain {
	id: number;
	project_id: number;
	domain: string;
	is_active: boolean;
	ssl_enabled: boolean;
	created_at: string;
	updated_at: string;
}

export interface BuildLog {
	id: number;
	deployment_id: number;
	log: string;
	log_type: 'info' | 'error' | 'warning';
	created_at: string;
	updated_at: string;
}

export interface DetectionResult {
	framework: FrameworkType;
	baas_type: BaaSType;
	detected: boolean;
	build_command: string;
	install_command: string;
	output_dir: string;
	start_command: string;
	dev_command: string;
	frontend_port: number;
	backend_port: number;
	node_version: string;
}

export interface GitHubRepository {
	name: string;
	full_name: string;
	private: boolean;
	html_url: string;
	clone_url: string;
	default_branch: string;
}

export interface GiteaRepository {
	name: string;
	full_name: string;
	private: boolean;
	html_url: string;
	clone_url: string;
	default_branch: string;
	owner: {
		username: string;
	};
}

export type ProviderType = 'github' | 'gitea' | 'gitlab';

export interface GitProvider {
	id: number;
	type: ProviderType;
	name: string;
	url?: string;
	connected: boolean;
	username?: string;
	is_default: boolean;
	created_at: string;
}

export interface CreateProviderRequest {
	type: ProviderType;
	name: string;
	url?: string;
	client_id: string;
	client_secret: string;
	is_default?: boolean;
}

export interface UpdateProviderRequest {
	name?: string;
	url?: string;
	client_id?: string;
	client_secret?: string;
	is_default?: boolean;
}
