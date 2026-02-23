// src/hooks/useClusterOverview.ts
import { useClusterOverviewContext } from '@/providers/ClusterOverviewProvider';

export function useClusterOverview() {
	return useClusterOverviewContext();
}
