import { apiV1Fetch } from './client';
import { ClusterBlockingState } from '@/types/blocking';

export async function getClusterBlocking(): Promise<ClusterBlockingState> {
	return apiV1Fetch<ClusterBlockingState>('/cluster/blocking');
}

export type BlockingPostBody = {
	blocking: boolean;
	timer?: number;
};
export async function setClusterBlocking(body: BlockingPostBody): Promise<ClusterBlockingState> {
	return apiV1Fetch<ClusterBlockingState>('/cluster/blocking', {
		method: 'POST',
		body: JSON.stringify(body),
	});
}

export async function setNodeBlocking(
	nodeId: number,
	body: BlockingPostBody,
): Promise<ClusterBlockingState> {
	return apiV1Fetch<ClusterBlockingState>(`/cluster/blocking/nodes/${nodeId}`, {
		method: 'POST',
		body: JSON.stringify(body),
	});
}

export type FlushCacheNodeResult = {
	node: { id: number; name: string; host: string };
	success: boolean;
	error?: string;
};

export type FlushCacheResult = {
	summary: { total: number; succeeded: number; failed: number };
	nodes: Record<number, FlushCacheNodeResult>;
};

export async function flushCache(): Promise<FlushCacheResult> {
	return apiV1Fetch<FlushCacheResult>('/cluster/blocking/flush-cache', { method: 'POST' });
}
