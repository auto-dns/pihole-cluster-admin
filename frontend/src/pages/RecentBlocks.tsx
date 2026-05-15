import { useState, useEffect, useCallback, useRef, Fragment, useMemo } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import { Search, RefreshCw, Plus, Minus, MoreHorizontal, Shield, X } from 'lucide-react';
import { getQueryLogs, type QueryLogParams } from '@/lib/api/logs';
import { listDomainRules, addDomainRule, removeDomainRule } from '@/lib/api/domainrules';
import type { MergedEntry } from '@/types/querylog';
import {
	isBlockedStatus,
	presetRange,
	formatTime,
	formatDate,
	shortStatus,
} from '@/utils/queryLogHelpers';
import styles from './RecentBlocks.module.scss';

type TimePreset = '1h' | '6h' | '24h' | '7d';

const INITIAL_PRESET: TimePreset = '1h';

const TIME_PRESETS: { label: string; value: TimePreset; minutes: number }[] = [
	{ label: '1h', value: '1h', minutes: 60 },
	{ label: '6h', value: '6h', minutes: 360 },
	{ label: '24h', value: '24h', minutes: 1440 },
	{ label: '7d', value: '7d', minutes: 10080 },
];

type GroupPending = {
	allowPending: boolean;
	allowFeedback: 'ok' | 'error' | null;
	removePending: boolean;
	removeFeedback: 'ok' | 'error' | null;
};

type GroupedBase = {
	key: string;
	domain: string;
	qtype: string;
	count: number;
	lastBlockedAt: string;
	entries: MergedEntry[];
};

type DisplayGroup = GroupedBase & {
	isAllowed: boolean;
	allowedKind: 'exact' | 'regex' | null;
} & GroupPending;

const DEFAULT_PENDING: GroupPending = {
	allowPending: false,
	allowFeedback: null,
	removePending: false,
	removeFeedback: null,
};

