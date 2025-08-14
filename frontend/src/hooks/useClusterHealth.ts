import { useEffect, useMemo, useState } from 'react';
import { useSSE } from './useSSE';
import { useFreshness } from './useFreshness';
import { HealthSummary, NodeHealth } from '../types/health';
import { getClusterHealthSummary, getNodeHealth } from '../lib/api/healthStatus';

const ACTIVE_INTERVAL_MS = 10_000;
const FRESH_WINDOW_MS = 2 * ACTIVE_INTERVAL_MS;

export function useClusterHealth() {
	const [summary, setSummary] = useState<HealthSummary | undefined>(undefined);
	const [nodeHealth, setNodeHealth] = useState<NodeHealth[]>([]);
	const [nodeHealthUpdatedAt, setNodeHealthUpdatedAt] = useState<number | undefined>(undefined);

	useEffect(() => {
		let cancelled = false;
		(async () => {
			try {
				const [summary, nodeHealth] = await Promise.all([
					getClusterHealthSummary(),
					getNodeHealth(),
				]);
				if (!cancelled) {
					setSummary(summary);
					setNodeHealth(nodeHealth);
					setNodeHealthUpdatedAt(Date.now());
				}
			} catch {
				// Do nothing
			}
		})();
		return () => {
			cancelled = true;
		};
	}, []);

	// Live updates via generic SSE
	useSSE<HealthSummary>('health_summary', (s) => setSummary(s));
	useSSE<NodeHealth[]>('node_health', (nh) => {
		setNodeHealth(nh);
		setNodeHealthUpdatedAt(Date.now());
	});

	const summaryUpdatedAtMs = useMemo(
		() => (summary ? Date.parse(summary.updatedAt) : undefined),
		[summary],
	);

	const nodeHealthById = useMemo(() => {
		const m = new Map<number, NodeHealth>();
		(nodeHealth ?? []).forEach((nh) => m.set(nh.id, nh));
		return m;
	}, [nodeHealth]);

	const summaryIsFresh = useFreshness(summaryUpdatedAtMs, FRESH_WINDOW_MS);
	const nodeHealthIsFresh = useFreshness(nodeHealthUpdatedAt, FRESH_WINDOW_MS);

	return {
		summary,
		summaryIsFresh,
		nodeHealth,
		nodeHealthIsFresh,
		nodeHealthById,
	};
}
