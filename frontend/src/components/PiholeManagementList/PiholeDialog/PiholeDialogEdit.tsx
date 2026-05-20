import { FormEvent, useEffect, useRef, useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { PiholeNodeForm } from '@/components/PiholeManagementList/PiholeNodeForm/PiholeNodeForm';
import { usePiholes } from '@/providers/PiholeProvider';
import { PiholePatchBody, rotatePiholePassword } from '@/lib/api/pihole';
import { PiholeNode } from '@/types/pihole';
import { PasswordField } from '@/components/PasswordField';
import { formatPiholeUrl } from '@/utils/urlUtils';
import styles from './PiholeDialog.module.scss';

interface Props {
	node: PiholeNode;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export function PiholeDialogEdit({ node, open, onOpenChange }: Props) {
	const { deleteNode, deletingNode, editNode, editingNode } = usePiholes();
	const [error, setError] = useState<Error | undefined>(undefined);
	const [dirty, setDirty] = useState(false);

	// Rotate password state
	const [showRotate, setShowRotate] = useState(false);
	const [newPwd, setNewPwd] = useState('');
	const [confirmPwd, setConfirmPwd] = useState('');
	const [rotating, setRotating] = useState(false);
	const [rotateErr, setRotateErr] = useState('');
	const [rotateOk, setRotateOk] = useState(false);
	const rotateOkTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

	useEffect(() => {
		return () => {
			if (rotateOkTimer.current) clearTimeout(rotateOkTimer.current);
		};
	}, []);

	function handleControlledOpen(next: boolean) {
		if (!next && dirty) {
			if (!window.confirm('Discard changes?')) return;
		}
		onOpenChange(next);
	}

	const nodeUrl = formatPiholeUrl({ scheme: node.scheme, host: node.host, port: node.port });
	function processDirtyStatus(name: string, url: string, password: string, description: string) {
		const dirty =
			name.trim() !== node.name ||
			url.trim() !== nodeUrl ||
			password.trim() !== '' ||
			description.trim() !== node.description;
		setDirty(dirty);
	}

	function validateFormStatus(
		name: string,
		url: string,
		password: string,
		description: string,
	): boolean {
		const valid =
			(name.trim() !== node.name.trim() && name.trim() !== '') ||
			(url.trim() !== nodeUrl && url.trim() !== '') ||
			password.trim() !== '' ||
			description.trim() !== node.description.trim();
		return valid;
	}

	function buildPatch(original: PiholeNode, updated: PiholePatchBody) {
		const patch: PiholePatchBody = {};

		// Name
		if (updated.name?.trim() && updated.name.trim() !== original.name.trim()) {
			patch.name = updated.name.trim();
		}

		// URL parts
		if (updated.scheme && updated.scheme !== original.scheme) {
			patch.scheme = updated.scheme;
		}
		if (updated.host && updated.host !== original.host) {
			patch.host = updated.host;
		}
		if (typeof updated.port === 'number' && updated.port !== original.port) {
			patch.port = updated.port;
		}

		// Password
		if (updated.password && updated.password.trim() !== '') {
			patch.password = updated.password;
		}

		// Description
		if (
			typeof updated.description === 'string' &&
			updated.description.trim() !== original.description.trim()
		) {
			patch.description = updated.description.trim();
		}

		return patch;
	}

	async function handleSubmit(id: number, updatedFull: PiholePatchBody) {
		try {
			const patch = buildPatch(node, updatedFull);
			if (Object.keys(patch).length === 0) {
				onOpenChange(false);
				return;
			}
			await editNode(id, patch);
			onOpenChange(false);
		} catch (err: unknown) {
			console.error(err);
			setError(err as Error);
		}
	}

	async function handleDelete() {
		if (!window.confirm(`Remove "${node.name}"? This can't be undone`)) return;
		try {
			await deleteNode(node.id);
			onOpenChange(false);
		} catch (err: unknown) {
			console.error(err);
			setError(err as Error);
		}
	}

	async function handleRotatePassword(e: FormEvent) {
		e.preventDefault();
		setRotateErr('');
		if (!newPwd.trim()) {
			setRotateErr('New password is required');
			return;
		}
		if (newPwd !== confirmPwd) {
			setRotateErr('Passwords do not match');
			return;
		}
		try {
			setRotating(true);
			await rotatePiholePassword(node.id, newPwd);
			setRotateOk(true);
			setNewPwd('');
			setConfirmPwd('');
			rotateOkTimer.current = setTimeout(() => setRotateOk(false), 3000);
		} catch (err: unknown) {
			setRotateErr((err as Error)?.message || 'Failed to rotate password');
		} finally {
			setRotating(false);
		}
	}

	return (
		<Dialog.Root open={open} onOpenChange={handleControlledOpen}>
			<Dialog.Portal>
				<Dialog.Overlay className={styles.overlay} />
				<Dialog.Content className={styles.content}>
					<Dialog.Title className={styles.title}>Edit Pi-hole</Dialog.Title>
					<PiholeNodeForm
						mode='edit'
						node={node}
						submitting={editingNode}
						deleting={deletingNode}
						onCancel={() => handleControlledOpen(false)}
						onSubmit={handleSubmit}
						onDelete={handleDelete}
						processDirtyStatus={processDirtyStatus}
						validateFormStatus={validateFormStatus}
					/>
					{error && <p className={styles.error}>{error.message}</p>}

					<div className={styles.rotateSection}>
						<button
							type='button'
							className='secondary'
							onClick={() => {
								setShowRotate((v) => !v);
								setRotateErr('');
								setRotateOk(false);
							}}
						>
							{showRotate ? 'Cancel password rotation' : 'Rotate Pi-hole password'}
						</button>
						{showRotate && (
							<form className={styles.rotateForm} onSubmit={handleRotatePassword}>
								<p className={styles.rotateHint}>
									Sets a new admin password directly on the Pi-hole node and updates the
									stored credential in cluster admin.
								</p>
								<PasswordField
									label='New password'
									value={newPwd}
									onChange={(e) => {
										setNewPwd(e.target.value);
										setRotateErr('');
										setRotateOk(false);
									}}
									autoComplete='new-password'
									disabled={rotating}
								/>
								<PasswordField
									label='Confirm new password'
									value={confirmPwd}
									onChange={(e) => {
										setConfirmPwd(e.target.value);
										setRotateErr('');
										setRotateOk(false);
									}}
									autoComplete='new-password'
									disabled={rotating}
								/>
								{rotateOk && (
									<p className={styles.rotateSuccess}>Password rotated successfully.</p>
								)}
								{rotateErr && <p className={styles.rotateErr}>{rotateErr}</p>}
								<div className={styles.rotateActions}>
									<button type='submit' className='danger' disabled={rotating}>
										{rotating ? 'Rotating…' : 'Rotate password'}
									</button>
								</div>
							</form>
						)}
					</div>

					<Dialog.Close asChild>
						<button className={styles.close} aria-label='Close'>
							✕
						</button>
					</Dialog.Close>
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>
	);
}
