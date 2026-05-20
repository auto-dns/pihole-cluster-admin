import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { ChevronRight } from 'lucide-react';
import { useClusterOverview } from '@/hooks/useClusterOverview';
import { getBlockingDisplayState } from '@/utils/blockingStatus';
import { PiholeStatusLight } from '@/components/StatusLight/PiholeStatusLight';
import { getStatsSummary } from '@/lib/api/stats';
import type { StatsSummaryResponse } from '@/types/stats';
import { formatCount, formatRelativeTime } from '@/utils/formatters';
import styles from './Home.module.scss';

export function Home() {
	const { blocking, nodes, healthFresh, healthSummary } = useClusterOverview();
	const blockingDisplay = getBlockingDisplayState(blocking?.summary);
	const { icon: StatusIcon, colorVar, label: blockingLabel } = blockingDisplay;

	const [statsSummary, setStatsSummary] = useState<StatsSummaryResponse | null>(null);

	useEffect(() => {
		getStatsSummary()
			.then(setStatsSummary)
			.catch(() => {});
	}, []);

	const cluster = statsSummary?.cluster;

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

			<section className={styles.section} aria-labelledby='stats-heading'>
				<div className={styles.sectionTitleRow}>
					<h2 id='stats-heading' className={styles.sectionTitle}>
						Stats
					</h2>
					<Link to='/stats' className={styles.statsLink}>
						View all <ChevronRight size={14} />
					</Link>
				</div>
				<div className={styles.summaryCards}>
					<div className={styles.card}>
						<div className={styles.statValue}>
							{cluster != null ? formatCount(cluster.queriesTotal) : '—'}
						</div>
						<div className={styles.cardSub}>Total Queries</div>
					</div>
					<div className={styles.card}>
						<div className={styles.statValue}>
							{cluster != null ? `${cluster.blockedPercent.toFixed(1)}%` : '—'}
						</div>
						<div className={styles.cardSub}>Blocked</div>
					</div>
					<div className={styles.card}>
						<div className={styles.statValue}>
							{cluster != null ? formatCount(cluster.gravitySize) : '—'}
						</div>
						<div className={styles.cardSub}>Gravity Size</div>
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
									<div className={styles.nodeDetail}>
										<dt>Pi-hole</dt>
										<dd>{health?.piholeVersion ?? '—'}</dd>
									</div>
									<div className={styles.nodeDetail}>
										<dt>FTL</dt>
										<dd>{health?.ftlVersion ?? '—'}</dd>
									</div>
									<div className={styles.nodeDetail}>
										<dt>Gravity</dt>
										<dd>
											{health?.gravityCount != null
												? formatCount(health.gravityCount)
												: '—'}
										</dd>
									</div>
									<div className={styles.nodeDetail}>
										<dt>Updated</dt>
										<dd>
											{health?.gravityUpdatedAt != null
												? formatRelativeTime(health.gravityUpdatedAt)
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
