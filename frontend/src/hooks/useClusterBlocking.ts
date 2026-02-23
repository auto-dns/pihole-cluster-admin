import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useStampedState } from './useStampedState';
import { useSSE } from './useSSE';
import { useFreshness } from './useFreshness';
import { getClusterBlocking } from '@/lib/api/blocking';
import type { ClusterBlockingState, ClusterBlockingNode } from '@/types/blocking';

const ACTIVE_INTERVAL_MS = 10_000;
const FRESH_WINDOW_MS = 2 * ACTIVE_INTERVAL_MS;
// Ignore SSE for this long after applyState (POST response) so a concurrent poll
// doesn't overwrite our correct state with stale GET data (poller runs every ~5s)
const APPLY_GRACE_MS = 6000;

export type RequestedTimer = { minSeconds: number; startedAt: number };

export type ApplyStateOpts = {
	requestedTimerSeconds?: number;
	/** After POST for a single node, keep this node's timer so the UI shows what we requested. */
	overrideNodeTimer?: { nodeId: number; seconds: number };
	/** After POST for cluster disable, set all nodes' timer to this so the UI shows what we requested. */
	overrideAllNodesTimer?: number;
};

export function useClusterBlocking() {
	const {
		value: state,
		set: setState,
		receivedAt,
	} = useStampedState<ClusterBlockingState | undefined>(undefined);

	const lastApplyAt = useRef<number>(0);
	const [requestedTimer, setRequestedTimer] = useState<RequestedTimer | null>(null);
	/** Display override so the UI shows the requested timer from the moment we click until the response is applied. */
	const [requestedNodeDisplay, setRequestedNodeDisplay] = useState<{
		nodeId: number;
		seconds: number;
		requestedAt: number;
	} | null>(null);

	const refetch = useCallback(() => {
		return getClusterBlocking()
			.then((s) => {
				setState(s);
				if (s?.summary?.mode === 'enabled') {
					setRequestedTimer(null);
				}
				return s;
			})
			.catch(() => undefined);
	}, [setState]);

	const applyState = useCallback(
		(next: ClusterBlockingState, opts?: ApplyStateOpts) => {
			lastApplyAt.current = Date.now();
			let toApply = next;

			const overrideId = opts?.overrideNodeTimer?.nodeId;
			const nodeFromResponse =
				overrideId != null && next.nodes
					? next.nodes[overrideId] ?? next.nodes[String(overrideId)]
					: undefined;
			if (opts?.overrideNodeTimer != null && nodeFromResponse) {
				const id = opts.overrideNodeTimer.nodeId;
				const sec = opts.overrideNodeTimer.seconds;
				const n = nodeFromResponse;
				toApply = {
					...next,
					nodes: {
						...next.nodes,
						[id]: { ...n, blocking: 'disabled' as const, timer: sec },
					},
				};
				// Do NOT clear requestedNodeDisplay here: clearing in the same tick as setState(toApply)
				// can cause one render with override=null and stale state (server timer), showing 11–13.
				// Caller clears in finally so override stays until response is applied.
			} else if (
				opts?.overrideAllNodesTimer != null &&
				next.nodes &&
				Object.keys(next.nodes).length > 0
			) {
				const sec = opts.overrideAllNodesTimer;
				const nodes: Record<number, ClusterBlockingNode> = {};
				for (const [k, node] of Object.entries(next.nodes)) {
					const id = Number(k);
					nodes[id] =
						node.blocking === 'disabled'
							? { ...node, timer: sec }
							: node;
				}
				toApply = { ...next, nodes };
			}

			// Always refresh receivedAt when applying server state so countdowns are "as of now";
			// applyOptimisticNodeDisable keeps receivedAt to avoid flashing other nodes' timers.
			setState(toApply);
			if (toApply.summary?.mode === 'enabled') {
				setRequestedTimer(null);
			} else if (opts?.requestedTimerSeconds != null && opts.requestedTimerSeconds > 0) {
				setRequestedTimer({
					minSeconds: opts.requestedTimerSeconds,
					startedAt: Date.now(),
				});
			}
		},
		[setState],
	);

	const clearRequestedTimer = useCallback(() => {
		setRequestedTimer(null);
	}, []);

	/** When a single node's countdown hits 0, set that node to enabled and recompute summary (e.g. mixed). */
	const setOptimisticNodeEnabled = useCallback(
		(nodeId: number) => {
			if (!state?.summary || !state.nodes || !state.nodes[nodeId]) return;
			const node = state.nodes[nodeId];
			if (node.blocking === 'enabled') return;
			const newNodes: Record<number, ClusterBlockingNode> = {
				...state.nodes,
				[nodeId]: { ...node, blocking: 'enabled' as const, timer: null },
			};
			const enabledCount = Object.values(newNodes).filter((n) => n.blocking === 'enabled').length;
			const disabledCount = Object.values(newNodes).filter((n) => n.blocking === 'disabled').length;
			const total = Object.keys(newNodes).length;
			const mode =
				disabledCount === total ? 'disabled' : enabledCount === total ? 'enabled' : 'mixed';
			lastApplyAt.current = Date.now();
			setState({
				...state,
				summary: {
					...state.summary,
					mode,
					unanimous: enabledCount === total || disabledCount === total,
					counts: {
						...state.summary.counts,
						enabled: enabledCount,
						disabled: disabledCount,
						total,
					},
					timers: {
						...state.summary.timers,
						present: disabledCount > 0 && Object.values(newNodes).some((n) => n.timer != null && n.timer > 0),
					},
				},
				nodes: newNodes,
			}, { keepReceivedAt: true });
		},
		[state, setState],
	);

	/** When client-side countdown hits 0, set state to enabled so UI updates immediately. */
	const setOptimisticEnabled = useCallback(() => {
		if (!state?.summary || !state.nodes) return;
		const total = state.summary.counts.total;
		const next: ClusterBlockingState = {
			summary: {
				...state.summary,
				mode: 'enabled',
				unanimous: true,
				counts: {
					...state.summary.counts,
					enabled: total,
					disabled: 0,
				},
				timers: { present: false },
			},
			nodes: Object.fromEntries(
				Object.entries(state.nodes).map(([id, n]) => [
					id,
					{ ...n, blocking: 'enabled' as const, timer: null },
				]),
			),
		};
		lastApplyAt.current = Date.now();
		setState(next);
		setRequestedTimer(null);
	}, [state, setState]);

	/** Set state and display override immediately so the UI shows the requested timer from click. */
	const applyOptimisticNodeDisable = useCallback(
		(nodeId: number, timerSeconds: number) => {
			if (!state?.nodes?.[nodeId]) return;
			const requestedAt = Date.now();
			setRequestedNodeDisplay({ nodeId, seconds: timerSeconds, requestedAt });
			const node = state.nodes[nodeId];
			const newNodes: Record<number, ClusterBlockingNode> = {
				...state.nodes,
				[nodeId]: { ...node, blocking: 'disabled', timer: timerSeconds },
			};
			const enabledCount = Object.values(newNodes).filter((n) => n.blocking === 'enabled').length;
			const disabledCount = Object.values(newNodes).filter((n) => n.blocking === 'disabled').length;
			const total = Object.keys(newNodes).length;
			const mode =
				disabledCount === total ? 'disabled' : enabledCount === total ? 'enabled' : 'mixed';
			lastApplyAt.current = requestedAt;
			// Keep receivedAt so other nodes' client-side countdown (timer - elapsed from receivedAt) stays correct; otherwise we'd flash their timer to the full value for one frame.
			setState(
				{
					...state,
					summary: {
						...state.summary,
						mode,
						unanimous: enabledCount === total || disabledCount === total,
						counts: {
							...state.summary.counts,
							enabled: enabledCount,
							disabled: disabledCount,
							total,
						},
						timers: { present: true },
					},
					nodes: newNodes,
				},
				{ keepReceivedAt: true },
			);
		},
		[state, setState],
	);

	const clearRequestedNodeDisplay = useCallback((nodeId?: number) => {
		setRequestedNodeDisplay((prev) =>
			nodeId != null && prev?.nodeId === nodeId ? null : prev,
		);
	}, []);

	/** Set state immediately to all nodes disabled + requested timer before cluster disable API call. */
	const applyOptimisticClusterDisable = useCallback(
		(timerSeconds: number) => {
			if (!state?.summary || !state.nodes) return;
			const total = state.summary.counts.total;
			const newNodes: Record<number, ClusterBlockingNode> = {};
			for (const [id, n] of Object.entries(state.nodes)) {
				newNodes[Number(id)] = { ...n, blocking: 'disabled' as const, timer: timerSeconds };
			}
			lastApplyAt.current = Date.now();
			setState({
				...state,
				summary: {
					...state.summary,
					mode: 'disabled',
					unanimous: true,
					counts: {
						...state.summary.counts,
						enabled: 0,
						disabled: total,
					},
					timers: { present: true },
				},
				nodes: newNodes,
			});
			setRequestedTimer({ minSeconds: timerSeconds, startedAt: Date.now() });
		},
		[state, setState],
	);

	// Set requested timer immediately on user action (e.g. pointer down) so we never
	// flash server value when menu closes or SSE delivers stale state before applyState runs.
	const setRequestedTimerDisplay = useCallback((value: RequestedTimer | null) => {
		setRequestedTimer(value);
	}, []);

	useEffect(() => {
		refetch();
	}, [refetch]);

	useSSE<ClusterBlockingState>('v1.cluster_blocking', (s) => {
		if (Date.now() - lastApplyAt.current < APPLY_GRACE_MS) return;
		setState(s);
	});
	const updatedAt = useMemo(() => (receivedAt ? new Date(receivedAt) : undefined), [receivedAt]);
	const isFresh = useFreshness(receivedAt, FRESH_WINDOW_MS);

	return {
		blocking: state,
		isFresh,
		updatedAt,
		refetch,
		applyState,
		setOptimisticEnabled,
		setOptimisticNodeEnabled,
		applyOptimisticNodeDisable,
		applyOptimisticClusterDisable,
		requestedNodeDisplay: requestedNodeDisplay,
		clearRequestedNodeDisplay,
		requestedTimer,
		clearRequestedTimer,
		setRequestedTimerDisplay,
	};
}
