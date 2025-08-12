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

	async function handleSubmit(id: number, node: PiholePatchBody) {
		try {
			await editNode(id, node);
			onOpenChange(false);
		} catch (err: unknown) {
			console.error(err);
			setError(err as Error);
		}
	}

	async function handleDelete() {
		if (!window.confirm('Remove this node?')) return;
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
