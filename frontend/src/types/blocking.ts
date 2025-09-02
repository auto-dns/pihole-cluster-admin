import { PiholeNodeRef } from './pihole';

export type NodeBlockingMode = 'enabled' | 'disabled' | 'failed' | 'unknown';
export type ClusterBlockingMode = 'enabled' | 'disabled' | 'mixed' | 'degraded';

export type ClusterBlockingSummary = {
	mode: ClusterBlockingMode;
	unanimous: boolean;
	counts: {
		total: number;
		enabled: number;
		disabled: number;
		failed: number;
		errors: number;
	};
	timers?: {
		present: boolean;
		minSeconds?: number;
		maxSeconds?: number;
	};
	took: {
		maxSeconds: number;
		avgSeconds: number;
	};
};

export type ClusterBlockingNode = {
	node: PiholeNodeRef;
	blocking: NodeBlockingMode;
	timer?: number | null; // seconds
	took: number; // seconds
	error?: string;
};

export type ClusterBlockingState = {
	summary: ClusterBlockingSummary;
	nodes: Record<number, ClusterBlockingNode>;
};
