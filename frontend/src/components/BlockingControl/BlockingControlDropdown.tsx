import { useState, useEffect, useRef } from 'react';
import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import {
	ChevronDown,
	Shield,
	ShieldOff,
	Infinity,
	Clock,
	Timer,
	Loader2,
} from 'lucide-react';
import { setClusterBlocking } from '@/lib/api/blocking';
import { CustomDisableModal } from './CustomDisableModal';
import styles from './BlockingControlDropdown.module.scss';

function formatCountdown(seconds: number): string {
	const h = Math.floor(seconds / 3600);
	const m = Math.floor((seconds % 3600) / 60);
	const s = Math.floor(seconds % 60);
	if (h > 0) {
		return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
	}
	return `${m}:${s.toString().padStart(2, '0')}`;
}

function useBlockingCountdown(
	minSeconds: number | undefined,
	receivedAt: Date | undefined,
	active: boolean,
): number | null {
	const [now, setNow] = useState(() => Date.now());

	useEffect(() => {
		if (!active || minSeconds == null || minSeconds <= 0 || !receivedAt) return;
		const id = setInterval(() => setNow(Date.now()), 1000);
		return () => clearInterval(id);
	}, [active, minSeconds, receivedAt]);

	if (!active || minSeconds == null || minSeconds <= 0 || !receivedAt) return null;
	const elapsed = Math.floor((now - receivedAt.getTime()) / 1000);
	return Math.max(0, minSeconds - elapsed);
}

import type { useClusterOverview } from '@/hooks/useClusterOverview';

type ClusterOverview = ReturnType<typeof useClusterOverview>;

type BlockingControlDropdownProps = {
	sidebarOpen: boolean;
	clusterOverview: ClusterOverview;
	onMobileClose?: () => void;
};

const PRESETS = [
	{ label: 'Indefinitely', icon: Infinity, timer: undefined },
	{ label: 'For 10 seconds', icon: Clock, timer: 10 },
	{ label: 'For 30 seconds', icon: Clock, timer: 30 },
	{ label: 'For 5 minutes', icon: Clock, timer: 300 },
];

const LONG_TIMER_SECONDS = 300; // 5 min: sync from server every 60s to correct drift
const LONG_TIMER_SYNC_INTERVAL_MS = 60_000;

