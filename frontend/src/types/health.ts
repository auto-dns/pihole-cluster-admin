export type ClusterHealth = {
	summary: HealthSummary;
	nodes: Record<number, NodeHealth>;
};

export type HealthSummary = {
	online: number;
	total: number;
};

export type NodeHealth = {
	id: number;
	name: string;
	status: NodeStatus;
	latencyMs: number;
	lastErr?: string;
	piholeVersion?: string;
	ftlVersion?: string;
	gravityCount?: number;
	gravityUpdatedAt?: number; // unix timestamp (seconds)
};

export const NodeStatus = {
	ONLINE: 'online',
	OFFLINE: 'offline',
	DEGRADED: 'degraded',
} as const;
export type NodeStatus = (typeof NodeStatus)[keyof typeof NodeStatus];
