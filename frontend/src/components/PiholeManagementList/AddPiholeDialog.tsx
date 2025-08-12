import * as Dialog from '@radix-ui/react-dialog';
import PiholeNodeForm from './PiholeNodeForm';
import { usePiholes } from '../../providers/PiholeProvider';
import { useState } from 'react';
import '../../styles/components/PiholeManagementList/pihole-management-modal.scss';
import { PiholeCreateBody } from '../../lib/api/pihole';

export function AddPiholeDialog({ trigger }: { trigger: React.ReactNode }) {
	const [open, setOpen] = useState(false);
	const [dirty, setDirty] = useState(false);
	const [error, setError] = useState<Error | undefined>(undefined);
	const { addNode, addingNode } = usePiholes();

	function handleOpenChange(next: boolean) {
		if (!next && dirty) {
			if (!window.confirm('Discard changes?')) {
				return;
			}
		}
		setOpen(next);
	}

	function processDirtyStatus(name: string, url: string, password: string, description: string) {
		const dirty =
			name.trim() !== '' ||
			url.trim() !== '' ||
			password.trim() !== '' ||
			description.trim() !== '';
		setDirty(dirty);
	}

	async function handleSubmit(node: PiholeCreateBody) {
		try {
			await addNode(node);
			setOpen(false);
		} catch (err: unknown) {
			console.error(err);
			setError(err as Error);
		}
	}

	return (
		<Dialog.Root open={open} onOpenChange={handleOpenChange}>
			<Dialog.Trigger asChild>{trigger}</Dialog.Trigger>
			<Dialog.Portal>
				<Dialog.Overlay className='modal-overlay' />
				<Dialog.Content className='modal-content'>
					<Dialog.Title>Add Pi-hole</Dialog.Title>
					<PiholeNodeForm
						mode='create'
						submitting={addingNode}
						onCancel={() => handleOpenChange(false)}
						onSubmit={handleSubmit}
						processDirtyStatus={processDirtyStatus}
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
