import { HealthSummary, NodeHealth } from '../../types/health';
import apiFetchUnversioned from './client';

export async function getClusterHealthSummary(): Promise<HealthSummary> {
	return apiFetchUnversioned<HealthSummary>('/cluster/health/summary');
}

export async function getNodeHealth(): Promise<NodeHealth[]> {
	return apiFetchUnversioned<NodeHealth[]>('/cluster/health/node');
}
