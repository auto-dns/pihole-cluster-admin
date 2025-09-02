import apiFetchV1 from './client';
import { ClusterBlockingState } from '../../types/blocking';

export async function getClusterBlocking(): Promise<ClusterBlockingState> {
	return apiFetchV1<ClusterBlockingState>('/cluster/blocking');
}

export type BlockingPostBody = {
	blocking: boolean;
	timer?: number;
};
export async function setClusterBlocking(body: BlockingPostBody): Promise<ClusterBlockingState> {
	return apiFetchV1<ClusterBlockingState>('/cluster/blocking', {
		method: 'POST',
		body: JSON.stringify(body),
	});
}
