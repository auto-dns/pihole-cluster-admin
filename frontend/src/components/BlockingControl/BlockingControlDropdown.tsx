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
import { useClusterOverview } from '@/hooks/useClusterOverview';
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

type BlockingControlDropdownProps = {
	sidebarOpen: boolean;
	onMobileClose?: () => void;
};

const PRESETS = [
	{ label: 'Indefinitely', icon: Infinity, timer: undefined },
	{ label: 'For 10 seconds', icon: Clock, timer: 10 },
	{ label: 'For 30 seconds', icon: Clock, timer: 30 },
	{ label: 'For 5 minutes', icon: Clock, timer: 300 },
];

export function BlockingControlDropdown({
	sidebarOpen,
	onMobileClose,
}: BlockingControlDropdownProps) {
	const { blocking, blockingUpdatedAt, refetchBlocking } = useClusterOverview();
	const [dropdownOpen, setDropdownOpen] = useState(false);
	const [customModalOpen, setCustomModalOpen] = useState(false);
	const [submitting, setSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const mode = blocking?.summary?.mode;
	const isEnabled = mode === 'enabled';
	const isLoading = blocking === undefined;
	const disabled = isLoading || submitting;

	const timers = blocking?.summary?.timers;
	const minSeconds = timers?.present ? timers.minSeconds ?? 0 : 0;
	const showCountdown = !isEnabled && minSeconds > 0;
	const remainingSeconds = useBlockingCountdown(
		minSeconds,
		blockingUpdatedAt,
		showCountdown,
	);

	// Refetch when countdown hits 0 so UI updates immediately when timer expires
	const hasRefetchedAtZero = useRef(false);
	useEffect(() => {
		if (remainingSeconds === 0 && showCountdown) {
			if (!hasRefetchedAtZero.current) {
				hasRefetchedAtZero.current = true;
				refetchBlocking();
			}
		} else if (remainingSeconds != null && remainingSeconds > 0) {
			hasRefetchedAtZero.current = false;
		}
	}, [remainingSeconds, showCountdown, refetchBlocking]);

	async function handleDisable(timer?: number) {
		setError(null);
		setSubmitting(true);
		try {
			await setClusterBlocking({ blocking: false, timer });
			setDropdownOpen(false);
			onMobileClose?.();
		} catch (err) {
			const msg = err instanceof Error ? err.message : 'Failed to disable blocking';
			setError(msg);
		} finally {
			setSubmitting(false);
		}
	}

	async function handleEnable() {
		setError(null);
		setSubmitting(true);
		try {
			await setClusterBlocking({ blocking: true });
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
