import { useMemo } from 'react';
import { ClusterBlockingState } from '@/types/blocking';
import { ClusterHealth } from '@/types/health';
import { useClusterOverview } from '@/hooks/useClusterOverview';
import { Logo } from '@/components/Logo';
import { StatusLight } from '@/components/StatusLight/StatusLight/StatusLight';
import { getBlockingDisplayState } from '@/utils/blockingStatus';
import classNames from 'classnames';
import styles from './ClusterHeader.module.scss';

export function ClusterHeader({
	open,
	clusterOverview,
}: {
	open: boolean;
	clusterOverview: ReturnType<typeof useClusterOverview>;
}) {
	const { blocking, blockingFresh, blockingUpdatedAt, health, healthFresh, healthUpdatedAt } =
		clusterOverview;

	return (
		<div className={classNames(styles.wrapper, { [styles.collapsed]: !open })}>
			<div
				className={classNames(styles.header, { [styles.collapsed]: !open })}
				aria-live='polite'
				role='status'
				aria-atomic='true'
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
	if (blocking === undefined) return null;
	const display = getBlockingDisplayState(blocking?.summary);
	const { icon: Icon, colorVar, variant } = display;

	const title =
		(variant === 'enabled'
			? 'Blocking enabled on all nodes'
			: variant === 'disabled'
				? 'Blocking disabled on all nodes'
				: variant === 'degraded'
					? 'All nodes failed to report blocking state'
					: variant === 'mixed-errors'
						? 'Some nodes failed to report blocking state'
						: 'Blocking state mixed across nodes') + (fresh ? '' : ' (stale)');
	const tooltipTime = updatedAt ? updatedAt.toLocaleString() : null;

	return (
		<div
			className={styles.blocking}
			data-variant={variant}
			title={tooltipTime ? `${title} • Updated ${tooltipTime}` : title}
			aria-label={title}
		>
			<p>Blocking:</p>
			<div className={styles.value}>
				<Icon size={16} className={styles.blockingIcon} style={{ color: colorVar }} />
			</div>
		</div>
	);
}
