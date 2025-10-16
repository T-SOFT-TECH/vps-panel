// WebSocket store using Svelte 5 runes
import { browser } from '$app/environment';

export type MessageType = 'deployment_status' | 'build_log' | 'deployment_start' | 'deployment_end';

export interface WebSocketMessage {
	type: MessageType;
	payload: any;
}

export interface DeploymentStatusPayload {
	deploymentId: number;
	projectId: number;
	status: string;
	error?: string;
}

export interface BuildLogPayload {
	deploymentId: number;
	projectId: number;
	message: string;
	level: string;
	timestamp: string;
}

type MessageHandler = (message: WebSocketMessage) => void;

class WebSocketStore {
	private ws: WebSocket | null = null;
	private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
	private reconnectDelay = 3000; // Start with 3 seconds
	private maxReconnectDelay = 30000; // Max 30 seconds
	private messageHandlers: Set<MessageHandler> = new Set();
	private shouldConnect = true;
	private projectId: number | null = null;

	connected = $state(false);
	connecting = $state(false);
	error = $state<string | null>(null);

	constructor() {
		if (browser) {
			// Auto-connect when store is created
			this.connect();
		}
	}

	/**
	 * Subscribe to a specific project's events
	 */
	subscribeToProject(projectId: number) {
		this.projectId = projectId;
		// Reconnect with project filter
		if (this.connected) {
			this.disconnect();
			this.connect();
		}
	}

	/**
	 * Subscribe to all projects
	 */
	subscribeToAll() {
		this.projectId = null;
		// Reconnect without project filter
		if (this.connected) {
			this.disconnect();
			this.connect();
		}
	}

	/**
	 * Connect to WebSocket server
	 */
	connect() {
		if (!browser || this.ws?.readyState === WebSocket.OPEN || this.connecting) {
			return;
		}

		this.connecting = true;
		this.error = null;
		this.shouldConnect = true;

		try {
			const token = localStorage.getItem('token');
			if (!token) {
				console.warn('No auth token found, WebSocket connection skipped');
				this.connecting = false;
				return;
			}

			// Determine WebSocket URL based on environment
			const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
			const host = window.location.host;

			// Build WebSocket URL with authentication
			let wsUrl = `${protocol}//${host}/api/v1/ws?token=${encodeURIComponent(token)}`;

			// Add project filter if specified
			if (this.projectId !== null) {
				wsUrl += `&projectId=${this.projectId}`;
			}

			this.ws = new WebSocket(wsUrl);

			this.ws.onopen = () => {
				console.log('✓ WebSocket connected');
				this.connected = true;
				this.connecting = false;
				this.error = null;
				this.reconnectDelay = 3000; // Reset delay on successful connection
			};

			this.ws.onmessage = (event) => {
				try {
					const message: WebSocketMessage = JSON.parse(event.data);
					// Notify all subscribers
					this.messageHandlers.forEach(handler => handler(message));
				} catch (err) {
					console.error('Failed to parse WebSocket message:', err);
				}
			};

			this.ws.onerror = (event) => {
				console.error('WebSocket error:', event);
				this.error = 'Connection error';
			};

			this.ws.onclose = (event) => {
				console.log('WebSocket closed:', event.code, event.reason);
				this.connected = false;
				this.connecting = false;
				this.ws = null;

				// Attempt to reconnect if connection should be maintained
				if (this.shouldConnect) {
					this.scheduleReconnect();
				}
			};
		} catch (err) {
			console.error('Failed to create WebSocket connection:', err);
			this.error = 'Failed to connect';
			this.connecting = false;
			this.scheduleReconnect();
		}
	}

	/**
	 * Disconnect from WebSocket server
	 */
	disconnect() {
		this.shouldConnect = false;

		if (this.reconnectTimeout) {
			clearTimeout(this.reconnectTimeout);
			this.reconnectTimeout = null;
		}

		if (this.ws) {
			this.ws.close();
			this.ws = null;
		}

		this.connected = false;
		this.connecting = false;
	}

	/**
	 * Schedule reconnection with exponential backoff
	 */
	private scheduleReconnect() {
		if (this.reconnectTimeout) {
			clearTimeout(this.reconnectTimeout);
		}

		console.log(`Reconnecting in ${this.reconnectDelay / 1000}s...`);

		this.reconnectTimeout = setTimeout(() => {
			this.connect();
			// Increase delay for next attempt (exponential backoff)
			this.reconnectDelay = Math.min(this.reconnectDelay * 1.5, this.maxReconnectDelay);
		}, this.reconnectDelay);
	}

	/**
	 * Subscribe to WebSocket messages
	 */
	subscribe(handler: MessageHandler): () => void {
		this.messageHandlers.add(handler);

		// Return unsubscribe function
		return () => {
			this.messageHandlers.delete(handler);
		};
	}

	/**
	 * Send a message through WebSocket (currently not used, but available for future)
	 */
	send(message: any) {
		if (this.ws?.readyState === WebSocket.OPEN) {
			this.ws.send(JSON.stringify(message));
		} else {
			console.warn('WebSocket is not connected, message not sent');
		}
	}
}

export const websocketStore = new WebSocketStore();
