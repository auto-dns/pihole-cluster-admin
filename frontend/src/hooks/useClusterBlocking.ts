import { useEffect, useMemo, useState } from 'react';
import { useSSE } from './useSSE';
import { useFreshness } from './useFreshness';
import { getClusterBlocking } from '../lib/api/blocking';
import type { ClusterBlockingState } from '../types/blocking';

const ACTIVE_INTERVAL_MS = 10_000;
const FRESH_WINDOW_MS = 2 * ACTIVE_INTERVAL_MS;

export function useClusterBlocking() {
	const [state, setState] = useState<ClusterBlockingState | null>(null);

	useEffect(() => {
		let cancelled = false;
		(async () => {
			try {
				const s = await getClusterBlocking();
				if (!cancelled) setState(s);
			} catch {
				/* ignore */
			}
		})();
		return () => {
			cancelled = true;
		};
	}, []);

	// Live updates
	useSSE<ClusterBlockingState>('v1.cluster_blocking', (s) => setState(s));

	const updatedAtMs = useMemo(
		() => (state?.updatedAt ? Date.parse(state.updatedAt) : undefined),
		[state],
	);
	const isFresh = useFreshness(updatedAtMs, FRESH_WINDOW_MS);

	return { blocking: state, isFresh };
}
