import { useMemo } from 'react';
import { ClusterBlockingState } from '@/types/blocking';
import { ClusterHealth } from '@/types/health';
import { Shield, ShieldOff, AlertTriangle } from 'lucide-react';
import { useClusterOverview } from '@/hooks/useClusterOverview';
import Logo from '@/components/Logo';
import StatusLight from '../StatusLight/StatusLight';
import classNames from 'classnames';
import styles from './index.module.scss';

export default function ClusterHeader({ open }: { open: boolean }) {
	const { blocking, blockingFresh, blockingUpdatedAt, health, healthFresh, healthUpdatedAt } =
		useClusterOverview();

	return (
		<div className={classNames(styles.wrapper, { [styles.collapsed]: !open })}>
			<div
				className={classNames(styles.header, { [styles.collapsed]: !open })}
				aria-live='polite'
			>
				<div
					className={classNames(styles.logoWrap, { [styles.minimized]: !open })}
					aria-hidden
				>
					<Logo className={styles.logo} />
				</div>

				<div className={styles.info}>
					<NodeHealthStatusCard
						health={health}
						fresh={healthFresh}
						updatedAt={healthUpdatedAt}
					/>
					<NodeBlockingStatusCard
						blocking={blocking}
						fresh={blockingFresh}
						updatedAt={blockingUpdatedAt}
					/>
				</div>
			</div>
		</div>
	);
}

type NodeHealthStatusCardProps = {
	health: ClusterHealth | undefined;
	fresh: boolean;
	updatedAt: Date | undefined;
};
export function NodeHealthStatusCard({ health, fresh, updatedAt }: NodeHealthStatusCardProps) {
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
			<p>Health:</p>
			<div className={styles.value}>
				<StatusLight
					label={`${online} of ${total} nodes online`}
					title={`${online} of ${total} nodes online`}
					color={healthColor}
					pulse={pulse}
					durationMs={pulse ? 1800 : 0}
					mode='blink'
				/>
			</div>
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
};
export function NodeBlockingStatusCard({
	blocking,
	fresh,
	updatedAt,
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
			<p>Blocking:</p>
			<div className={styles.value}>
				<Icon size={16} className={styles.blockingIcon} />
			</div>
		</div>
	);
}
