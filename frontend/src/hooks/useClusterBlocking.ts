import { useCallback, useEffect, useMemo, useRef } from 'react';
import { useStampedState } from './useStampedState';
import { useSSE } from './useSSE';
import { useFreshness } from './useFreshness';
import { getClusterBlocking } from '@/lib/api/blocking';
import type { ClusterBlockingState } from '@/types/blocking';

const ACTIVE_INTERVAL_MS = 10_000;
const FRESH_WINDOW_MS = 2 * ACTIVE_INTERVAL_MS;
// Ignore SSE for this long after applyState (POST response) so a concurrent poll
// doesn't overwrite our correct state with stale GET data (poller runs every ~5s)
const APPLY_GRACE_MS = 6000;

export function useClusterBlocking() {
	const {
		value: state,
		set: setState,
		receivedAt,
	} = useStampedState<ClusterBlockingState | undefined>(undefined);

	const lastApplyAt = useRef<number>(0);

	const refetch = useCallback(() => {
		return getClusterBlocking()
			.then((s) => {
				setState(s);
				return s;
			})
			.catch(() => undefined);
	}, [setState]);

	const applyState = useCallback(
		(next: ClusterBlockingState) => {
			lastApplyAt.current = Date.now();
			setState(next);
		},
		[setState],
	);

	useEffect(() => {
		refetch();
	}, [refetch]);

	useSSE<ClusterBlockingState>('v1.cluster_blocking', (s) => {
		if (Date.now() - lastApplyAt.current < APPLY_GRACE_MS) return;
		setState(s);
	});
	const updatedAt = useMemo(() => (receivedAt ? new Date(receivedAt) : undefined), [receivedAt]);
	const isFresh = useFreshness(receivedAt, FRESH_WINDOW_MS);

	return { blocking: state, isFresh, updatedAt, refetch, applyState };
}
