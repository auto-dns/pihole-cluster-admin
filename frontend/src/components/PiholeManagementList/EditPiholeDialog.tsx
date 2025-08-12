import * as Dialog from '@radix-ui/react-dialog';
import PiholeNodeForm from './PiholeNodeForm';
import { usePiholes } from '../../providers/PiholeProvider';
import { useState } from 'react';
import '../../styles/components/PiholeManagementList/pihole-management-modal.scss';
import { PiholePatchBody } from '../../lib/api/pihole';
import { PiholeNode } from '../../types/pihole';
import { formatPiholeUrl } from '../../utils/urlUtils';

interface Props {
	node: PiholeNode;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export function EditPiholeDialog({ node, open, onOpenChange }: Props) {
	const { deleteNode, deletingNode, editNode, editingNode } = usePiholes();
	const [error, setError] = useState<Error | undefined>(undefined);
	const [dirty, setDirty] = useState(false);

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

	return (
		<Dialog.Root open={open} onOpenChange={handleControlledOpen}>
			<Dialog.Portal>
				<Dialog.Overlay className='modal-overlay' />
				<Dialog.Content className='modal-content'>
					<Dialog.Title>Edit Pi-hole</Dialog.Title>
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
					{error && <p>{error.message}</p>}
					<Dialog.Close asChild>
						<button className='modal-close' aria-label='Close'>
							✕
						</button>
					</Dialog.Close>
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>
	);
}
