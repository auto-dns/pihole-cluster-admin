import { useState, useEffect, useCallback, useRef, Fragment } from 'react';
import { Search, RefreshCw, Shield, ShieldOff, ChevronDown, ChevronUp } from 'lucide-react';
import { getQueryLogs, type QueryLogParams } from '@/lib/api/logs';
import { addDomainRule } from '@/lib/api/domainrules';
import type { MergedEntry, QueryLogResponse } from '@/types/querylog';
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

function presetRange(minutes: number): { from: string; until: string } {
	const until = new Date();
	const from = new Date(until.getTime() - minutes * 60_000);
	return { from: from.toISOString(), until: until.toISOString() };
}

function formatTime(iso: string): string {
	return new Date(iso).toLocaleTimeString();
}

function formatDate(iso: string): string {
	const d = new Date(iso);
	return `${d.toLocaleDateString()} ${d.toLocaleTimeString()}`;
}

function statusColor(status: string): string {
	if (status.startsWith('BLOCKED')) return 'var(--accent-danger)';
	if (status.startsWith('OK')) return 'var(--accent-success)';
	return 'var(--text-secondary)';
}

function shortStatus(status: string): string {
	if (status === 'BLOCKED_GRAVITY') return 'Gravity';
	if (status === 'BLOCKED_REGEX') return 'Regex';
	if (status === 'BLOCKED_BLACKLIST') return 'Exact';
	if (status === 'BLOCKED_EXTERNAL_IP') return 'Ext.IP';
	if (status === 'BLOCKED_EXTERNAL_NXDOMAIN') return 'Ext.NX';
	if (status === 'BLOCKED_EXTERNAL_REFUSED') return 'Ext.RF';
	if (status === 'OK_FORWARDED') return 'Forwarded';
	if (status === 'OK_CACHE') return 'Cached';
	if (status === 'OK_RETRIED') return 'Retried';
	if (status === 'SPECIAL_DOMAIN') return 'Special';
	if (status.startsWith('BLOCKED')) return 'Blocked';
	if (status.startsWith('OK')) return 'OK';
	return status;
}

function mergeAndSort(resp: QueryLogResponse): MergedEntry[] {
	const out: MergedEntry[] = [];
	for (const n of resp.nodes) {
		if (n.success && n.page) {
			for (const e of n.page.entries) {
				out.push({ ...e, nodeId: n.node.id, nodeName: n.node.name });
			}
		}
	}
	out.sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime());
	return out;
}

