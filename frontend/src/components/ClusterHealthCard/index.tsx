import { useMemo } from 'react';
import StatusLight from '../StatusLight/StatusLight';
import { Shield, ShieldOff, AlertTriangle } from 'lucide-react';
import { useClusterOverview } from '../../hooks/useClusterOverview';
import styles from './index.module.scss';
import { ClusterHealth } from '@/types/health';
import { ClusterBlockingState } from '@/types/blocking';

export default function ClusterHealthCard({ open }: { open: boolean }) {
	const { blocking, blockingFresh, blockingUpdatedAt, health, healthFresh, healthUpdatedAt } =
		useClusterOverview();
	return (
		<div className={styles.wrapper} aria-live='polite'>
			<NodeHealthStatusCard
				health={health}
				fresh={healthFresh}
				updatedAt={healthUpdatedAt}
				open={open}
			/>
			<NodeBlockingStatusCard
				blocking={blocking}
				fresh={blockingFresh}
				updatedAt={blockingUpdatedAt}
				open={open}
			/>
		</div>
	);
}

type NodeHealthStatusCardProps = {
	health: ClusterHealth | undefined;
	fresh: boolean;
	updatedAt: Date | undefined;
	open: boolean;
};
export function NodeHealthStatusCard({
	health,
	fresh,
	updatedAt,
	open,
}: NodeHealthStatusCardProps) {
	const total = health?.summary?.total ?? 0;
	const online = health?.summary?.online ?? 0;

	const healthColor = useMemo(() => {
		if (total === 0) return 'var(--border-primary)';
		if (online === 0) return 'var(--accent-danger)';
		if (online !== total) return 'var(--accent-warn)';
		return 'var(--accent-success)';
	}, [online, total]);
	const pulse = total !== 0 && (online === total || online !== 0);
	const title = `${online}/${total} nodes online${fresh ? '' : ' (stale)'}`;
	const tooltipTime = updatedAt ? updatedAt.toLocaleString() : null;

	return (
		<div
			className={styles.health}
			data-count={`${online}/${total}`}
			title={tooltipTime ? `${title} • Updated ${tooltipTime}` : title}
		>
			<StatusLight
				label={`${online} of ${total} nodes online`}
				title={`${online} of ${total} nodes online`}
				color={healthColor}
				pulse={pulse}
				durationMs={pulse ? 1800 : 0}
				mode='blink'
			/>
			{open && (
				<>
					<strong>
						{online}/{total}
					</strong>
					<span className={styles.muted}>nodes</span>
				</>
			)}
		</div>
	);
}

const BLOCKING_ICON: Record<string, React.ComponentType<any>> = {
	enabled: Shield,
	disabled: ShieldOff,
	degraded: AlertTriangle,
	mixed: AlertTriangle,
};

type NodeBlockingStatusCardProps = {
	blocking: ClusterBlockingState | undefined;
	fresh: boolean;
	updatedAt: Date | undefined;
	open: boolean;
};
export function NodeBlockingStatusCard({
	blocking,
	fresh,
	updatedAt,
	open,
}: NodeBlockingStatusCardProps) {
	const mode = blocking?.summary?.mode ?? 'degraded';
	const Icon = BLOCKING_ICON[mode] ?? AlertTriangle;

	const title =
		(mode === 'enabled'
			? 'Blocking enabled on all nodes'
			: mode === 'disabled'
				? 'Blocking disabled on all nodes'
				: mode === 'degraded'
					? 'Some nodes failed to report blocking state'
					: 'Blocking state mixed across nodes') + (fresh ? '' : ' (stale)');
	const tooltipTime = updatedAt ? updatedAt.toLocaleString() : null;

	return (
		<div
			className={styles.blocking}
			data-mode={mode}
			title={tooltipTime ? `${title} • Updated ${tooltipTime}` : title}
			aria-label={title}
		>
			<Icon size={16} className={styles.blockingIcon} />
			{open && <span className={styles.muted}>{mode}</span>}
		</div>
	);
}