export function BlockingControlDropdown({
	sidebarOpen,
	clusterOverview,
	onMobileClose,
}: BlockingControlDropdownProps) {
	const {
		blocking,
		blockingUpdatedAt,
		refetchBlocking,
		applyBlockingState,
	} = clusterOverview;
	const [dropdownOpen, setDropdownOpen] = useState(false);
	const [customModalOpen, setCustomModalOpen] = useState(false);
	const [submitting, setSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	// Local timer override: when user initiates disable-with-timer, we use the requested
	// value for display (avoids server/Pi-hole drift and SSE overwrites). Page reload
	// uses server state since local state doesn't survive.
	const [localTimer, setLocalTimer] = useState<{ minSeconds: number; startedAt: number } | null>(
		null,
	);
	// Ref is set synchronously on click so we read it immediately even if an SSE-triggered
	// re-render happens before React commits our setLocalTimer (prevents showing stale 0:13)
	const localTimerRef = useRef<{ minSeconds: number; startedAt: number } | null>(null);

	const mode = blocking?.summary?.mode;
	const isEnabled = mode === 'enabled';
	const isLoading = blocking === undefined;
	const disabled = isLoading || submitting;

	const serverTimers = blocking?.summary?.timers;
	const serverMinSeconds = serverTimers?.present ? serverTimers.minSeconds ?? 0 : 0;

	// Effective values: ref (sync) beats state (async) beats server; ref prevents SSE race
	const effectiveTimer = localTimerRef.current ?? localTimer;
	const effectiveMinSeconds = effectiveTimer ? effectiveTimer.minSeconds : serverMinSeconds;
	const effectiveReceivedAt = effectiveTimer
		? new Date(effectiveTimer.startedAt)
		: blockingUpdatedAt;
	const showCountdown = !isEnabled && effectiveMinSeconds > 0;
	const remainingSeconds = useBlockingCountdown(
		effectiveMinSeconds,
		effectiveReceivedAt,
		showCountdown,
	);

	// When countdown hits 0: clear local timer and ref, refetch to update mode/shield,
	// and follow-up refetches to catch Pi-hole re-enabling (clock skew)
	const hasRefetchedAtZero = useRef(false);
	useEffect(() => {
		if (remainingSeconds === 0 && showCountdown) {
			if (!hasRefetchedAtZero.current) {
				hasRefetchedAtZero.current = true;
				localTimerRef.current = null;
				setLocalTimer(null);
				refetchBlocking();
				const t1 = window.setTimeout(() => refetchBlocking(), 1500);
				const t2 = window.setTimeout(() => refetchBlocking(), 3500);
				return () => {
					clearTimeout(t1);
					clearTimeout(t2);
				};
			}
		} else if (remainingSeconds != null && remainingSeconds > 0) {
			hasRefetchedAtZero.current = false;
		}
	}, [remainingSeconds, showCountdown, refetchBlocking]);

	// Long-timer drift sync: for timers >= 5 min, refetch every 60s and re-anchor
	// to server minSeconds to correct clock drift across nodes
	useEffect(() => {
		if (!localTimer || localTimer.minSeconds < LONG_TIMER_SECONDS) return;
		const id = window.setInterval(() => {
			refetchBlocking().then((state) => {
				const serverMin = state?.summary?.timers?.minSeconds;
				if (serverMin != null && serverMin > 0) {
					const t = { minSeconds: serverMin, startedAt: Date.now() };
					localTimerRef.current = t;
					setLocalTimer(t);
				}
			});
		}, LONG_TIMER_SYNC_INTERVAL_MS);
		return () => clearInterval(id);
	}, [localTimer?.minSeconds, refetchBlocking]);

	function handleApplyState(
		nextState: Parameters<typeof applyBlockingState>[0],
		opts?: { requestedTimerSeconds: number },
	) {
		applyBlockingState(nextState);
		if (opts?.requestedTimerSeconds != null && opts.requestedTimerSeconds > 0) {
			const t = {
				minSeconds: opts.requestedTimerSeconds,
				startedAt: Date.now(),
			};
			localTimerRef.current = t;
			setLocalTimer(t);
		}
	}

	async function handleDisable(timer?: number) {
		setError(null);
		setSubmitting(true);
		// Ref is set synchronously so we read it even if SSE triggers a re-render before
		// our setState is committed; state is for persistence across renders
		if (timer != null && timer > 0) {
			const t = { minSeconds: timer, startedAt: Date.now() };
			localTimerRef.current = t;
			setLocalTimer(t);
		}
		try {
			const nextState = await setClusterBlocking({ blocking: false, timer });
			applyBlockingState(nextState);
			setDropdownOpen(false);
			onMobileClose?.();
		} catch (err) {
			const msg = err instanceof Error ? err.message : 'Failed to disable blocking';
			setError(msg);
			if (timer != null && timer > 0) {
				localTimerRef.current = null;
				setLocalTimer(null);
			}
		} finally {
			setSubmitting(false);
		}
	}

	async function handleEnable() {
		setError(null);
		setSubmitting(true);
		try {
			const nextState = await setClusterBlocking({ blocking: true });
			applyBlockingState(nextState);
			onMobileClose?.();
		} catch (err) {
			const msg = err instanceof Error ? err.message : 'Failed to enable blocking';
			setError(msg);
		} finally {
			setSubmitting(false);
		}
	}

	function openCustomModal() {
		setDropdownOpen(false);
		setCustomModalOpen(true);
		onMobileClose?.();
	}

	function handleCustomSuccess() {
		setCustomModalOpen(false);
		setError(null);
	}

	function handleCustomError() {
		// Modal handles its own error display
	}

	if (isEnabled) {
		return (
			<>
				<DropdownMenu.Root open={dropdownOpen} onOpenChange={setDropdownOpen}>
					<DropdownMenu.Trigger asChild>
						<button
							type="button"
							className={styles.trigger}
							disabled={disabled}
							title="Disable blocking"
							aria-label="Disable blocking"
							aria-busy={submitting}
							aria-haspopup="menu"
						>
							{submitting ? (
								<Loader2 size={18} className={styles.icon} />
							) : (
								<Shield size={18} className={styles.icon} />
							)}
							{sidebarOpen && (
								<>
									<span className={styles.label}>Disable blocking</span>
									<ChevronDown size={16} className={styles.icon} />
								</>
							)}
						</button>
					</DropdownMenu.Trigger>
					<DropdownMenu.Portal>
						<DropdownMenu.Content
							className={styles.dropdownContent}
							align="start"
							sideOffset={4}
						>
							{PRESETS.map(({ label, icon: Icon, timer }) => (
								<DropdownMenu.Item
									key={label}
									className={styles.dropdownItem}
									onSelect={() => handleDisable(timer)}
									disabled={submitting}
								>
									<Icon size={16} className={styles.icon} />
									{label}
								</DropdownMenu.Item>
							))}
							<DropdownMenu.Item
								className={styles.dropdownItem}
								onSelect={openCustomModal}
								disabled={submitting}
							>
								<Timer size={16} className={styles.icon} />
								Custom time…
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Portal>
				</DropdownMenu.Root>
				{error && <p className={styles.error}>{error}</p>}
				<CustomDisableModal
					open={customModalOpen}
					onOpenChange={setCustomModalOpen}
					onApplyState={handleApplyState}
					onSuccess={handleCustomSuccess}
					onError={handleCustomError}
				/>
			</>
		);
	}

	return (
		<>
			<button
				type="button"
				className={styles.trigger}
				disabled={disabled}
				onClick={handleEnable}
				title={
					remainingSeconds != null && remainingSeconds > 0
						? `Enable blocking (${formatCountdown(remainingSeconds)} remaining)`
						: 'Enable blocking'
				}
				aria-label={
					remainingSeconds != null && remainingSeconds > 0
						? `Enable blocking in ${formatCountdown(remainingSeconds)}`
						: 'Enable blocking'
				}
				aria-busy={submitting}
			>
				{submitting ? (
					<Loader2 size={18} className={styles.icon} />
				) : (
					<ShieldOff size={18} className={styles.icon} />
				)}
				{sidebarOpen && (
					<span className={styles.label}>
						Enable blocking
						{remainingSeconds != null &&
							remainingSeconds > 0 && (
								<> ({formatCountdown(remainingSeconds)})</>
							)}
					</span>
				)}
			</button>
			{error && <p className={styles.error}>{error}</p>}
		</>
	);
}
