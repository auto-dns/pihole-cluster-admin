import { useEffect, useMemo } from 'react';
import { useStampedState } from './useStampedState';
import { useSSE } from './useSSE';
import { useFreshness } from './useFreshness';
import { getClusterBlocking } from '@/lib/api/blocking';
import type { ClusterBlockingState } from '@/types/blocking';

const ACTIVE_INTERVAL_MS = 10_000;
const FRESH_WINDOW_MS = 2 * ACTIVE_INTERVAL_MS;

export function useClusterBlocking() {
	const {
		value: state,
		set: setState,
		receivedAt,
	} = useStampedState<ClusterBlockingState | undefined>(undefined);

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
	const updatedAt = useMemo(() => (receivedAt ? new Date(receivedAt) : undefined), [receivedAt]);
	const isFresh = useFreshness(receivedAt, FRESH_WINDOW_MS);

	return { blocking: state, isFresh, updatedAt };
}
