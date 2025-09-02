// src/hooks/useClusterOverview.ts
import { useMemo } from 'react';
import { useClusterHealth } from './useClusterHealth';
import { useClusterBlocking } from './useClusterBlocking';

export function useClusterOverview() {
	const { health, summary: healthSummary, isFresh: healthFresh } = useClusterHealth();
	const { blocking, isFresh: blockingFresh } = useClusterBlocking();

	const nodes = useMemo(() => {
		// join by id: health.nodes (map) + blocking.nodes (map)
		const out: Array<{
			id: number;
			name: string;
			host?: string;
			status?: string; // health
			latencyMs?: number; // health
			blocking?: string; // blocking
			timer?: number | null;
			error?: string; // either side
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
				name: h?.name ?? b?.node.name ?? '',
				host: b?.node.host ?? undefined,
				status: h?.status,
				latencyMs: h ? h.latencyMs : undefined,
				blocking: b?.blocking,
				timer: b?.timer ?? null,
				error: b?.error || h?.lastErr,
			});
		});

		return out;
	}, [health, blocking]);

	return {
		health,
		healthSummary,
		blocking,
		nodes, // merged per-node view for the card
		isFresh: healthFresh && blockingFresh,
	};
}
