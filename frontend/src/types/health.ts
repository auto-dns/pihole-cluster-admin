export type HealthSummary = {
	online: number;
	total: number;
	updatedAt: string;
};

export const NodeStatus = {
	ONLINE: 'ONLINE',
	OFFLINE: 'OFFLINE',
	DEGRADED: 'DEGRADED',
} as const;

export type NodeStatus = (typeof NodeStatus)[keyof typeof NodeStatus];

export type NodeHealth = {
	id: number;
	name: string;
	status: NodeStatus;
	latencyMs: number;
	lastErr?: string;
	updatedAt: number;
};
