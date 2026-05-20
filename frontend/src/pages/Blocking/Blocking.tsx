import { useState, useEffect, useRef } from 'react';
import { Shield, Loader2, ChevronDown, RefreshCw, RotateCcw } from 'lucide-react';
import { useClusterOverview } from '@/hooks/useClusterOverview';
import { setClusterBlocking, setNodeBlocking, flushCache, restartDNS } from '@/lib/api/blocking';
import type { FlushCacheResult, RestartDNSResult } from '@/lib/api/blocking';
import type { ClusterBlockingNode } from '@/types/blocking';
import { getBlockingDisplayState } from '@/utils/blockingStatus';
import styles from './Blocking.module.scss';

const CLUSTER_PRESETS = [
	{ label: '10 sec', seconds: 10 },
	{ label: '30 sec', seconds: 30 },
	{ label: '5 min', seconds: 300 },
	{ label: '1 hour', seconds: 3600 },
];

type CustomUnit = 'secs' | 'mins' | 'hours';
const CUSTOM_UNITS: { value: CustomUnit; label: string; max: number }[] = [
	{ value: 'secs', label: 'Seconds', max: 86400 },
	{ value: 'mins', label: 'Minutes', max: 1440 },
	{ value: 'hours', label: 'Hours', max: 24 },
];
function customToSeconds(value: number, unit: CustomUnit): number {
	if (unit === 'secs') return value;
	if (unit === 'mins') return value * 60;
	return value * 3600;
}
function formatCustomLabel(value: number, unit: CustomUnit): string {
	if (unit === 'secs') return `${value}s`;
	if (unit === 'mins') return value === 1 ? '1 min' : `${value} min`;
	return value === 1 ? '1 hour' : `${value} hr`;
}

function formatCountdown(seconds: number): string {
	const h = Math.floor(seconds / 3600);
	const m = Math.floor((seconds % 3600) / 60);
	const s = Math.floor(seconds % 60);
	if (h > 0) {
		return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
	}
	return `${m}:${s.toString().padStart(2, '0')}`;
}

