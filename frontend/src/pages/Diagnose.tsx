import { useState, useEffect, useRef, useCallback } from 'react';
import { Circle, RefreshCw, Shield, Square, Timer } from 'lucide-react';
import { addDomainRule, removeDomainRule } from '@/lib/api/domainrules';
import classNames from 'classnames';
import styles from './Diagnose.module.scss';

const DEFAULT_DURATION_S = 60;
const EXTEND_S = 30;

type BlockedBy = 'rule' | 'gravity';

type ActionState = 'idle' | 'loading' | 'ok' | 'error';

type BlockedEntry = {
	domain: string;
	client: string;
	node: string;
	status: string;
	timestamp: string;
	blockedBy: BlockedBy;
	actionState: ActionState;
};

type RawEvent = {
	domain: string;
	client: string;
	node: string;
	status: string;
	timestamp: string;
};

type RecordState = 'idle' | 'recording' | 'stopped';

function classifyStatus(status: string): BlockedBy {
	if (status === 'BLACKLIST' || status === 'BLACKLIST_CNAME') return 'rule';
	return 'gravity';
}

export function Diagnose() {
	const [recordState, setRecordState] = useState<RecordState>('idle');
	const [clientIP, setClientIP] = useState('');
	const [entries, setEntries] = useState<BlockedEntry[]>([]);
	const [secondsLeft, setSecondsLeft] = useState(DEFAULT_DURATION_S);

	const esRef = useRef<EventSource | null>(null);
	const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);
	const feedbackTimers = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

	const stopRecording = useCallback(() => {
		esRef.current?.close();
		esRef.current = null;
		if (countdownRef.current) {
			clearInterval(countdownRef.current);
			countdownRef.current = null;
		}
		setRecordState('stopped');
	}, []);

	const startRecording = useCallback(() => {
		setEntries([]);
		setSecondsLeft(DEFAULT_DURATION_S);
		setRecordState('recording');

		const qs = clientIP.trim() ? `?client_ip=${encodeURIComponent(clientIP.trim())}` : '';
		const es = new EventSource(`/api/v1/diagnose/stream${qs}`, { withCredentials: true });

		es.addEventListener('blocked', (ev: MessageEvent) => {
			try {
				const raw: RawEvent = JSON.parse(ev.data);
				setEntries((prev) => {
					if (prev.some((e) => e.domain === raw.domain)) return prev;
					return [
						{
							domain: raw.domain,
							client: raw.client,
							node: raw.node,
							status: raw.status,
							timestamp: raw.timestamp,
							blockedBy: classifyStatus(raw.status),
							actionState: 'idle',
						},
						...prev,
					];
				});
			} catch {
				// malformed event — ignore
			}
		});

		es.onerror = () => {
			// SSE errors are non-fatal; the browser will retry automatically.
			// Don't stop the session on transient errors.
		};

		esRef.current = es;

		const interval = setInterval(() => {
			setSecondsLeft((s) => {
				if (s <= 1) {
					stopRecording();
					return 0;
				}
				return s - 1;
			});
		}, 1000);
		countdownRef.current = interval;
	}, [clientIP, stopRecording]);

	const handleExtend = () => {
		setSecondsLeft((s) => s + EXTEND_S);
	};

	const handleReset = () => {
		setEntries([]);
		setRecordState('idle');
		setSecondsLeft(DEFAULT_DURATION_S);
	};

	useEffect(() => {
		return () => {
			esRef.current?.close();
			if (countdownRef.current) clearInterval(countdownRef.current);
			for (const t of feedbackTimers.current.values()) clearTimeout(t);
		};
	}, []);

	function scheduleActionReset(domain: string) {
		const existing = feedbackTimers.current.get(domain);
		if (existing) clearTimeout(existing);
		const t = setTimeout(() => {
			setEntries((prev) =>
				prev.map((e) => (e.domain === domain ? { ...e, actionState: 'idle' } : e)),
			);
			feedbackTimers.current.delete(domain);
		}, 3000);
		feedbackTimers.current.set(domain, t);
	}

	async function handleAction(entry: BlockedEntry) {
		setEntries((prev) =>
			prev.map((e) => (e.domain === entry.domain ? { ...e, actionState: 'loading' } : e)),
		);
		try {
			if (entry.blockedBy === 'rule') {
				await removeDomainRule('deny', 'exact', entry.domain);
			} else {
				await addDomainRule('allow', 'exact', entry.domain);
			}
			setEntries((prev) =>
				prev.map((e) => (e.domain === entry.domain ? { ...e, actionState: 'ok' } : e)),
			);
		} catch {
			setEntries((prev) =>
				prev.map((e) => (e.domain === entry.domain ? { ...e, actionState: 'error' } : e)),
			);
		}
		scheduleActionReset(entry.domain);
	}

	const isRecording = recordState === 'recording';
	const isStopped = recordState === 'stopped';
	const isIdle = recordState === 'idle';
	const isLow = secondsLeft <= 15;

	return (
		<div className={styles.page}>
			<div className={styles.pageHeader}>
				<h1 className={styles.pageTitle}>Site Diagnoser</h1>
				<p className={styles.pageSubtitle}>
					Start recording, open the broken site in another tab, then stop to review blocked
					domains.
				</p>
			</div>

			<div className={styles.controls}>
				<div className={styles.ipField}>
					<label htmlFor='diag-ip' className={styles.ipLabel}>
						Client IP (optional — leave blank to watch all devices)
					</label>
					<input
						id='diag-ip'
						type='text'
						className={styles.ipInput}
						placeholder='e.g. 192.168.1.42'
						value={clientIP}
						onChange={(e) => setClientIP(e.target.value)}
						disabled={isRecording}
					/>
				</div>

				<div className={styles.controlActions}>
					{(isIdle || isStopped) && (
						<button type='button' className={styles.startBtn} onClick={startRecording}>
							<Circle size={14} />
							{isStopped ? 'Record again' : 'Start recording'}
						</button>
					)}
					{isRecording && (
						<>
							<button type='button' className={styles.extendBtn} onClick={handleExtend}>
								<Timer size={14} />
								+{EXTEND_S}s
							</button>
							<button type='button' className={styles.stopBtn} onClick={stopRecording}>
								<Square size={14} />
								Stop
							</button>
						</>
					)}
					{isStopped && entries.length > 0 && (
						<button type='button' className={styles.resetBtn} onClick={handleReset}>
							Clear
						</button>
					)}
				</div>
			</div>

			{isRecording && (
				<div className={styles.timerPanel}>
					<span className={styles.timerDot} />
					<span className={styles.timerLabel}>
						Recording — open the broken site in another tab
					</span>
					<span
						className={classNames(styles.timerCountdown, { [styles.timerCountdownLow]: isLow })}
					>
						{secondsLeft}s
					</span>
				</div>
			)}

			{entries.length > 0 ? (
				<>
					<div className={styles.resultsHeader}>
						<span className={styles.resultsCount}>
							{entries.length} blocked {entries.length === 1 ? 'domain' : 'domains'} captured
						</span>
						{isRecording && (
							<span className={styles.liveIndicator}>
								<span className={styles.liveDot} />
								Live
							</span>
						)}
					</div>
					<div className={styles.tableWrap}>
						<table className={styles.table}>
							<thead>
								<tr>
									<th>Domain</th>
									<th>Client</th>
									<th>Node</th>
									<th>Blocked by</th>
									<th aria-label='Action' />
								</tr>
							</thead>
							<tbody>
								{entries.map((entry) => (
									<tr key={entry.domain}>
										<td className={styles.domainCell}>{entry.domain}</td>
										<td className={styles.clientCell}>{entry.client || '—'}</td>
										<td>
											<span className={styles.nodeTag}>{entry.node}</span>
										</td>
										<td>
											<span
												className={classNames(styles.statusBadge, {
													[styles.statusRule]: entry.blockedBy === 'rule',
													[styles.statusGravity]: entry.blockedBy === 'gravity',
												})}
											>
												{entry.blockedBy === 'rule' ? 'Block rule' : 'Gravity list'}
											</span>
										</td>
										<td className={styles.actionCell}>
											{entry.actionState === 'ok' && (
												<span
													className={classNames(
														styles.actionFeedback,
														styles.feedbackOk,
													)}
												>
													✓{' '}
													{entry.blockedBy === 'rule'
														? 'Rule removed'
														: 'Allow rule added'}
												</span>
											)}
											{entry.actionState === 'error' && (
												<span
													className={classNames(
														styles.actionFeedback,
														styles.feedbackErr,
													)}
												>
													✗ Failed
												</span>
											)}
											{entry.actionState === 'idle' &&
												entry.blockedBy === 'gravity' && (
													<button
														type='button'
														className={styles.allowBtn}
														onClick={() => handleAction(entry)}
														title={`Add "${entry.domain}" to allowlist`}
													>
														<Shield size={12} />
														Allow
													</button>
												)}
											{entry.actionState === 'idle' &&
												entry.blockedBy === 'rule' && (
													<button
														type='button'
														className={styles.removeBtn}
														onClick={() => handleAction(entry)}
														title={`Remove block rule for "${entry.domain}"`}
													>
														Remove rule
													</button>
												)}
											{entry.actionState === 'loading' && (
												<RefreshCw size={14} className={styles.spin} />
											)}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				</>
			) : (
				<div className={styles.emptyState}>
					{isIdle && (
						<>
							<Circle
								size={36}
								className={styles.idleIcon}
								aria-hidden='true'
							/>
							Enter an optional client IP, then click Start recording. Open the broken
							site in another tab — blocked domains will appear here in real time.
						</>
					)}
					{isRecording && 'Waiting for blocked DNS queries…'}
					{isStopped && 'No blocked domains were captured.'}
				</div>
			)}
		</div>
	);
}
