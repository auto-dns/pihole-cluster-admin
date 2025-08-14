import { HealthSummary, NodeHealth } from '../../types/health';
import apiFetch from './client';

export async function getClusterHealthSummary(): Promise<HealthSummary> {
	return apiFetch<HealthSummary>('/cluster/health/summary');
}

export async function getNodeHealth(): Promise<NodeHealth[]> {
	return apiFetch<NodeHealth[]>('/cluster/health/node');
}