export function QueryLogs() {
	const [timePreset, setTimePreset] = useState<TimePreset>('5m');
	const [domain, setDomain] = useState('');
	const [clientIp, setClientIp] = useState('');

	const [entries, setEntries] = useState<MergedEntry[]>([]);
	const [cursor, setCursor] = useState('');
	const [endOfResults, setEndOfResults] = useState(false);
	const [loading, setLoading] = useState(false);
	const [loadingMore, setLoadingMore] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [nodeErrors, setNodeErrors] = useState<{ nodeId: number; name: string; err: string }[]>([]);
	const [totalShown, setTotalShown] = useState(0);

	const [actioningDomain, setActioningDomain] = useState<string | null>(null);
	const [actionFeedback, setActionFeedback] = useState<{ domain: string; type: 'allow' | 'deny'; ok: boolean } | null>(null);
	const feedbackTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

	const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

	const appliedFilters = useRef({ timePreset, domain: '', clientIp: '' });

	const fetchLogs = useCallback(async (opts: {
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
			setTotalShown(0);
		}

		try {
			const preset = TIME_PRESETS.find(p => p.value === filters.timePreset)!;
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
				.filter(n => !n.success && n.error)
				.map(n => ({ nodeId: n.node.id, name: n.node.name, err: n.error! }));

			if (append) {
				setEntries(prev => {
					const combined = [...prev, ...merged];
					combined.sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime());
					return combined;
				});
				setTotalShown(prev => prev + merged.length);
			} else {
				setEntries(merged);
				setTotalShown(merged.length);
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
	}, []);

	useEffect(() => {
		appliedFilters.current = { timePreset: INITIAL_PRESET, domain: '', clientIp: '' };
		fetchLogs({ timePreset: INITIAL_PRESET, domain: '', clientIp: '' });
	}, [fetchLogs]);

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
		setExpandedRows(prev => {
			const next = new Set(prev);
			if (next.has(key)) next.delete(key);
			else next.add(key);
			return next;
		});
	}

	return (
		<div className={styles.page}>
			<div className={styles.filterBar}>
				<div className={styles.filterGroup}>
					<span className={styles.filterLabel}>Time</span>
					<div className={styles.presets} role="group" aria-label="Time range">
						{TIME_PRESETS.map(p => (
							<button
								key={p.value}
								type="button"
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
					<label htmlFor="ql-domain" className={styles.filterLabel}>Domain</label>
					<input
						id="ql-domain"
						type="text"
						className={styles.filterInput}
						placeholder="e.g. example.com"
						value={domain}
						onChange={e => setDomain(e.target.value)}
						onKeyDown={e => e.key === 'Enter' && handleSearch()}
						disabled={loading}
					/>
				</div>

				<div className={styles.filterGroup}>
					<label htmlFor="ql-client" className={styles.filterLabel}>Client IP</label>
					<input
						id="ql-client"
						type="text"
						className={styles.filterInput}
						placeholder="e.g. 192.168.1.10"
						value={clientIp}
						onChange={e => setClientIp(e.target.value)}
						onKeyDown={e => e.key === 'Enter' && handleSearch()}
						disabled={loading}
					/>
				</div>

				<div className={styles.filterActions}>
					<button type="button" className={styles.searchBtn} onClick={handleSearch} disabled={loading} aria-label="Search">
						<Search size={16} />
						Search
					</button>
					<button type="button" className={styles.refreshBtn} onClick={handleRefresh} disabled={loading} aria-label="Refresh">
						<RefreshCw size={16} className={loading ? styles.spin : undefined} />
					</button>
				</div>
			</div>

			{nodeErrors.length > 0 && (
				<div className={styles.nodeErrors}>
					{nodeErrors.map(e => (
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
						Showing {totalShown} {totalShown === 1 ? 'entry' : 'entries'}
					</div>
					<div className={styles.tableWrap}>
						<table className={styles.table}>
							<thead>
								<tr>
									<th aria-label="Expand" />
									<th>Time</th>
									<th>Domain</th>
									<th>Status</th>
									<th>Type</th>
									<th>Client</th>
									<th>Node</th>
									<th aria-label="Actions" />
								</tr>
							</thead>
							<tbody>
								{entries.map((entry, idx) => {
									const rowKey = `${entry.nodeId}-${entry.id}-${idx}`;
									const expanded = expandedRows.has(rowKey);
									const actioning = actioningDomain === entry.domain;
									const feedback = actionFeedback?.domain === entry.domain ? actionFeedback : null;

									return (
										<Fragment key={rowKey}>
											<tr className={styles.row} data-expanded={expanded || undefined}>
												<td>
													<button
														type="button"
														className={styles.expandBtn}
														onClick={() => toggleExpand(rowKey)}
														aria-label={expanded ? 'Collapse' : 'Expand'}
														aria-expanded={expanded}
													>
														{expanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
													</button>
												</td>
												<td className={styles.mono}>{formatTime(entry.time)}</td>
												<td className={styles.domainCell} title={entry.domain}>{entry.domain}</td>
												<td>
													<span
														className={styles.statusBadge}
														style={{ color: statusColor(entry.status) }}
														title={entry.status}
													>
														{shortStatus(entry.status)}
													</span>
												</td>
												<td className={styles.mono}>{entry.qtype}</td>
												<td className={styles.clientCell} title={entry.clientName ?? entry.clientIp}>
													{entry.clientName ?? entry.clientIp}
												</td>
												<td className={styles.nodeCell}>{entry.nodeName}</td>
												<td className={styles.actionCell}>
													{feedback ? (
														<span className={styles.actionDone} data-ok={feedback.ok || undefined}>
															{feedback.ok ? '✓' : '✗'}
														</span>
													) : (
														<>
															<button
																type="button"
																className={styles.allowBtn}
																onClick={() => handleAction(entry.domain, 'allow')}
																disabled={actioning}
																aria-label={`Allow ${entry.domain}`}
																title="Add to allowlist"
															>
																<Shield size={13} />
															</button>
															<button
																type="button"
																className={styles.denyBtn}
																onClick={() => handleAction(entry.domain, 'deny')}
																disabled={actioning}
																aria-label={`Block ${entry.domain}`}
																title="Add to blocklist"
															>
																<ShieldOff size={13} />
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
																<dd>{formatDate(entry.time)}</dd>
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
																<dd>{entry.replyType} ({entry.replyTimeMs}ms)</dd>
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
								})}
							</tbody>
						</table>
					</div>

					<div className={styles.paginationRow}>
						{!endOfResults ? (
							<button
								type="button"
								className={styles.loadMoreBtn}
								onClick={handleLoadMore}
								disabled={loadingMore}
							>
								{loadingMore ? (
									<><RefreshCw size={16} className={styles.spin} /> Loading…</>
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
