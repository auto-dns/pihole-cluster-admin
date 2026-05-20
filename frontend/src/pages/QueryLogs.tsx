import { useState, useEffect, useCallback, useRef, Fragment, useMemo } from 'react';
import { useSearchParams } from 'react-router';
import { Search, RefreshCw, Shield, ShieldOff, Plus, Minus } from 'lucide-react';
import { getQueryLogs, type QueryLogParams } from '@/lib/api/logs';
import { addDomainRule } from '@/lib/api/domainrules';
import type { MergedEntry } from '@/types/querylog';
import {
	isBlockedStatus,
	isForwardedStatus,
	isCachedStatus,
	presetRange,
	formatTime,
	formatDate,
	statusColor,
	shortStatus,
	mergeAndSort,
} from '@/utils/queryLogHelpers';
import styles from './QueryLogs.module.scss';

type TimePreset = '5m' | '15m' | '1h' | '6h' | '24h';
const INITIAL_PRESET: TimePreset = '5m';

const TIME_PRESETS: { label: string; value: TimePreset; minutes: number }[] = [
	{ label: '5m', value: '5m', minutes: 5 },
	{ label: '15m', value: '15m', minutes: 15 },
	{ label: '1h', value: '1h', minutes: 60 },
	{ label: '6h', value: '6h', minutes: 360 },
	{ label: '24h', value: '24h', minutes: 1440 },
];

const STATUS_OPTIONS = [
	{ label: 'All statuses', value: '' },
	{ label: 'Blocked (any)', value: 'blocked' },
	{ label: 'Forwarded', value: 'forwarded' },
	{ label: 'Cached', value: 'cached' },
	{ label: 'Gravity blocked', value: 'BLOCKED_GRAVITY' },
	{ label: 'Regex blocked', value: 'BLOCKED_REGEX' },
	{ label: 'Exact blocked', value: 'BLOCKED_BLACKLIST' },
	{ label: 'Ext. IP blocked', value: 'BLOCKED_EXTERNAL_IP' },
	{ label: 'Special domain', value: 'SPECIAL_DOMAIN' },
];

const QTYPE_OPTIONS = [
	{ label: 'All types', value: '' },
	{ label: 'A', value: 'A' },
	{ label: 'AAAA', value: 'AAAA' },
	{ label: 'CNAME', value: 'CNAME' },
	{ label: 'MX', value: 'MX' },
	{ label: 'TXT', value: 'TXT' },
	{ label: 'PTR', value: 'PTR' },
	{ label: 'SRV', value: 'SRV' },
	{ label: 'HTTPS', value: 'HTTPS' },
	{ label: 'SOA', value: 'SOA' },
	{ label: 'NS', value: 'NS' },
];