export function RecentBlocks() {
	const [preset, setPreset] = useState<TimePreset>(INITIAL_PRESET);
	const [domainSearch, setDomainSearch] = useState('');
	const [clientIpSearch, setClientIpSearch] = useState('');
	const appliedFilters = useRef({ preset: INITIAL_PRESET, domain: '', clientIp: '' });

	const [entries, setEntries] = useState<MergedEntry[]>([]);
	const [cursor, setCursor] = useState('');
	const [endOfResults, setEndOfResults] = useState(false);
	const [loading, setLoading] = useState(false);
	const [loadingMore, setLoadingMore] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [nodeErrors, setNodeErrors] = useState<{ nodeId: number; name: string; err: string }[]>(
		[],
	);

	const [expandedKeys, setExpandedKeys] = useState<Set<string>>(new Set());
	// domain → 'exact' | 'regex': all domains with an active allow rule (pre-loaded + session-added)
	const [allowedDomains, setAllowedDomains] = useState<Map<string, 'exact' | 'regex'>>(new Map());
	// group.key → pending/feedback state for in-progress actions
	const [groupPending, setGroupPending] = useState<Map<string, GroupPending>>(new Map());
	const feedbackTimers = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

	const [regexModal, setRegexModal] = useState<{
		open: boolean;
		domain: string;
		groupKey: string;
		pattern: string;
		submitting: boolean;
		error: string | null;
	}>({ open: false, domain: '', groupKey: '', pattern: '', submitting: false, error: null });

	const groups = useMemo<GroupedBase[]>(() => {
		const map = new Map<string, GroupedBase>();
		for (const entry of entries) {
			if (!isBlockedStatus(entry.status)) continue;
			const key = `${entry.domain}::${entry.qtype}`;
			let g = map.get(key);
			if (!g) {
				g = {
					key,
					domain: entry.domain,
					qtype: entry.qtype,
					count: 0,
					lastBlockedAt: entry.time,
					entries: [],
				};
				map.set(key, g);
			}
			g.count++;
			g.entries.push(entry);
			if (new Date(entry.time) > new Date(g.lastBlockedAt)) g.lastBlockedAt = entry.time;
		}
		return Array.from(map.values()).sort(
			(a, b) => new Date(b.lastBlockedAt).getTime() - new Date(a.lastBlockedAt).getTime(),
		);
	}, [entries]);

	const displayGroups = useMemo<DisplayGroup[]>(() => {
		return groups.map((g) => {
			const kind = allowedDomains.get(g.domain) ?? null;
			const pending = groupPending.get(g.key) ?? DEFAULT_PENDING;
			return { ...g, isAllowed: kind !== null, allowedKind: kind, ...pending };
		});
	}, [groups, allowedDomains, groupPending]);

	const totalBlockedEntries = useMemo(
		() => displayGroups.reduce((s, g) => s + g.count, 0),
		[displayGroups],
	);

	const fetchLogs = useCallback(
		async (opts: {
			preset: TimePreset;
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
				const presetDef = TIME_PRESETS.find((p) => p.value === filters.preset)!;
				const params: QueryLogParams = { length: 500 };
				if (cur) {
					params.cursor = cur;
				} else {
					const range = presetRange(presetDef.minutes);
					params.from = range.from;
					params.until = range.until;
					if (filters.domain) params.domain = filters.domain;
					if (filters.clientIp) params.clientIp = filters.clientIp;
				}

				const resp = await getQueryLogs(params);
				const merged: MergedEntry[] = [];
				for (const n of resp.nodes) {
					if (n.success && n.page) {
						for (const e of n.page.entries) {
							merged.push({ ...e, nodeId: n.node.id, nodeName: n.node.name });
						}
					}
				}
				merged.sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime());

				const errs = resp.nodes
					.filter((n) => !n.success && n.error)
					.map((n) => ({ nodeId: n.node.id, name: n.node.name, err: n.error! }));

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

	const loadAllowRules = useCallback(async () => {
		try {
			const resp = await listDomainRules();
			const allowed = new Map<string, 'exact' | 'regex'>();
			for (const nodeResult of Object.values(resp.nodes)) {
				for (const rule of nodeResult.rules) {
					if (rule.type === 'allow' && !allowed.has(rule.domain)) {
						allowed.set(rule.domain, rule.kind as 'exact' | 'regex');
					}
				}
			}
			setAllowedDomains(allowed);
		} catch {
			// non-fatal: page still works without pre-loaded allow state
		}
	}, []);

	useEffect(() => {
		appliedFilters.current = { preset: INITIAL_PRESET, domain: '', clientIp: '' };
		fetchLogs({ preset: INITIAL_PRESET, domain: '', clientIp: '' });
		loadAllowRules();
		return () => {
			for (const t of feedbackTimers.current.values()) clearTimeout(t);
		};
	}, [fetchLogs, loadAllowRules]);

	function updatePending(groupKey: string, patch: Partial<GroupPending>) {
		setGroupPending((prev) => {
			const m = new Map(prev);
			m.set(groupKey, { ...(prev.get(groupKey) ?? DEFAULT_PENDING), ...patch });
			return m;
		});
	}

	function scheduleFeedbackClear(groupKey: string, field: 'allowFeedback' | 'removeFeedback') {
		const timerKey = `${groupKey}:${field}`;
		const existing = feedbackTimers.current.get(timerKey);
		if (existing) clearTimeout(existing);
		const t = setTimeout(() => {
			setGroupPending((prev) => {
				const m = new Map(prev);
				const cur = prev.get(groupKey);
				if (cur) m.set(groupKey, { ...cur, [field]: null });
				return m;
			});
			feedbackTimers.current.delete(timerKey);
		}, 3000);
		feedbackTimers.current.set(timerKey, t);
	}

	async function handleAllow(groupKey: string, domain: string) {
		updatePending(groupKey, { allowPending: true });
		try {
			await addDomainRule('allow', 'exact', domain);
			setAllowedDomains((prev) => new Map([...prev, [domain, 'exact']]));
			updatePending(groupKey, { allowPending: false, allowFeedback: 'ok' });
			scheduleFeedbackClear(groupKey, 'allowFeedback');
		} catch {
			updatePending(groupKey, { allowPending: false, allowFeedback: 'error' });
			scheduleFeedbackClear(groupKey, 'allowFeedback');
		}
	}

	async function handleRemove(groupKey: string, domain: string) {
		updatePending(groupKey, { removePending: true });
		try {
			await removeDomainRule('allow', 'exact', domain);
			setAllowedDomains((prev) => {
				const m = new Map(prev);
				m.delete(domain);
				return m;
			});
			updatePending(groupKey, { removePending: false, removeFeedback: 'ok' });
			scheduleFeedbackClear(groupKey, 'removeFeedback');
		} catch {
			updatePending(groupKey, { removePending: false, removeFeedback: 'error' });
			scheduleFeedbackClear(groupKey, 'removeFeedback');
		}
	}

	async function handleRegexAllow() {
		const { domain, groupKey, pattern } = regexModal;
		if (!pattern.trim()) return;
		setRegexModal((m) => ({ ...m, submitting: true, error: null }));
		try {
			await addDomainRule('allow', 'regex', pattern.trim());
			setAllowedDomains((prev) => new Map([...prev, [domain, 'regex']]));
			setRegexModal((m) => ({ ...m, open: false, submitting: false }));
			updatePending(groupKey, { allowFeedback: 'ok' });
			scheduleFeedbackClear(groupKey, 'allowFeedback');
		} catch (err) {
			setRegexModal((m) => ({
				...m,
				submitting: false,
				error: err instanceof Error ? err.message : 'Failed to add rule',
			}));
		}
	}

	function openRegexModal(domain: string, groupKey: string) {
		setRegexModal({
			open: true,
			domain,
			groupKey,
			pattern: domain,
			submitting: false,
			error: null,
		});
	}

	function handlePreset(p: TimePreset) {
		setPreset(p);
		appliedFilters.current = { preset: p, domain: domainSearch, clientIp: clientIpSearch };
		fetchLogs({ preset: p, domain: domainSearch, clientIp: clientIpSearch });
	}

	function handleSearch() {
		appliedFilters.current = { preset, domain: domainSearch, clientIp: clientIpSearch };
		fetchLogs({ preset, domain: domainSearch, clientIp: clientIpSearch });
	}

	function handleRefresh() {
		const f = appliedFilters.current;
		fetchLogs({ preset: f.preset, domain: f.domain, clientIp: f.clientIp });
	}

	function handleLoadMore() {
		if (!cursor || endOfResults || loadingMore) return;
		fetchLogs({ ...appliedFilters.current, cursor, append: true });
	}

	function toggleExpand(key: string) {
		setExpandedKeys((prev) => {
			const next = new Set(prev);
			if (next.has(key)) next.delete(key);
			else next.add(key);
			return next;
		});
	}

	return (
		<div className={styles.page}>
			<div className={styles.filterBar}>
				<div className={styles.filterRow}>
					<div className={styles.filterGroup}>
						<span className={styles.filterLabel}>Time</span>
						<div className={styles.presets} role='group' aria-label='Time range'>
							{TIME_PRESETS.map((p) => (
								<button
									key={p.value}
									type='button'
									className={styles.presetBtn}
									data-active={preset === p.value || undefined}
									onClick={() => handlePreset(p.value)}
									disabled={loading}
								>
									{p.label}
								</button>
							))}
						</div>
					</div>

					<div className={styles.filterGroup}>
						<label htmlFor='rb-domain' className={styles.filterLabel}>
							Domain
						</label>
						<input
							id='rb-domain'
							type='text'
							className={styles.filterInput}
							placeholder='e.g. example.com'
							value={domainSearch}
							onChange={(e) => setDomainSearch(e.target.value)}
							onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
							disabled={loading}
						/>
					</div>

					<div className={styles.filterGroup}>
						<label
							htmlFor='rb-client'
							className={styles.filterLabel}
							title='IP address of the device that made the DNS query'
						>
							Client IP
						</label>
						<input
							id='rb-client'
							type='text'
							className={styles.filterInput}
							placeholder='e.g. 192.168.1.10'
							value={clientIpSearch}
							onChange={(e) => setClientIpSearch(e.target.value)}
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

			{!loading && !error && entries.length > 0 && displayGroups.length === 0 && (
				<div className={styles.emptyState}>
					No blocked entries found in this time window.
				</div>
			)}

			{displayGroups.length > 0 && (
				<>
					<div className={styles.tableInfo}>
						{displayGroups.length}{' '}
						{displayGroups.length === 1 ? 'blocked domain' : 'blocked domains'}
						{' · '}
						{totalBlockedEntries} {totalBlockedEntries === 1 ? 'entry' : 'entries'}{' '}
						total
					</div>
					<div className={styles.tableWrap}>
						<table className={styles.table}>
							<thead>
								<tr>
									<th aria-label='Expand' />
									<th>Domain</th>
									<th>Type</th>
									<th>Count</th>
									<th>Last Blocked</th>
									<th aria-label='Allowed status' />
									<th aria-label='Actions' />
								</tr>
							</thead>
							<tbody>
								{displayGroups.map((group) => {
									const expanded = expandedKeys.has(group.key);
									const isPending = group.allowPending || group.removePending;

									return (
										<Fragment key={group.key}>
											<tr
												className={styles.groupRow}
												data-expanded={expanded || undefined}
												data-allowed={group.isAllowed || undefined}
											>
												<td>
													<button
														type='button'
														className={styles.expandBtn}
														onClick={() => toggleExpand(group.key)}
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
												<td
													className={styles.domainCell}
													title={group.domain}
												>
													{group.domain}
												</td>
												<td>
													<span className={styles.qtypeBadge}>
														{group.qtype}
													</span>
												</td>
												<td className={styles.countCell}>{group.count}</td>
												<td
													className={styles.timeCell}
													title={formatDate(group.lastBlockedAt)}
												>
													{formatTime(group.lastBlockedAt)}
												</td>
												<td>
													{group.isAllowed && (
														<span className={styles.allowedBadge}>
															✓ Allowed
														</span>
													)}
												</td>
												<td className={styles.actionCell}>
													{group.allowFeedback === 'ok' && (
														<span className={styles.actionDone}>
															✓ Allowed
														</span>
													)}
													{group.allowFeedback === 'error' && (
														<span className={styles.actionError}>
															✗ Failed
														</span>
													)}
													{group.removeFeedback === 'ok' && (
														<span className={styles.actionDone}>
															✓ Removed
														</span>
													)}
													{group.removeFeedback === 'error' && (
														<span className={styles.actionError}>
															✗ Failed
														</span>
													)}
													{!group.allowFeedback &&
														!group.removeFeedback &&
														!group.isAllowed && (
															<>
																<button
																	type='button'
																	className={styles.allowBtn}
																	onClick={() =>
																		handleAllow(
																			group.key,
																			group.domain,
																		)
																	}
																	disabled={isPending}
																	title={`Add "${group.domain}" to allowlist`}
																>
																	{group.allowPending ? (
																		<RefreshCw
																			size={12}
																			className={styles.spin}
																		/>
																	) : (
																		<Shield size={12} />
																	)}
																	Allow
																</button>
																<DropdownMenu.Root>
																	<DropdownMenu.Trigger asChild>
																		<button
																			type='button'
																			className={
																				styles.menuBtn
																			}
																			aria-label='More options'
																			disabled={isPending}
																		>
																			<MoreHorizontal
																				size={14}
																			/>
																		</button>
																	</DropdownMenu.Trigger>
																	<DropdownMenu.Portal>
																		<DropdownMenu.Content
																			className={
																				styles.dropdownContent
																			}
																			align='end'
																			sideOffset={4}
																		>
																			<DropdownMenu.Item
																				className={
																					styles.dropdownItem
																				}
																				onSelect={() =>
																					openRegexModal(
																						group.domain,
																						group.key,
																					)
																				}
																			>
																				Allow (Regex)
																			</DropdownMenu.Item>
																		</DropdownMenu.Content>
																	</DropdownMenu.Portal>
																</DropdownMenu.Root>
															</>
														)}
													{!group.allowFeedback &&
														!group.removeFeedback &&
														group.isAllowed &&
														group.allowedKind === 'exact' && (
															<button
																type='button'
																className={styles.removeBtn}
																onClick={() =>
																	handleRemove(
																		group.key,
																		group.domain,
																	)
																}
																disabled={isPending}
																title={`Remove "${group.domain}" from allowlist`}
															>
																{group.removePending && (
																	<RefreshCw
																		size={12}
																		className={styles.spin}
																	/>
																)}
																Remove
															</button>
														)}
													{!group.allowFeedback &&
														!group.removeFeedback &&
														group.isAllowed &&
														group.allowedKind === 'regex' && (
															<span className={styles.actionNote}>
																Regex rule
															</span>
														)}
												</td>
											</tr>
											{expanded && (
												<tr className={styles.detailRow}>
													<td colSpan={7}>
														{group.isAllowed && (
															<div className={styles.allowedNotice}>
																Allowlisted — entries below were
																blocked before the rule was added
															</div>
														)}
														<div className={styles.detailEntries}>
															{group.entries.map((entry, idx) => (
																<div
																	key={`${entry.nodeId}-${entry.id}-${idx}`}
																	className={styles.detailEntry}
																>
																	<span className={styles.mono}>
																		{formatDate(entry.time)}
																	</span>
																	<span
																		className={styles.nodeTag}
																	>
																		{entry.nodeName}
																	</span>
																	<span
																		className={styles.clientTag}
																	>
																		{entry.clientName ??
																			entry.clientIp}
																	</span>
																	<span
																		className={
																			styles.statusDetail
																		}
																		title={entry.status}
																	>
																		{shortStatus(entry.status)}
																	</span>
																</div>
															))}
														</div>
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

			{/* Regex allow modal */}
			<Dialog.Root
				open={regexModal.open}
				onOpenChange={(open) => {
					if (!open) setRegexModal((m) => ({ ...m, open: false, error: null }));
				}}
			>
				<Dialog.Portal>
					<Dialog.Overlay className={styles.overlay} />
					<Dialog.Content className={styles.dialog}>
						<Dialog.Title className={styles.dialogTitle}>Allow (Regex)</Dialog.Title>
						<p className={styles.dialogHint}>
							Pattern will be applied as a regex allow rule across the cluster.
						</p>
						<div className={styles.field}>
							<label htmlFor='rb-regex-pattern' className={styles.fieldLabel}>
								Pattern
							</label>
							<input
								id='rb-regex-pattern'
								type='text'
								className={styles.input}
								value={regexModal.pattern}
								onChange={(e) =>
									setRegexModal((m) => ({ ...m, pattern: e.target.value }))
								}
								onKeyDown={(e) => e.key === 'Enter' && handleRegexAllow()}
								disabled={regexModal.submitting}
								autoFocus
							/>
						</div>
						{regexModal.error && (
							<p className={styles.dialogError}>{regexModal.error}</p>
						)}
						<div className={styles.dialogActions}>
							<Dialog.Close asChild>
								<button
									type='button'
									className={styles.cancelBtn}
									disabled={regexModal.submitting}
								>
									Cancel
								</button>
							</Dialog.Close>
							<button
								type='button'
								className={styles.submitBtn}
								onClick={handleRegexAllow}
								disabled={regexModal.submitting || !regexModal.pattern.trim()}
							>
								{regexModal.submitting ? (
									<RefreshCw size={15} className={styles.spin} />
								) : null}
								Allow
							</button>
						</div>
						<Dialog.Close asChild>
							<button className={styles.dialogClose} aria-label='Close'>
								<X size={18} />
							</button>
						</Dialog.Close>
					</Dialog.Content>
				</Dialog.Portal>
			</Dialog.Root>
		</div>
	);
}
