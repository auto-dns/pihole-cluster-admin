import { useCallback, useEffect, useRef, useState } from 'react';
import {
	LineChart,
	Line,
	XAxis,
	YAxis,
	CartesianGrid,
	Tooltip,
	Legend,
	ResponsiveContainer,
} from 'recharts';
import { RefreshCw } from 'lucide-react';
import classNames from 'classnames';
import {
	getStatsSummary,
	getStatsHistory,
	getStatsTopDomains,
	getStatsTopClients,
} from '@/lib/api/stats';
import type {
	StatsSummaryResponse,
	StatsHistoryResponse,
	StatsTopDomainsResponse,
	StatsTopClientsResponse,
	StatsRange,
} from '@/types/stats';
import { formatCount } from '@/utils/formatters';
import styles from './Stats.module.scss';

const RANGES: { label: string; value: StatsRange }[] = [
	{ label: '1h', value: '1h' },
	{ label: '6h', value: '6h' },
	{ label: '24h', value: '24h' },
];

function formatTimestamp(ts: string, range: StatsRange): string {
	const d = new Date(ts);
	if (range === '24h') {
		return d.toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}
	return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export function Stats() {
	const [range, setRange] = useState<StatsRange>('24h');
	const [summary, setSummary] = useState<StatsSummaryResponse | null>(null);
	const [history, setHistory] = useState<StatsHistoryResponse | null>(null);
	const [topDomains, setTopDomains] = useState<StatsTopDomainsResponse | null>(null);
	const [topClients, setTopClients] = useState<StatsTopClientsResponse | null>(null);
	const [loading, setLoading] = useState(true);
	const [rangeLoading, setRangeLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const loadedRef = useRef(false);

	const loadSummary = useCallback(async () => {
		const res = await getStatsSummary();
		setSummary(res);
	}, []);

	const loadRangeData = useCallback(async (r: StatsRange) => {
		const [h, td, tc] = await Promise.all([
			getStatsHistory(r),
			getStatsTopDomains(r),
			getStatsTopClients(r),
		]);
		setHistory(h);
		setTopDomains(td);
		setTopClients(tc);
	}, []);

	const loadAll = useCallback(
		async (r: StatsRange) => {
			setError(null);
			try {
				await Promise.all([loadSummary(), loadRangeData(r)]);
			} catch (e) {
				setError(e instanceof Error ? e.message : 'Failed to load stats');
			}
		},
		[loadSummary, loadRangeData],
	);

	useEffect(() => {
		if (loadedRef.current) return;
		loadedRef.current = true;
		setLoading(true);
		loadAll(range).finally(() => setLoading(false));
	}, [loadAll, range]);

	const handleRangeChange = async (r: StatsRange) => {
		if (r === range) return;
		setRange(r);
		setRangeLoading(true);
		setError(null);
		try {
			await loadRangeData(r);
		} catch (e) {
			setError(e instanceof Error ? e.message : 'Failed to load stats');
		} finally {
			setRangeLoading(false);
		}
	};

	const handleRefresh = async () => {
		setRangeLoading(true);
		setError(null);
		try {
			await loadAll(range);
		} finally {
			setRangeLoading(false);
		}
	};

	const cluster = summary?.cluster;
	const chartData =
		history?.cluster.map((p) => ({
			time: formatTimestamp(p.timestamp, range),
			allowed: p.total - p.blocked,
			blocked: p.blocked,
		})) ?? [];

	const partialNodes = [
		...(summary?.nodes ?? []),
		...(history?.nodes ?? []),
		...(topDomains?.nodes ?? []),
		...(topClients?.nodes ?? []),
	].filter((n) => !n.success);
	const uniqueFailedNodes = [...new Map(partialNodes.map((n) => [n.node.id, n])).values()];

	return (
		<div className={styles.page}>
			<div className={styles.pageHeader}>
				<h1 className={styles.pageTitle}>Stats &amp; Analytics</h1>
				<button
					className={styles.refreshBtn}
					onClick={handleRefresh}
					disabled={rangeLoading}
					aria-label='Refresh stats'
				>
					<RefreshCw size={16} className={classNames({ [styles.spinning]: rangeLoading })} />
					Refresh
				</button>
			</div>

			{error && <div className={styles.errorBanner}>{error}</div>}

			{uniqueFailedNodes.length > 0 && (
				<div className={styles.partialWarning}>
					Partial data — nodes unavailable:{' '}
					{uniqueFailedNodes.map((n) => n.node.name).join(', ')}
				</div>
			)}

			{/* Summary cards */}
			<section className={styles.section} aria-labelledby='summary-heading'>
				<h2 id='summary-heading' className={styles.sectionTitle}>
					Cluster Summary
				</h2>
				<p className={styles.sectionNote}>Session totals — not filtered by range</p>
				<div className={styles.summaryCards}>
					<div className={styles.card}>
						<div className={styles.statValue}>
							{loading ? '—' : formatCount(cluster?.queriesTotal ?? 0)}
						</div>
						<div className={styles.cardSub}>Total Queries</div>
					</div>
					<div className={styles.card}>
						<div className={styles.statValue}>
							{loading ? '—' : `${(cluster?.blockedPercent ?? 0).toFixed(1)}%`}
						</div>
						<div className={styles.cardSub}>Blocked</div>
					</div>
					<div className={styles.card}>
						<div className={styles.statValue}>
							{loading ? '—' : formatCount(cluster?.gravitySize ?? 0)}
						</div>
						<div className={styles.cardSub}>Gravity Size</div>
					</div>
					<div
						className={styles.card}
						title='Sum across all nodes — clients seen by multiple nodes may be counted more than once'
					>
						<div className={styles.statValue}>
							{loading ? '—' : formatCount(cluster?.uniqueClients ?? 0)}
						</div>
						<div className={styles.cardSub}>Unique Clients</div>
					</div>
				</div>
			</section>

			{/* Range controls */}
			<div className={styles.rangeBar}>
				<span className={styles.rangeLabel}>Range:</span>
				<div className={styles.rangeButtons} role='group' aria-label='Time range'>
					{RANGES.map(({ label, value }) => (
						<button
							key={value}
							className={classNames(styles.rangeBtn, { [styles.rangeBtnActive]: range === value })}
							onClick={() => handleRangeChange(value)}
							disabled={rangeLoading}
							aria-pressed={range === value}
						>
							{label}
						</button>
					))}
				</div>
			</div>

			{/* History chart */}
			<section className={styles.section} aria-labelledby='history-heading'>
				<h2 id='history-heading' className={styles.sectionTitle}>
					Query History
				</h2>
				<div className={styles.chartCard}>
					{rangeLoading || loading ? (
						<div className={styles.chartPlaceholder}>Loading…</div>
					) : chartData.length === 0 ? (
						<div className={styles.chartPlaceholder}>No data</div>
					) : (
						<ResponsiveContainer width='100%' height={260}>
							<LineChart data={chartData} margin={{ top: 4, right: 16, bottom: 0, left: 0 }}>
								<CartesianGrid strokeDasharray='3 3' stroke='var(--border-primary)' />
								<XAxis
									dataKey='time'
									tick={{ fontSize: 11, fill: 'var(--text-secondary)' }}
									interval='preserveStartEnd'
								/>
								<YAxis
									tick={{ fontSize: 11, fill: 'var(--text-secondary)' }}
									tickFormatter={formatCount}
									width={48}
								/>
								<Tooltip
									contentStyle={{
										background: 'var(--bg-card)',
										border: '1px solid var(--border-card)',
										borderRadius: 'var(--border-radius)',
										fontSize: 12,
									}}
									formatter={(value: number, name: string) => [
										formatCount(value),
										name === 'blocked' ? 'Blocked' : 'Allowed',
									]}
								/>
								<Legend
									wrapperStyle={{ fontSize: 12, paddingTop: 8 }}
									formatter={(value) => (value === 'blocked' ? 'Blocked' : 'Allowed')}
								/>
								<Line
									type='monotone'
									dataKey='allowed'
									stroke='var(--accent-success)'
									strokeWidth={2}
									dot={false}
									activeDot={{ r: 4 }}
								/>
								<Line
									type='monotone'
									dataKey='blocked'
									stroke='var(--accent-danger)'
									strokeWidth={2}
									dot={false}
									activeDot={{ r: 4 }}
								/>
							</LineChart>
						</ResponsiveContainer>
					)}
				</div>
			</section>

			{/* Top tables */}
			<div className={styles.tablesGrid}>
				<section className={styles.section} aria-labelledby='top-domains-heading'>
					<h2 id='top-domains-heading' className={styles.sectionTitle}>
						Top Domains
					</h2>
					<div className={styles.tableCard}>
						<p className={styles.tableSubheading}>Top permitted</p>
						<table className={styles.table}>
							<thead>
								<tr>
									<th>Domain</th>
									<th className={styles.countCol}>Allowed</th>
								</tr>
							</thead>
							<tbody>
								{rangeLoading || loading ? (
									<tr>
										<td colSpan={2} className={styles.tableEmpty}>
											Loading…
										</td>
									</tr>
								) : (topDomains?.clusterTopQueried ?? []).length === 0 ? (
									<tr>
										<td colSpan={2} className={styles.tableEmpty}>
											No data
										</td>
									</tr>
								) : (
									(topDomains?.clusterTopQueried ?? []).map((d) => (
										<tr key={d.domain}>
											<td className={styles.domainCell}>{d.domain}</td>
											<td className={styles.countCol}>{formatCount(d.count)}</td>
										</tr>
									))
								)}
							</tbody>
						</table>

						<p className={styles.tableSubheadingSpaced}>
							Most blocked
						</p>
						<table className={styles.table}>
							<thead>
								<tr>
									<th>Domain</th>
									<th className={styles.countCol}>Blocked</th>
								</tr>
							</thead>
							<tbody>
								{rangeLoading || loading ? (
									<tr>
										<td colSpan={2} className={styles.tableEmpty}>
											Loading…
										</td>
									</tr>
								) : (topDomains?.clusterTopBlocked ?? []).length === 0 ? (
									<tr>
										<td colSpan={2} className={styles.tableEmpty}>
											No data
										</td>
									</tr>
								) : (
									(topDomains?.clusterTopBlocked ?? []).map((d) => (
										<tr key={d.domain}>
											<td className={styles.domainCell}>{d.domain}</td>
											<td className={styles.countCol}>{formatCount(d.count)}</td>
										</tr>
									))
								)}
							</tbody>
						</table>
					</div>
				</section>

				<section className={styles.section} aria-labelledby='top-clients-heading'>
					<h2 id='top-clients-heading' className={styles.sectionTitle}>
						Top Clients
					</h2>
					<div className={styles.tableCard}>
						<table className={styles.table}>
							<thead>
								<tr>
									<th>Client</th>
									<th className={styles.countCol}>Queries</th>
								</tr>
							</thead>
							<tbody>
								{rangeLoading || loading ? (
									<tr>
										<td colSpan={2} className={styles.tableEmpty}>
											Loading…
										</td>
									</tr>
								) : (topClients?.clusterClients ?? []).length === 0 ? (
									<tr>
										<td colSpan={2} className={styles.tableEmpty}>
											No data
										</td>
									</tr>
								) : (
									(topClients?.clusterClients ?? []).map((c) => (
										<tr key={c.ip}>
											<td>
												<span className={styles.clientName}>{c.name || c.ip}</span>
												{c.name && (
													<span className={styles.clientIp}>{c.ip}</span>
												)}
											</td>
											<td className={styles.countCol}>{formatCount(c.count)}</td>
										</tr>
									))
								)}
							</tbody>
						</table>
					</div>
				</section>
			</div>

			{/* Per-node breakdown */}
			{summary && summary.nodes.length > 1 && (
				<section className={styles.section} aria-labelledby='node-breakdown-heading'>
					<h2 id='node-breakdown-heading' className={styles.sectionTitle}>
						Per-node Breakdown
					</h2>
					<div className={styles.nodeGrid}>
						{summary.nodes.map(({ node, success, error: nodeErr, data }) => (
							<div key={node.id} className={styles.nodeCard} data-success={success}>
								<div className={styles.nodeName}>{node.name}</div>
								{!success ? (
									<div className={styles.nodeError}>{nodeErr || 'Unavailable'}</div>
								) : (
									<dl className={styles.nodeStats}>
										<div className={styles.nodeStat}>
											<dt>Queries</dt>
											<dd>{formatCount(data.queriesTotal)}</dd>
										</div>
										<div className={styles.nodeStat}>
											<dt>Blocked</dt>
											<dd>{data.blockedPercent.toFixed(1)}%</dd>
										</div>
										<div className={styles.nodeStat}>
											<dt>Gravity</dt>
											<dd>{formatCount(data.gravitySize)}</dd>
										</div>
										<div className={styles.nodeStat}>
											<dt>Clients</dt>
											<dd>{formatCount(data.uniqueClients)}</dd>
										</div>
									</dl>
								)}
							</div>
						))}
					</div>
				</section>
			)}
		</div>
	);
}