export function QueryLogs() {
	const [searchParams] = useSearchParams();
	const initialClientIp = useRef(searchParams.get('client') ?? '').current;

	const [timePreset, setTimePreset] = useState<TimePreset>('5m');
	const [domain, setDomain] = useState('');
	const [clientIp, setClientIp] = useState(initialClientIp);

	// Client-side filters
	const [statusFilter, setStatusFilter] = useState('');
	const [qtypeFilter, setQtypeFilter] = useState('');
	const [nodeFilter, setNodeFilter] = useState<number | null>(null);
	const [availableNodes, setAvailableNodes] = useState<{ id: number; name: string }[]>([]);

	const [entries, setEntries] = useState<MergedEntry[]>([]);
	const [cursor, setCursor] = useState('');
	const [endOfResults, setEndOfResults] = useState(false);
	const [loading, setLoading] = useState(false);
	const [loadingMore, setLoadingMore] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [nodeErrors, setNodeErrors] = useState<{ nodeId: number; name: string; err: string }[]>(
		[],
	);

	const [actioningDomain, setActioningDomain] = useState<string | null>(null);
	const [actionFeedback, setActionFeedback] = useState<{
		domain: string;
		type: 'allow' | 'deny';
		ok: boolean;
	} | null>(null);
	const feedbackTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

	const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

	const appliedFilters = useRef({ timePreset, domain: '', clientIp: '' });

	const displayedEntries = useMemo(() => {
		let filtered = entries;
		if (nodeFilter !== null) {
			filtered = filtered.filter((e) => e.nodeId === nodeFilter);
		}
		if (statusFilter === 'blocked') {
			filtered = filtered.filter((e) => isBlockedStatus(e.status));
		} else if (statusFilter === 'forwarded') {
			filtered = filtered.filter((e) => isForwardedStatus(e.status));
		} else if (statusFilter === 'cached') {
			filtered = filtered.filter((e) => isCachedStatus(e.status));
		} else if (statusFilter) {
			filtered = filtered.filter((e) => e.status === statusFilter);
		}
		if (qtypeFilter) {
			filtered = filtered.filter((e) => e.qtype.toUpperCase() === qtypeFilter.toUpperCase());
		}
		return filtered;
	}, [entries, nodeFilter, statusFilter, qtypeFilter]);

	const fetchLogs = useCallback(
		async (opts: {
			timePreset: TimePreset;
			domain: string;
			clientIp: string;
			cursor?: string;
			append?: boolean;
		}) => {
			const { append, cursor: cur, ...filters } = opts;
			if (append) {
				setLoadingMore(true);
			} else {
				setLoading(true);
				setError(null);
				setNodeErrors([]);
				setEntries([]);
				setCursor('');
				setEndOfResults(false);
			}

			try {
				const preset = TIME_PRESETS.find((p) => p.value === filters.timePreset)!;
				const params: QueryLogParams = {};

				if (cur) {
					params.cursor = cur;
				} else {
					const range = presetRange(preset.minutes);
					params.from = range.from;
					params.until = range.until;
					if (filters.domain) params.domain = filters.domain;
					if (filters.clientIp) params.clientIp = filters.clientIp;
				}

				const resp = await getQueryLogs(params);
				const merged = mergeAndSort(resp);

				const errs = resp.nodes
					.filter((n) => !n.success && n.error)
					.map((n) => ({ nodeId: n.node.id, name: n.node.name, err: n.error! }));

				if (!append) {
					setAvailableNodes(
						resp.nodes.map((n) => ({ id: n.node.id, name: n.node.name })),
					);
				}

				if (append) {
					setEntries((prev) => {
						const combined = [...prev, ...merged];
						combined.sort(
							(a, b) => new Date(b.time).getTime() - new Date(a.time).getTime(),
						);
						return combined;
					});
				} else {
					setEntries(merged);
				}

				setCursor(resp.cursor);
				setEndOfResults(resp.endOfResults);
				if (errs.length > 0) setNodeErrors(errs);
			} catch (err) {
				setError(err instanceof Error ? err.message : 'Failed to load query logs');
			} finally {
				setLoading(false);
				setLoadingMore(false);
			}
		},
		[],
	);

	useEffect(() => {
		appliedFilters.current = { timePreset: INITIAL_PRESET, domain: '', clientIp: initialClientIp };
		fetchLogs({ timePreset: INITIAL_PRESET, domain: '', clientIp: initialClientIp });
	}, [fetchLogs, initialClientIp]);

	function handlePreset(preset: TimePreset) {
		setTimePreset(preset);
		appliedFilters.current = { timePreset: preset, domain, clientIp };
		fetchLogs({ timePreset: preset, domain, clientIp });
	}

	function handleSearch() {
		appliedFilters.current = { timePreset, domain, clientIp };
		fetchLogs({ timePreset, domain, clientIp });
	}

	function handleRefresh() {
		const f = appliedFilters.current;
		fetchLogs({ timePreset: f.timePreset, domain: f.domain, clientIp: f.clientIp });
	}

	function handleLoadMore() {
		if (!cursor || endOfResults || loadingMore) return;
		fetchLogs({ ...appliedFilters.current, cursor, append: true });
	}

	async function handleAction(entryDomain: string, type: 'allow' | 'deny') {
		setActioningDomain(entryDomain);
		try {
			await addDomainRule(type, 'exact', entryDomain);
			if (feedbackTimer.current) clearTimeout(feedbackTimer.current);
			setActionFeedback({ domain: entryDomain, type, ok: true });
			feedbackTimer.current = setTimeout(() => setActionFeedback(null), 3000);
		} catch {
			if (feedbackTimer.current) clearTimeout(feedbackTimer.current);
			setActionFeedback({ domain: entryDomain, type, ok: false });
			feedbackTimer.current = setTimeout(() => setActionFeedback(null), 4000);
		} finally {
			setActioningDomain(null);
		}
	}

	function toggleExpand(key: string) {
		setExpandedRows((prev) => {
			const next = new Set(prev);
			if (next.has(key)) next.delete(key);
			else next.add(key);
			return next;
		});
	}

	return (
		<div className={styles.page}>
			<div className={styles.filterBar}>
				{/* Row 1: server-side filters */}
				<div className={styles.filterRow}>
					<div className={styles.filterGroup}>
						<span className={styles.filterLabel}>Time</span>
						<div className={styles.presets} role='group' aria-label='Time range'>
							{TIME_PRESETS.map((p) => (
								<button
									key={p.value}
									type='button'
									className={styles.presetBtn}
									data-active={timePreset === p.value || undefined}
									onClick={() => handlePreset(p.value)}
									disabled={loading}
								>
									{p.label}
								</button>
							))}
						</div>
					</div>

					<div className={styles.filterGroup}>
						<label htmlFor='ql-domain' className={styles.filterLabel}>
							Domain
						</label>
						<input
							id='ql-domain'
							type='text'
							className={styles.filterInput}
							placeholder='e.g. example.com'
							value={domain}
							onChange={(e) => setDomain(e.target.value)}
							onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
							disabled={loading}
						/>
					</div>

					<div className={styles.filterGroup}>
						<label
							htmlFor='ql-client'
							className={styles.filterLabel}
							title='IP address of the device that made the DNS query, not the Pi-hole node'
						>
							Client IP
						</label>
						<input
							id='ql-client'
							type='text'
							className={styles.filterInput}
							placeholder='DNS client, e.g. 192.168.1.10'
							value={clientIp}
							onChange={(e) => setClientIp(e.target.value)}
							onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
							disabled={loading}
						/>
					</div>

					<div className={styles.filterActions}>
						<button
							type='button'
							className={styles.searchBtn}
							onClick={handleSearch}
							disabled={loading}
						>
							<Search size={15} />
							Search
						</button>
						<button
							type='button'
							className={styles.refreshBtn}
							onClick={handleRefresh}
							disabled={loading}
						>
							<RefreshCw size={15} className={loading ? styles.spin : undefined} />
							Refresh
						</button>
					</div>
				</div>

				{/* Row 2: client-side filters */}
				<div className={styles.filterRow}>
					<div className={styles.filterGroup}>
						<label htmlFor='ql-status' className={styles.filterLabel}>
							Status
						</label>
						<select
							id='ql-status'
							className={styles.filterSelect}
							value={statusFilter}
							onChange={(e) => setStatusFilter(e.target.value)}
						>
							{STATUS_OPTIONS.map((o) => (
								<option key={o.value} value={o.value}>
									{o.label}
								</option>
							))}
						</select>
					</div>

					<div className={styles.filterGroup}>
						<label htmlFor='ql-type' className={styles.filterLabel}>
							Record Type
						</label>
						<select
							id='ql-type'
							className={styles.filterSelect}
							value={qtypeFilter}
							onChange={(e) => setQtypeFilter(e.target.value)}
						>
							{QTYPE_OPTIONS.map((o) => (
								<option key={o.value} value={o.value}>
									{o.label}
								</option>
							))}
						</select>
					</div>

					{availableNodes.length > 1 && (
						<div className={styles.filterGroup}>
							<label htmlFor='ql-node' className={styles.filterLabel}>
								Node
							</label>
							<select
								id='ql-node'
								className={styles.filterSelect}
								value={nodeFilter ?? ''}
								onChange={(e) =>
									setNodeFilter(e.target.value ? Number(e.target.value) : null)
								}
							>
								<option value=''>All nodes</option>
								{availableNodes.map((n) => (
									<option key={n.id} value={n.id}>
										{n.name}
									</option>
								))}
							</select>
						</div>
					)}
				</div>
			</div>

			{nodeErrors.length > 0 && (
				<div className={styles.nodeErrors}>
					{nodeErrors.map((e) => (
						<span key={e.nodeId} className={styles.nodeError}>
							{e.name}: {e.err}
						</span>
					))}
				</div>
			)}

			{actionFeedback && (
				<div className={styles.feedback} data-ok={actionFeedback.ok || undefined}>
					{actionFeedback.ok
						? `Added "${actionFeedback.domain}" to ${actionFeedback.type === 'allow' ? 'allowlist' : 'blocklist'}`
						: `Failed to add "${actionFeedback.domain}"`}
				</div>
			)}

			{error && <div className={styles.error}>{error}</div>}

			{loading && entries.length === 0 && (
				<div className={styles.loadingState}>
					<RefreshCw size={20} className={styles.spin} />
					Loading…
				</div>
			)}

			{!loading && !error && entries.length === 0 && (
				<div className={styles.emptyState}>No log entries found for this time range.</div>
			)}

			{entries.length > 0 && (
				<>
					<div className={styles.tableInfo}>
						Showing {displayedEntries.length}
						{displayedEntries.length !== entries.length &&
							` of ${entries.length} loaded`}{' '}
						{displayedEntries.length === 1 ? 'entry' : 'entries'}
					</div>
					<div className={styles.tableWrap}>
						<table className={styles.table}>
							<thead>
								<tr>
									<th aria-label='Expand' />
									<th>Time</th>
									<th>Domain</th>
									<th>Status</th>
									<th>Type</th>
									<th title='The device that made the DNS query (IP or hostname)'>
										Client
									</th>
									<th>Node</th>
									<th aria-label='Actions' />
								</tr>
							</thead>
							<tbody>
								{displayedEntries.length === 0 ? (
									<tr>
										<td colSpan={8} className={styles.noFilterMatch}>
											No entries match the current filters.
										</td>
									</tr>
								) : (
									displayedEntries.map((entry, idx) => {
										const rowKey = `${entry.nodeId}-${entry.id}-${idx}`;
										const expanded = expandedRows.has(rowKey);
										const actioning = actioningDomain === entry.domain;
										const feedback =
											actionFeedback?.domain === entry.domain
												? actionFeedback
												: null;

										return (
											<Fragment key={rowKey}>
												<tr
													className={styles.row}
													data-expanded={expanded || undefined}
												>
													<td>
														<button
															type='button'
															className={styles.expandBtn}
															onClick={() => toggleExpand(rowKey)}
															aria-label={
																expanded ? 'Collapse' : 'Expand'
															}
															aria-expanded={expanded}
														>
															{expanded ? (
																<Minus size={14} />
															) : (
																<Plus size={14} />
															)}
														</button>
													</td>
													<td className={styles.mono}>
														{formatTime(entry.time)}
													</td>
													<td
														className={styles.domainCell}
														title={entry.domain}
													>
														{entry.domain}
													</td>
													<td>
														<span
															className={styles.statusBadge}
															style={{
																color: statusColor(entry.status),
															}}
															title={entry.status}
														>
															{shortStatus(entry.status)}
														</span>
													</td>
													<td className={styles.mono}>{entry.qtype}</td>
													<td
														className={styles.clientCell}
														title={
															entry.clientName
																? `${entry.clientName} (${entry.clientIp})`
																: entry.clientIp
														}
													>
														{entry.clientName ?? entry.clientIp}
													</td>
													<td className={styles.nodeCell}>
														{entry.nodeName}
													</td>
													<td className={styles.actionCell}>
														{feedback ? (
															<span
																className={styles.actionDone}
																data-ok={feedback.ok || undefined}
															>
																{feedback.ok
																	? `✓ ${feedback.type === 'allow' ? 'Allowed' : 'Blocked'}`
																	: '✗ Failed'}
															</span>
														) : (
															<>
																<button
																	type='button'
																	className={styles.allowBtn}
																	onClick={() =>
																		handleAction(
																			entry.domain,
																			'allow',
																		)
																	}
																	disabled={actioning}
																	title={`Add "${entry.domain}" to allowlist`}
																>
																	<Shield size={12} />
																	Allow
																</button>
																<button
																	type='button'
																	className={styles.denyBtn}
																	onClick={() =>
																		handleAction(
																			entry.domain,
																			'deny',
																		)
																	}
																	disabled={actioning}
																	title={`Add "${entry.domain}" to blocklist`}
																>
																	<ShieldOff size={12} />
																	Block
																</button>
															</>
														)}
													</td>
												</tr>
												{expanded && (
													<tr className={styles.detailRow}>
														<td colSpan={8}>
															<dl className={styles.detailGrid}>
																<div>
																	<dt>Time</dt>
																	<dd>
																		{formatDate(entry.time)}
																	</dd>
																</div>
																<div>
																	<dt>Domain</dt>
																	<dd>{entry.domain}</dd>
																</div>
																<div>
																	<dt>Status</dt>
																	<dd>{entry.status}</dd>
																</div>
																<div>
																	<dt>Query type</dt>
																	<dd>{entry.qtype}</dd>
																</div>
																<div>
																	<dt>Client IP</dt>
																	<dd>{entry.clientIp}</dd>
																</div>
																{entry.clientName && (
																	<div>
																		<dt>Client name</dt>
																		<dd>{entry.clientName}</dd>
																	</div>
																)}
																{entry.upstream && (
																	<div>
																		<dt>Upstream</dt>
																		<dd>{entry.upstream}</dd>
																	</div>
																)}
																<div>
																	<dt>Reply</dt>
																	<dd>
																		{entry.replyType} (
																		{entry.replyTimeMs}ms)
																	</dd>
																</div>
																{entry.cname && (
																	<div>
																		<dt>CNAME</dt>
																		<dd>{entry.cname}</dd>
																	</div>
																)}
																{entry.edeText && (
																	<div>
																		<dt>EDE</dt>
																		<dd>{entry.edeText}</dd>
																	</div>
																)}
																<div>
																	<dt>Node</dt>
																	<dd>{entry.nodeName}</dd>
																</div>
															</dl>
														</td>
													</tr>
												)}
											</Fragment>
										);
									})
								)}
							</tbody>
						</table>
					</div>

					<div className={styles.paginationRow}>
						{!endOfResults ? (
							<button
								type='button'
								className={styles.loadMoreBtn}
								onClick={handleLoadMore}
								disabled={loadingMore}
							>
								{loadingMore ? (
									<>
										<RefreshCw size={16} className={styles.spin} /> Loading…
									</>
								) : (
									'Load more'
								)}
							</button>
						) : (
							<span className={styles.endMsg}>End of results</span>
						)}
					</div>
				</>
			)}
		</div>
	);
}
