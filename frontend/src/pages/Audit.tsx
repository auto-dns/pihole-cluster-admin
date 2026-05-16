import { useState, useEffect, useCallback } from 'react';
import { RefreshCw, CheckCircle, XCircle, ChevronDown, ChevronRight, RotateCcw } from 'lucide-react';
import { listAuditEntries, rollbackAuditEntry } from '@/lib/api/audit';
import type { AuditEntry, AuditAction, RollbackNodeResult } from '@/types/audit';
import styles from './Audit.module.scss';

const ACTION_LABELS: Record<AuditAction, string> = {
	add_domain_rule: 'Added domain rule',
	remove_domain_rule: 'Removed domain rule',
	set_cluster_blocking: 'Set cluster blocking',
	set_node_blocking: 'Set node blocking',
};

function actionLabel(action: AuditAction) {
	return ACTION_LABELS[action] ?? action;
}

function entryDescription(entry: AuditEntry): string {
	switch (entry.action) {
		case 'add_domain_rule':
		case 'remove_domain_rule': {
			const typeLabel = entry.ruleType === 'allow' ? 'allowlist' : 'blocklist';
			const kindLabel = entry.ruleKind === 'regex' ? ' (regex)' : '';
			return `${entry.targetDomain ?? '—'} → ${typeLabel}${kindLabel}`;
		}
		case 'set_cluster_blocking':
			return entry.blockingEnabled
				? `Enabled${entry.blockingTimer ? ` (${entry.blockingTimer}s timer)` : ''}`
				: 'Disabled';
		case 'set_node_blocking':
			return `${entry.targetNodeName ?? `node ${entry.targetNodeId}`}: ${
				entry.blockingEnabled
					? `enabled${entry.blockingTimer ? ` (${entry.blockingTimer}s)` : ''}`
					: 'disabled'
			}`;
		default:
			return '—';
	}
}

function formatTime(iso: string): string {
	const d = new Date(iso);
	return d.toLocaleString(undefined, {
		month: 'short',
		day: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
		second: '2-digit',
	});
}

const PAGE_SIZE = 50;

const ROLLBACK_ACTIONS = new Set<AuditAction>(['add_domain_rule', 'remove_domain_rule']);

type RollbackState =
	| { status: 'idle' }
	| { status: 'confirming' }
	| { status: 'loading' }
	| { status: 'done'; nodes: RollbackNodeResult[] }
	| { status: 'error'; message: string };

function AuditRow({ entry }: { entry: AuditEntry }) {
	const [expanded, setExpanded] = useState(false);
	const [rollback, setRollback] = useState<RollbackState>({ status: 'idle' });

	const hasResults = entry.nodeResults.length > 0;
	const allSuccess = entry.nodeResults.every((r) => r.success);
	const anyFailure = entry.nodeResults.some((r) => !r.success);
	const canRollback = ROLLBACK_ACTIONS.has(entry.action);

	async function handleRollback() {
		if (rollback.status === 'confirming') {
			setRollback({ status: 'loading' });
			try {
				const result = await rollbackAuditEntry(entry.id);
				setRollback({ status: 'done', nodes: result.nodes });
			} catch (err) {
				setRollback({
					status: 'error',
					message: err instanceof Error ? err.message : 'Rollback failed',
				});
			}
		} else {
			setRollback({ status: 'confirming' });
		}
	}

	function cancelRollback() {
		setRollback({ status: 'idle' });
	}

	const colSpan = 7;

	return (
		<>
			<tr
				className={styles.row}
				data-expandable={hasResults || undefined}
				onClick={() => hasResults && rollback.status === 'idle' && setExpanded((v) => !v)}
			>
				<td className={styles.chevronCell}>
					{hasResults ? (
						expanded ? (
							<ChevronDown size={14} className={styles.chevron} />
						) : (
							<ChevronRight size={14} className={styles.chevron} />
						)
					) : null}
				</td>
				<td className={styles.timeCell}>{formatTime(entry.createdAt)}</td>
				<td>
					<span className={styles.actionBadge} data-action={entry.action}>
						{actionLabel(entry.action)}
					</span>
				</td>
				<td className={styles.descCell}>{entryDescription(entry)}</td>
				<td className={styles.actorCell}>{entry.actor}</td>
				<td className={styles.statusCell}>
					{anyFailure ? (
						<XCircle size={15} className={styles.iconFail} />
					) : (
						<CheckCircle size={15} className={styles.iconOk} />
					)}
					{!allSuccess && anyFailure && (
						<span className={styles.statusText}>
							{entry.nodeResults.filter((r) => r.success).length}/
							{entry.nodeResults.length}
						</span>
					)}
				</td>
				<td className={styles.undoCell} onClick={(e) => e.stopPropagation()}>
					{canRollback && rollback.status === 'idle' && (
						<button
							type='button'
							className={styles.undoBtn}
							onClick={handleRollback}
							title='Undo this action'
						>
							<RotateCcw size={13} />
						</button>
					)}
					{canRollback && rollback.status === 'confirming' && (
						<span className={styles.undoConfirm}>
							Undo?{' '}
							<button type='button' className={styles.undoYes} onClick={handleRollback}>
								Yes
							</button>{' '}
							<button type='button' className={styles.undoNo} onClick={cancelRollback}>
								No
							</button>
						</span>
					)}
					{rollback.status === 'loading' && (
						<RefreshCw size={13} className={styles.spin} />
					)}
				</td>
			</tr>
			{(expanded && hasResults) || rollback.status === 'done' || rollback.status === 'error' ? (
				<tr className={styles.expandedRow}>
					<td colSpan={colSpan}>
						{rollback.status === 'done' && (
							<div className={styles.rollbackResults}>
								<span className={styles.rollbackLabel}>Undo result:</span>
								{rollback.nodes.map((nr) => (
									<div key={nr.nodeId} className={styles.nodeResult}>
										{nr.success ? (
											<CheckCircle size={13} className={styles.iconOk} />
										) : (
											<XCircle size={13} className={styles.iconFail} />
										)}
										<span className={styles.nodeName}>{nr.nodeName}</span>
										{nr.error && (
											<span className={styles.nodeError}>{nr.error}</span>
										)}
									</div>
								))}
							</div>
						)}
						{rollback.status === 'error' && (
							<div className={styles.rollbackError}>{rollback.message}</div>
						)}
						{rollback.status !== 'done' && rollback.status !== 'error' && expanded && hasResults && (
							<div className={styles.nodeResults}>
								{entry.nodeResults.map((nr) => (
									<div key={nr.nodeId} className={styles.nodeResult}>
										{nr.success ? (
											<CheckCircle size={13} className={styles.iconOk} />
										) : (
											<XCircle size={13} className={styles.iconFail} />
										)}
										<span className={styles.nodeName}>{nr.nodeName}</span>
										{nr.error && (
											<span className={styles.nodeError}>{nr.error}</span>
										)}
									</div>
								))}
							</div>
						)}
					</td>
				</tr>
			) : null}
		</>
	);
}

