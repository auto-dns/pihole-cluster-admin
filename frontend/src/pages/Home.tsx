import { useClusterOverview } from '@/hooks/useClusterOverview';
import { getBlockingDisplayState } from '@/utils/blockingStatus';
import { PiholeStatusLight } from '@/components/StatusLight/PiholeStatusLight';
import styles from './Home.module.scss';

export function Home() {
	const { blocking, nodes, healthFresh, healthSummary } = useClusterOverview();
	const blockingDisplay = getBlockingDisplayState(blocking?.summary);
	const { icon: StatusIcon, colorVar, label: blockingLabel } = blockingDisplay;

	return (
		<div className={styles.page}>
			<section className={styles.section} aria-labelledby='cluster-status-heading'>
				<h2 id='cluster-status-heading' className={styles.sectionTitle}>
					Cluster Status
				</h2>
				<div className={styles.summaryCards}>
					<div className={styles.card}>
						{blocking !== undefined ? (
							<div className={styles.blockingRow}>
								<StatusIcon size={22} style={{ color: colorVar }} aria-hidden />
								<span className={styles.blockingLabel}>{blockingLabel}</span>
							</div>
						) : (
							<div className={styles.statValue}>—</div>
						)}
						<div className={styles.cardSub}>Blocking</div>
					</div>
					<div className={styles.card}>
						<div className={styles.statValue}>
							{healthSummary ? (
								<>
									{healthSummary.online}
									<span className={styles.statDenominator}>
										{' '}
										/ {healthSummary.total}
									</span>
								</>
							) : (
								'—'
							)}
						</div>
						<div className={styles.cardSub}>Nodes online</div>
					</div>
				</div>
			</section>

			<section className={styles.section} aria-labelledby='nodes-heading'>
				<h2 id='nodes-heading' className={styles.sectionTitle}>
					Nodes
				</h2>
				{nodes.length === 0 ? (
					<p className={styles.empty}>No nodes configured.</p>
				) : (
					<div className={styles.nodeGrid}>
						{nodes.map(({ id, node, health }) => (
							<div
								key={id}
								className={styles.nodeCard}
								data-status={health?.status ?? 'unknown'}
							>
								<div className={styles.nodeHeader}>
									<PiholeStatusLight
										name={node?.name ?? `Node ${id}`}
										health={health}
										fresh={healthFresh}
									/>
									<span className={styles.nodeName}>
										{node?.name ?? `Node ${id}`}
									</span>
								</div>
								<dl className={styles.nodeDetails}>
									<div className={styles.nodeDetail}>
										<dt>Status</dt>
										<dd
											data-status={health?.status ?? 'unknown'}
											className={styles.nodeStatusValue}
										>
											{health?.status ?? 'unknown'}
										</dd>
									</div>
									<div className={styles.nodeDetail}>
										<dt>Latency</dt>
										<dd>
											{health?.latencyMs != null
												? `${health.latencyMs}ms`
												: '—'}
										</dd>
									</div>
								</dl>
								{health?.lastErr && (
									<div className={styles.nodeError}>{health.lastErr}</div>
								)}
							</div>
						))}
					</div>
				)}
			</section>
		</div>
	);
}
