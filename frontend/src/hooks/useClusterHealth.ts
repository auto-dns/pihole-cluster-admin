import { useEffect, useMemo, useState } from 'react';
import { useSSE } from './useSSE';
import { useFreshness } from './useFreshness';
import { ClusterHealth, NodeHealth } from '../types/health';
import { getClusterHealth } from '../lib/api/healthStatus';

const ACTIVE_INTERVAL_MS = 10_000;
const FRESH_WINDOW_MS = 2 * ACTIVE_INTERVAL_MS;

export function useClusterHealth() {
	const [health, setHealth] = useState<ClusterHealth | null>(null);

	useEffect(() => {
		let cancelled = false;
		(async () => {
			try {
				const h = await getClusterHealth();
				if (!cancelled) setHealth(h);
			} catch {
				// ignore
			}
		})();
		return () => {
			cancelled = true;
		};
	}, []);

	// Live updates via generic SSE
	useSSE<ClusterHealth>('v1.cluster_health', (s) => setHealth(s));

	const summary = health?.summary;
	const nodeHealthArray: NodeHealth[] = useMemo(
		() => (health ? Object.values(health.nodes ?? {}) : []),
		[health],
	);

	const nodeHealthById = useMemo(() => {
		const m = new Map<number, NodeHealth>();
		nodeHealthArray.forEach((nh) => m.set(nh.id, nh));
		return m;
	}, [nodeHealthArray]);

	const updatedAtMs = useMemo(
		() => (health ? Date.parse(health.updatedAt) : undefined),
		[health],
	);
	const isFresh = useFreshness(updatedAtMs, FRESH_WINDOW_MS);

	return {
		health, // full object if someone wants it
		summary, // convenience
		nodeHealth: nodeHealthArray,
		nodeHealthById,
		isFresh,
	};
}