export function Blocking() {
	const clusterOverview = useClusterOverview();
	const {
		blocking,
		blockingUpdatedAt,
		refetchBlocking,
		applyBlockingState,
		setOptimisticEnabled,
		setOptimisticNodeEnabled,
		applyOptimisticNodeDisable,
		applyOptimisticClusterDisable,
		requestedNodeDisplay,
		clearRequestedNodeDisplay,
	} = clusterOverview;

	const [clusterSubmitting, setClusterSubmitting] = useState(false);
	const [clusterError, setClusterError] = useState<string | null>(null);
	const [flushSubmitting, setFlushSubmitting] = useState(false);
	const [flushResult, setFlushResult] = useState<FlushCacheResult | null>(null);
	const [flushError, setFlushError] = useState<string | null>(null);
	const flushResultTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
	const [restartSubmitting, setRestartSubmitting] = useState(false);
	const [restartResult, setRestartResult] = useState<RestartDNSResult | null>(null);
	const [restartError, setRestartError] = useState<string | null>(null);
	const restartResultTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
	const [customSeconds, setCustomSeconds] = useState(60);
	const [customUnit, setCustomUnit] = useState<CustomUnit>('mins');
	const [customSecondsByNode, setCustomSecondsByNode] = useState<Record<number, number>>({});
	const [customUnitByNode, setCustomUnitByNode] = useState<Record<number, CustomUnit>>({});
	const [submittingNodeId, setSubmittingNodeId] = useState<number | null>(null);
	const [now, setNow] = useState(() => Date.now());
	/** Local state so setting it forces a re-render and we always paint with the requested timer first. */
	const [pendingNodeDisplay, setPendingNodeDisplay] = useState<{
		nodeId: number;
		seconds: number;
		requestedAt: number;
	} | null>(null);

	const mode = blocking?.summary?.mode ?? 'degraded';
	const isClusterEnabled = mode === 'enabled';
	const receivedAt = blockingUpdatedAt;

	const nodes = blocking?.nodes
		? (Object.entries(blocking.nodes) as [string, ClusterBlockingNode][])
				.map(([id, n]) => ({ id: Number(id), ...n }))
				.sort((a, b) => a.node.name.localeCompare(b.node.name))
		: [];

	// Single tick for all countdowns; run when any node has a timer or we're showing an override
	const anyNodeHasTimer =
		nodes.some((n) => n.timer != null && n.timer > 0) ||
		requestedNodeDisplay != null ||
		pendingNodeDisplay != null;
	useEffect(() => {
		if (!receivedAt || !anyNodeHasTimer) return;
		const id = setInterval(() => setNow(Date.now()), 1000);
		return () => clearInterval(id);
	}, [receivedAt, anyNodeHasTimer]);

	// Derive per-node and cluster remaining; use requestedNodeDisplay when present so the UI shows the requested timer from click
	const receivedAtMs = receivedAt?.getTime();
	function getNodeRemaining(
		nodeTimer: number | undefined | null,
		atMs: number | undefined,
	): number | null {
		const at = atMs ?? receivedAtMs;
		if (at == null || nodeTimer == null || nodeTimer <= 0) return null;
		// Clamp elapsed to >= 0 so when `now` is stale (e.g. from previous tick), we never show remaining > timer
		const elapsed = Math.max(0, Math.floor((now - at) / 1000));
		return Math.max(0, nodeTimer - elapsed);
	}
	const nodeRemainings = new Map<number, number | null>(
		nodes.map((n) => {
			const pending =
				pendingNodeDisplay != null && pendingNodeDisplay.nodeId == n.id
					? pendingNodeDisplay
					: null;
			const useOverride =
				pending ?? (requestedNodeDisplay?.nodeId == n.id ? requestedNodeDisplay : null);
			const timer = useOverride ? useOverride.seconds : n.timer;
			const at = useOverride ? useOverride.requestedAt : receivedAtMs;
			const remaining = getNodeRemaining(timer, at);
			return [n.id, remaining] as const;
		}),
	);
	const clusterRemaining =
		nodes.length === 0
			? 0
			: Math.max(
					0,
					...nodes.map((n) => {
						const pending =
							pendingNodeDisplay != null && pendingNodeDisplay.nodeId == n.id
								? pendingNodeDisplay
								: null;
						const useOverride =
							pending ??
							(requestedNodeDisplay?.nodeId == n.id ? requestedNodeDisplay : null);
						const timer = useOverride ? useOverride.seconds : n.timer;
						const at = useOverride ? useOverride.requestedAt : receivedAtMs;
						return getNodeRemaining(timer, at) ?? 0;
					}),
				);
	const showClusterCountdown = !isClusterEnabled && clusterRemaining > 0;
	// Hide "Enable blocking" when countdown has hit 0 (re-enabled); avoid stale button until refetch completes
	const showClusterEnableButton = !isClusterEnabled && clusterRemaining > 0;

	const prevClusterRemaining = useRef<number | null>(null);
	const prevNodeRemainings = useRef<Map<number, number | null>>(new Map());
	// When countdown reaches 0: update UI to enabled immediately, then refetch after a delay
	// so upstream nodes have time to settle (avoids fetching before Pi-holes have re-enabled).
	useEffect(() => {
		const prev = prevClusterRemaining.current;
		prevClusterRemaining.current = clusterRemaining;
		const justHitZero = prev != null && prev > 0 && clusterRemaining === 0;
		if (justHitZero) {
			setOptimisticEnabled();
			const refetchLater = window.setTimeout(() => {
				refetchBlocking();
			}, 4000);
			return () => window.clearTimeout(refetchLater);
		}
	}, [clusterRemaining, setOptimisticEnabled, refetchBlocking]);

	// When a single node's countdown hits 0 (cluster still has time left), optimistically set that node to enabled and summary to mixed.
	useEffect(() => {
		if (clusterRemaining === 0 || !anyNodeHasTimer) return;
		const prevMap = prevNodeRemainings.current;
		for (const [nid, remaining] of nodeRemainings) {
			const prevRem = prevMap.get(nid);
			prevMap.set(nid, remaining);
			const justHitZero = prevRem != null && prevRem > 0 && remaining === 0;
			if (justHitZero) {
				setOptimisticNodeEnabled(nid);
			}
		}
	}, [clusterRemaining, anyNodeHasTimer, nodeRemainings, setOptimisticNodeEnabled]);

	async function handleClusterEnable() {
		setClusterError(null);
		setClusterSubmitting(true);
		try {
			const next = await setClusterBlocking({ blocking: true });
			applyBlockingState(next);
		} catch (err) {
			setClusterError(err instanceof Error ? err.message : 'Failed to enable blocking');
		} finally {
			setClusterSubmitting(false);
		}
	}

	async function handleClusterDisable(timerSeconds: number) {
		setClusterError(null);
		setClusterSubmitting(true);
		applyOptimisticClusterDisable(timerSeconds);
		try {
			const next = await setClusterBlocking({ blocking: false, timer: timerSeconds });
			applyBlockingState(next, {
				requestedTimerSeconds: timerSeconds,
				overrideAllNodesTimer: timerSeconds,
			});
		} catch (err) {
			setClusterError(err instanceof Error ? err.message : 'Failed to disable blocking');
			refetchBlocking();
		} finally {
			setClusterSubmitting(false);
		}
	}

	async function handleNodeEnable(nodeId: number) {
		setSubmittingNodeId(nodeId);
		try {
			const next = await setNodeBlocking(nodeId, { blocking: true });
			applyBlockingState(next);
		} finally {
			setSubmittingNodeId(null);
		}
	}

	async function handleNodeDisable(nodeId: number, timerSeconds: number) {
		setSubmittingNodeId(nodeId);
		// Set state immediately so the next paint shows requested timer (state update triggers re-render).
		const requestedAt = Date.now();
		setPendingNodeDisplay({ nodeId, seconds: timerSeconds, requestedAt });
		applyOptimisticNodeDisable(nodeId, timerSeconds);
		try {
			const next = await setNodeBlocking(nodeId, { blocking: false, timer: timerSeconds });
			setPendingNodeDisplay((prev) =>
				prev?.nodeId === nodeId
					? { nodeId, seconds: timerSeconds, requestedAt: Date.now() }
					: prev,
			);
			applyBlockingState(next, {
				overrideNodeTimer: { nodeId, seconds: timerSeconds },
			});
			// Clear after paint so we never show server timer first.
			requestAnimationFrame(() => {
				requestAnimationFrame(() => {
					clearRequestedNodeDisplay(nodeId);
					setPendingNodeDisplay((prev) => (prev?.nodeId === nodeId ? null : prev));
				});
			});
		} catch {
			refetchBlocking();
			clearRequestedNodeDisplay(nodeId);
			setPendingNodeDisplay((prev) => (prev?.nodeId === nodeId ? null : prev));
		} finally {
			setSubmittingNodeId(null);
		}
	}

	async function handleRestartDNS() {
		setRestartSubmitting(true);
		setRestartResult(null);
		setRestartError(null);
		if (restartResultTimer.current) clearTimeout(restartResultTimer.current);
		try {
			const result = await restartDNS();
			setRestartResult(result);
			restartResultTimer.current = setTimeout(() => setRestartResult(null), 5000);
		} catch (err) {
			const msg = err instanceof Error ? err.message : 'Failed to restart DNS';
			setRestartError(msg);
			restartResultTimer.current = setTimeout(() => setRestartError(null), 5000);
		} finally {
			setRestartSubmitting(false);
		}
	}

	async function handleFlushCache() {
		setFlushSubmitting(true);
		setFlushResult(null);
		setFlushError(null);
		if (flushResultTimer.current) clearTimeout(flushResultTimer.current);
		try {
			const result = await flushCache();
			setFlushResult(result);
			flushResultTimer.current = setTimeout(() => setFlushResult(null), 5000);
		} catch (err) {
			const msg = err instanceof Error ? err.message : 'Failed to flush cache';
			setFlushError(msg);
			flushResultTimer.current = setTimeout(() => setFlushError(null), 5000);
		} finally {
			setFlushSubmitting(false);
		}
	}

	useEffect(() => {
		return () => {
			if (flushResultTimer.current) clearTimeout(flushResultTimer.current);
			if (restartResultTimer.current) clearTimeout(restartResultTimer.current);
		};
	}, []);

	const blockingDisplay = getBlockingDisplayState(blocking?.summary);
	const { icon: StatusIcon, colorVar, label: statusLabel } = blockingDisplay;

	return (
		<div className={styles.page}>
			<section className={styles.clusterSection} aria-labelledby='cluster-status-heading'>
				<h2 id='cluster-status-heading' className={styles.sectionTitle}>
					Cluster
				</h2>
				<div className={styles.clusterCard}>
					<div className={styles.statusRow}>
						<StatusIcon
							size={24}
							className={styles.statusIcon}
							style={{ color: colorVar }}
							aria-hidden
						/>
						<div>
							<span className={styles.statusLabel}>{statusLabel}</span>
							{showClusterCountdown && (
								<span className={styles.countdown} aria-live='polite'>
									{formatCountdown(clusterRemaining)} left
								</span>
							)}
						</div>
					</div>
					{clusterError && <p className={styles.error}>{clusterError}</p>}
					<div className={styles.actions}>
						{showClusterEnableButton && (
							<button
								type='button'
								className={styles.primaryButton}
								onClick={handleClusterEnable}
								disabled={clusterSubmitting}
								aria-busy={clusterSubmitting}
							>
								{clusterSubmitting ? (
									<Loader2 size={18} className={styles.spin} />
								) : (
									<Shield size={18} />
								)}
								Enable blocking
							</button>
						)}
						<div
							className={styles.presets}
							role='group'
							aria-label='Disable blocking for a duration'
						>
							{CLUSTER_PRESETS.map(({ label, seconds }) => (
								<button
									key={label}
									type='button'
									className={styles.presetButton}
									onClick={() => handleClusterDisable(seconds)}
									disabled={clusterSubmitting}
									aria-busy={clusterSubmitting}
									aria-label={`Disable blocking for ${label}`}
								>
									Disable {label}
								</button>
							))}
						</div>
						<div className={styles.customRow}>
							<label htmlFor='blocking-custom-value' className={styles.customLabel}>
								Custom:
							</label>
							<input
								id='blocking-custom-value'
								type='number'
								min={1}
								max={CUSTOM_UNITS.find((u) => u.value === customUnit)?.max ?? 86400}
								value={customSeconds}
								onChange={(e) => {
									const v = e.target.valueAsNumber;
									setCustomSeconds(Number.isFinite(v) && v >= 0 ? v : 0);
								}}
								className={styles.customInput}
								disabled={clusterSubmitting}
								aria-label='Custom duration'
							/>
							<span className={styles.unitSelectWrap}>
								<select
									className={styles.unitSelect}
									value={customUnit}
									onChange={(e) => setCustomUnit(e.target.value as CustomUnit)}
									disabled={clusterSubmitting}
									aria-label='Time unit'
								>
									{CUSTOM_UNITS.map(({ value, label }) => (
										<option key={value} value={value}>
											{label}
										</option>
									))}
								</select>
								<ChevronDown
									size={16}
									className={styles.unitSelectIcon}
									aria-hidden
								/>
							</span>
							<button
								type='button'
								className={styles.presetButton}
								onClick={() =>
									handleClusterDisable(customToSeconds(customSeconds, customUnit))
								}
								disabled={
									clusterSubmitting ||
									customSeconds <= 0 ||
									customToSeconds(customSeconds, customUnit) > 86400
								}
								aria-label={`Disable blocking for ${formatCustomLabel(customSeconds, customUnit)}`}
							>
								Disable for {formatCustomLabel(customSeconds, customUnit)}
							</button>
						</div>
					</div>
				</div>
			</section>

			<section className={styles.nodesSection} aria-labelledby='nodes-heading'>
				<h2 id='nodes-heading' className={styles.sectionTitle}>
					Nodes
				</h2>
				<div className={styles.nodeGrid}>
					{nodes.map(
						({
							id,
							node,
							blocking: nodeBlocking,
							timer: _nodeTimer,
							error: nodeError,
						}) => {
							const nodeDisabled =
								nodeBlocking === 'disabled' ||
								requestedNodeDisplay?.nodeId == id ||
								(pendingNodeDisplay != null && pendingNodeDisplay.nodeId == id);
							const remaining = nodeRemainings.get(id);
							const hasTimeLeft = remaining != null && remaining > 0;
							const isSubmitting = submittingNodeId === id;
							const nodeCustomValue = customSecondsByNode[id] ?? 60;
							const nodeUnit = customUnitByNode[id] ?? 'mins';
							const nodeCustomSeconds = customToSeconds(nodeCustomValue, nodeUnit);
							return (
								<div
									key={id}
									className={styles.nodeCard}
									data-status={nodeBlocking}
								>
									<div className={styles.nodeTop}>
										<div className={styles.nodeHeader}>
											<span className={styles.nodeName}>{node.name}</span>
											<span className={styles.nodeStatus}>
												{nodeBlocking}
											</span>
											{hasTimeLeft && (
												<span className={styles.nodeTimer}>
													{formatCountdown(remaining!)} left
												</span>
											)}
										</div>
										{nodeDisabled && hasTimeLeft && (
											<button
												type='button'
												className={styles.nodeReenable}
												onClick={() => handleNodeEnable(id)}
												disabled={isSubmitting}
												aria-busy={isSubmitting}
												aria-label={`Re-enable blocking on ${node.name}`}
											>
												{isSubmitting ? (
													<Loader2 size={14} className={styles.spin} />
												) : (
													<Shield size={14} />
												)}
												Re-enable
											</button>
										)}
									</div>
									{nodeError && <p className={styles.nodeError}>{nodeError}</p>}
									<div
										className={styles.nodeDisableRow}
										role='group'
										aria-label={`Disable blocking on ${node.name}`}
									>
										{CLUSTER_PRESETS.map(({ label, seconds }) => (
											<button
												key={label}
												type='button'
												className={styles.nodeChip}
												onClick={() => handleNodeDisable(id, seconds)}
												disabled={isSubmitting}
												aria-label={`Disable ${label} on ${node.name}`}
											>
												{label}
											</button>
										))}
										<span className={styles.nodeCustomWrap}>
											<input
												id={`blocking-node-${id}-custom`}
												type='number'
												min={1}
												max={
													CUSTOM_UNITS.find((u) => u.value === nodeUnit)
														?.max ?? 86400
												}
												value={nodeCustomValue}
												onChange={(e) => {
													const v = e.target.valueAsNumber;
													setCustomSecondsByNode((prev) => ({
														...prev,
														[id]: Number.isFinite(v) && v >= 0 ? v : 0,
													}));
												}}
												className={styles.nodeCustomInput}
												disabled={isSubmitting}
												aria-label={`Custom duration for ${node.name}`}
											/>
											<span className={styles.nodeUnitSelectWrap}>
												<select
													className={styles.nodeUnitSelect}
													value={nodeUnit}
													onChange={(e) =>
														setCustomUnitByNode((prev) => ({
															...prev,
															[id]: e.target.value as CustomUnit,
														}))
													}
													disabled={isSubmitting}
													aria-label={`Unit for ${node.name}`}
												>
													{CUSTOM_UNITS.map(({ value, label }) => (
														<option key={value} value={value}>
															{label}
														</option>
													))}
												</select>
												<ChevronDown
													size={14}
													className={styles.unitSelectIcon}
													aria-hidden
												/>
											</span>
											<button
												type='button'
												className={styles.nodeChip}
												onClick={() =>
													handleNodeDisable(id, nodeCustomSeconds)
												}
												disabled={
													isSubmitting ||
													nodeCustomValue <= 0 ||
													nodeCustomSeconds > 86400
												}
												aria-label={`Disable ${formatCustomLabel(nodeCustomValue, nodeUnit)} on ${node.name}`}
											>
												{formatCustomLabel(nodeCustomValue, nodeUnit)}
											</button>
										</span>
									</div>
								</div>
							);
						},
					)}
				</div>
			</section>

			<section className={styles.quickActionsSection} aria-labelledby='quick-actions-heading'>
				<h2 id='quick-actions-heading' className={styles.sectionTitle}>
					Quick Actions
				</h2>
				<div className={styles.clusterCard}>
					<div className={styles.quickAction}>
						<div className={styles.quickActionInfo}>
							<span className={styles.quickActionLabel}>Flush DNS Cache</span>
							<span className={styles.quickActionDesc}>
								Clear cached DNS responses on all nodes. Use after adding an allow
								rule to avoid waiting for TTL expiry.
							</span>
						</div>
						<button
							type='button'
							className={styles.quickActionBtn}
							onClick={handleFlushCache}
							disabled={flushSubmitting}
							aria-busy={flushSubmitting}
						>
							{flushSubmitting ? (
								<Loader2 size={16} className={styles.spin} />
							) : (
								<RefreshCw size={16} />
							)}
							{flushSubmitting ? 'Flushing…' : 'Flush Cache'}
						</button>
					</div>
					{flushError && <p className={styles.error}>{flushError}</p>}
					{flushResult && (
						<div className={styles.actionResults} role='status' aria-live='polite'>
							{Object.values(flushResult.nodes)
								.sort((a, b) => a.node.name.localeCompare(b.node.name))
								.map((nr) => (
									<span
										key={nr.node.id}
										className={
											nr.success ? styles.actionNodeOk : styles.actionNodeErr
										}
										title={nr.error || undefined}
									>
										{nr.success ? '✓' : '✗'} {nr.node.name}
										{!nr.success && nr.error && (
											<span className={styles.actionNodeErrMsg}>
												{' '}
												— {nr.error}
											</span>
										)}
									</span>
								))}
						</div>
					)}
					<div className={styles.quickAction}>
						<div className={styles.quickActionInfo}>
							<span className={styles.quickActionLabel}>Restart DNS</span>
							<span className={styles.quickActionDesc}>
								Restart the FTL DNS resolver on all nodes. Use after
								configuration changes that require a resolver reload.
							</span>
						</div>
						<button
							type='button'
							className={styles.quickActionBtn}
							onClick={handleRestartDNS}
							disabled={restartSubmitting}
							aria-busy={restartSubmitting}
						>
							{restartSubmitting ? (
								<Loader2 size={16} className={styles.spin} />
							) : (
								<RotateCcw size={16} />
							)}
							{restartSubmitting ? 'Restarting…' : 'Restart DNS'}
						</button>
					</div>
					{restartError && <p className={styles.error}>{restartError}</p>}
					{restartResult && (
						<div className={styles.actionResults} role='status' aria-live='polite'>
							{Object.values(restartResult.nodes)
								.sort((a, b) => a.node.name.localeCompare(b.node.name))
								.map((nr) => (
									<span
										key={nr.node.id}
										className={
											nr.success ? styles.actionNodeOk : styles.actionNodeErr
										}
										title={nr.error || undefined}
									>
										{nr.success ? '✓' : '✗'} {nr.node.name}
										{!nr.success && nr.error && (
											<span className={styles.actionNodeErrMsg}>
												{' '}
												— {nr.error}
											</span>
										)}
									</span>
								))}
						</div>
					)}
				</div>
			</section>
		</div>
	);
}
