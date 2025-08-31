import { ClusterHealth } from '../../types/health';
import apiFetchV1 from './client';

export async function getClusterHealth(): Promise<ClusterHealth> {
	return apiFetchV1<ClusterHealth>('/cluster/health');
}
