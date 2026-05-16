import { apiV1Fetch } from './client';
import type { SyncResponse } from '@/types/sync';

export async function syncFromNode(sourceNodeId: number): Promise<SyncResponse> {
	return apiV1Fetch<SyncResponse>('/sync', {
		method: 'POST',
		body: JSON.stringify({ sourceNodeId }),
	});
}
