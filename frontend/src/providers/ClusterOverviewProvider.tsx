import {
	createContext,
	useContext,
	useMemo,
	ReactNode,
} from 'react';
import { useClusterHealth } from '@/hooks/useClusterHealth';
import { useClusterBlocking } from '@/hooks/useClusterBlocking';
import type { PiholeNodeRef } from '@/types/pihole';
import type { NodeHealth } from '@/types/health';
import type { ClusterBlockingNode } from '@/types/blocking';

export type ClusterOverviewValue = {
	health: ReturnType<typeof useClusterHealth>['health'];
	healthSummary: ReturnType<typeof useClusterHealth>['summary'];
	healthFresh: boolean;
	healthUpdatedAt: Date | undefined;
	blocking: ReturnType<typeof useClusterBlocking>['blocking'];
	blockingFresh: boolean;
	blockingUpdatedAt: Date | undefined;
	refetchBlocking: ReturnType<typeof useClusterBlocking>['refetch'];
	applyBlockingState: ReturnType<typeof useClusterBlocking>['applyState'];
	setOptimisticEnabled: ReturnType<typeof useClusterBlocking>['setOptimisticEnabled'];
	setOptimisticNodeEnabled: ReturnType<typeof useClusterBlocking>['setOptimisticNodeEnabled'];
	applyOptimisticNodeDisable: ReturnType<typeof useClusterBlocking>['applyOptimisticNodeDisable'];
	applyOptimisticClusterDisable: ReturnType<typeof useClusterBlocking>['applyOptimisticClusterDisable'];
	requestedNodeDisplay: ReturnType<typeof useClusterBlocking>['requestedNodeDisplay'];
	clearRequestedNodeDisplay: ReturnType<typeof useClusterBlocking>['clearRequestedNodeDisplay'];
	requestedTimer: ReturnType<typeof useClusterBlocking>['requestedTimer'];
	clearRequestedTimer: ReturnType<typeof useClusterBlocking>['clearRequestedTimer'];
	setRequestedTimerDisplay: ReturnType<typeof useClusterBlocking>['setRequestedTimerDisplay'];
	nodes: Array<{
		id: number;
		node?: PiholeNodeRef;
		health?: NodeHealth;
		blocking?: ClusterBlockingNode;
		error?: string;
	}>;
	isFresh: boolean;
};

const ClusterOverviewContext = createContext<ClusterOverviewValue | null>(null);

export function ClusterOverviewProvider({ children }: { children: ReactNode }) {
	const health = useClusterHealth();
	const blocking = useClusterBlocking();

	const value = useMemo<ClusterOverviewValue>(() => {
		const ids = new Set<number>([
			...Object.keys(health.health?.nodes ?? {}).map(Number),
			...Object.keys(blocking.blocking?.nodes ?? {}).map(Number),
		]);
		const nodes = Array.from(ids)
			.map((id) => {
				const h = health.health?.nodes?.[id];
				const b = blocking.blocking?.nodes?.[id];
				return {
					id,
					node: b?.node,
					health: h,
					blocking: b,
					error: h?.lastErr || b?.error,
				};
			})
			.sort((a, b) => (a.node?.name ?? '').localeCompare(b.node?.name ?? ''));

		return {
			health: health.health,
			healthSummary: health.summary,
			healthFresh: health.isFresh,
			healthUpdatedAt: health.updatedAt,
			blocking: blocking.blocking,
			blockingFresh: blocking.isFresh,
			blockingUpdatedAt: blocking.updatedAt,
			refetchBlocking: blocking.refetch,
			applyBlockingState: blocking.applyState,
			setOptimisticEnabled: blocking.setOptimisticEnabled,
			setOptimisticNodeEnabled: blocking.setOptimisticNodeEnabled,
			applyOptimisticNodeDisable: blocking.applyOptimisticNodeDisable,
			applyOptimisticClusterDisable: blocking.applyOptimisticClusterDisable,
			requestedNodeDisplay: blocking.requestedNodeDisplay,
			clearRequestedNodeDisplay: blocking.clearRequestedNodeDisplay,
			requestedTimer: blocking.requestedTimer,
			clearRequestedTimer: blocking.clearRequestedTimer,
			setRequestedTimerDisplay: blocking.setRequestedTimerDisplay,
			nodes,
			isFresh: health.isFresh && blocking.isFresh,
		};
	}, [
		health.health,
		health.summary,
		health.isFresh,
		health.updatedAt,
		blocking.blocking,
		blocking.isFresh,
		blocking.updatedAt,
		blocking.refetch,
		blocking.applyState,
		blocking.setOptimisticEnabled,
		blocking.setOptimisticNodeEnabled,
		blocking.applyOptimisticNodeDisable,
		blocking.applyOptimisticClusterDisable,
		blocking.requestedNodeDisplay,
		blocking.clearRequestedNodeDisplay,
		blocking.requestedTimer,
		blocking.clearRequestedTimer,
		blocking.setRequestedTimerDisplay,
	]);

	return (
		<ClusterOverviewContext.Provider value={value}>
			{children}
		</ClusterOverviewContext.Provider>
	);
}

export function useClusterOverviewContext(): ClusterOverviewValue {
	const ctx = useContext(ClusterOverviewContext);
	if (ctx == null) {
		throw new Error('useClusterOverviewContext must be used within ClusterOverviewProvider');
	}
	return ctx;
}