export function Audit() {
	const [entries, setEntries] = useState<AuditEntry[]>([]);
	const [total, setTotal] = useState(0);
	const [offset, setOffset] = useState(0);
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const fetchPage = useCallback(async (pageOffset: number) => {
		setLoading(true);
		setError(null);
		try {
			const resp = await listAuditEntries(PAGE_SIZE, pageOffset);
			setEntries(resp.entries ?? []);
			setTotal(resp.total);
			setOffset(pageOffset);
		} catch (err) {
			setError(err instanceof Error ? err.message : 'Failed to load audit log');
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		fetchPage(0);
	}, [fetchPage]);

	const totalPages = Math.ceil(total / PAGE_SIZE);
	const currentPage = Math.floor(offset / PAGE_SIZE) + 1;

	return (
		<div className={styles.page}>
			<div className={styles.toolbar}>
				<span className={styles.totalCount}>
					{total} {total === 1 ? 'entry' : 'entries'}
				</span>
				<button
					type='button'
					className={styles.refreshBtn}
					onClick={() => fetchPage(offset)}
					disabled={loading}
					aria-label='Refresh'
				>
					<RefreshCw size={16} className={loading ? styles.spin : undefined} />
				</button>
			</div>

			{error && <div className={styles.error}>{error}</div>}

			{loading && entries.length === 0 && (
				<div className={styles.loadingState}>
					<RefreshCw size={20} className={styles.spin} />
					Loading…
				</div>
			)}

			{!loading && !error && entries.length === 0 && (
				<div className={styles.emptyState}>No audit log entries yet.</div>
			)}

			{entries.length > 0 && (
				<div className={styles.tableWrap}>
					<table className={styles.table}>
						<thead>
							<tr>
								<th aria-label='Expand' />
								<th>Time</th>
								<th>Action</th>
								<th>Detail</th>
								<th>Actor</th>
								<th>Result</th>
								<th aria-label='Undo' />
							</tr>
						</thead>
						<tbody>
							{entries.map((e) => (
								<AuditRow key={e.id} entry={e} />
							))}
						</tbody>
					</table>
				</div>
			)}

			{totalPages > 1 && (
				<div className={styles.pagination}>
					<button
						type='button'
						className={styles.pageBtn}
						disabled={offset === 0 || loading}
						onClick={() => fetchPage(Math.max(0, offset - PAGE_SIZE))}
					>
						Previous
					</button>
					<span className={styles.pageInfo}>
						{currentPage} / {totalPages}
					</span>
					<button
						type='button'
						className={styles.pageBtn}
						disabled={offset + PAGE_SIZE >= total || loading}
						onClick={() => fetchPage(offset + PAGE_SIZE)}
					>
						Next
					</button>
				</div>
			)}
		</div>
	);
}
