import { useState, useEffect } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { setClusterBlocking } from '@/lib/api/blocking';
import type { ClusterBlockingState } from '@/types/blocking';
import styles from './CustomDisableModal.module.scss';

type CustomDisableModalProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onApplyState?: (state: ClusterBlockingState, opts?: { requestedTimerSeconds: number }) => void;
	onSuccess?: () => void;
	onError?: (err: Error) => void;
};

export function CustomDisableModal({
	open,
	onOpenChange,
	onApplyState,
	onSuccess,
	onError,
}: CustomDisableModalProps) {
	const [value, setValue] = useState(60);
	const [unit, setUnit] = useState<'secs' | 'mins'>('mins');
	const [submitting, setSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		if (open) {
			setValue(60);
			setUnit('mins');
			setError(null);
		}
	}, [open]);

	function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		const num = Number.parseInt(String(value), 10);
		if (!Number.isFinite(num) || num <= 0) {
			setError('Enter a positive number');
			return;
		}
		const seconds = unit === 'secs' ? num : num * 60;
		setError(null);
		setSubmitting(true);
		setClusterBlocking({ blocking: false, timer: seconds })
			.then((nextState) => {
				onApplyState?.(nextState, { requestedTimerSeconds: seconds });
				onOpenChange(false);
				onSuccess?.();
			})
			.catch((err: Error) => {
				setError(err.message || 'Failed to disable blocking');
				onError?.(err);
			})
			.finally(() => {
				setSubmitting(false);
			});
	}

	function handleOpenChange(next: boolean) {
		if (!next) {
			setError(null);
		}
		onOpenChange(next);
	}

	return (
		<Dialog.Root open={open} onOpenChange={handleOpenChange}>
			<Dialog.Portal>
				<Dialog.Overlay className={styles.overlay} />
				<Dialog.Content className={styles.content}>
					<Dialog.Title className={styles.title}>Custom disable timeout</Dialog.Title>
					<form onSubmit={handleSubmit} className={styles.form}>
						<div className={styles.inputRow}>
							<input
								type="number"
								className={styles.input}
								value={value}
								onChange={(e) => {
									const v = e.target.valueAsNumber;
									setValue(Number.isFinite(v) ? v : 0);
								}}
								min={1}
								max={unit === 'mins' ? 1440 : 86400}
								disabled={submitting}
								autoFocus
								aria-label="Duration value"
							/>
							<div className={styles.unitToggle} role="group" aria-label="Time unit">
								<button
									type="button"
									className={`${styles.unitButton} ${unit === 'secs' ? styles.active : ''}`}
									onClick={() => setUnit('secs')}
									disabled={submitting}
								>
									Secs
								</button>
								<button
									type="button"
									className={`${styles.unitButton} ${unit === 'mins' ? styles.active : ''}`}
									onClick={() => setUnit('mins')}
									disabled={submitting}
								>
									Mins
								</button>
							</div>
						</div>
						{error && <p className={styles.error}>{error}</p>}
						<div className={styles.actions}>
							<Dialog.Close asChild>
								<button
									type="button"
									className={styles.unitButton}
									disabled={submitting}
									style={{ background: 'var(--btn-cancel-bg)', border: '1px solid var(--btn-cancel-border)' }}
								>
									Close
								</button>
							</Dialog.Close>
							<button
								type="submit"
								disabled={submitting}
								style={{
									background: 'var(--btn-primary-bg)',
									color: 'var(--btn-primary-text)',
									padding: 'var(--input-padding) 1rem',
									borderRadius: 'var(--input-radius)',
									border: 'none',
									cursor: submitting ? 'not-allowed' : 'pointer',
									fontSize: '0.9rem',
								}}
							>
								{submitting ? 'Submitting…' : 'Submit'}
							</button>
						</div>
					</form>
					<Dialog.Close asChild>
						<button className={styles.close} aria-label="Close">
							✕
						</button>
					</Dialog.Close>
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>
	);
}
