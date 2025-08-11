import * as Dialog from '@radix-ui/react-dialog';
import PiholeNodeForm from './PiholeNodeForm';
import { usePiholes } from '../../providers/PiholeProvider';
import { useState } from 'react';
import '../../styles/components/PiholeManagementList/pihole-management-modal.scss';
import { PiholePatchBody } from '../../lib/api/pihole';
import { PiholeNode } from '../../types/pihole';

interface Props {
	node: PiholeNode;
	trigger: React.ReactNode;
}

export function EditPiholeDialog({ node, trigger }: Props) {
	const [open, setOpen] = useState(false);
	const [dirty, setDirty] = useState(false);
	const [error, setError] = useState<Error | undefined>(undefined);
	const { editNode, editingNode } = usePiholes();

	function handleOpenChange(next: boolean) {
		if (!next && dirty) {
			if (!window.confirm('Discard changes?')) {
				return;
			}
		}
		setOpen(next);
	}

	async function handleSubmit(id: number, node: PiholePatchBody) {
		try {
			await editNode(id, node);
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
					<Dialog.Title>Edit Pi-hole</Dialog.Title>
					<PiholeNodeForm
						mode='edit'
						node={node}
						submitting={editingNode}
						onCancel={() => handleOpenChange(false)}
						onSubmit={handleSubmit}
						onDirtyChange={setDirty}
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
