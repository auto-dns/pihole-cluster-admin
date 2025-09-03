// src/hooks/useClusterOverview.ts
import { useMemo } from 'react';
import { useClusterHealth } from './useClusterHealth';
import { useClusterBlocking } from './useClusterBlocking';
import { PiholeNodeRef } from '@/types/pihole';
import { NodeHealth } from '@/types/health';
import { ClusterBlockingNode } from '@/types/blocking';

export function useClusterOverview() {
	const { health, summary: healthSummary, isFresh: healthFresh } = useClusterHealth();
	const { blocking, isFresh: blockingFresh } = useClusterBlocking();

	const nodes = useMemo(() => {
		// join by id: health.nodes (map) + blocking.nodes (map)
		const out: Array<{
			id: number;
			node?: PiholeNodeRef;
			health?: NodeHealth;
			blocking?: ClusterBlockingNode;
			error?: string;
		}> = [];

		const ids = new Set<number>([
			...Object.keys(health?.nodes ?? {}).map(Number),
			...Object.keys(blocking?.nodes ?? {}).map(Number),
		]);

		ids.forEach((id) => {
			const h = health?.nodes[id];
			const b = blocking?.nodes[id];
			out.push({
				id,
				node: b?.node,
				health: h,
				blocking: b,
				error: h?.lastErr || b?.error,
			});
		});

		return out;
	}, [health, blocking]);

	return {
		health,
		healthSummary,
		blocking,
		nodes,
		isFresh: healthFresh && blockingFresh,
	};
}
