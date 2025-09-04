import { ClusterHealth } from '@/types/health';
import { apiV1Fetch } from './client';

export async function getClusterHealth(): Promise<ClusterHealth> {
	return apiV1Fetch<ClusterHealth>('/cluster/health');
}
