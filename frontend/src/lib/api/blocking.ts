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
