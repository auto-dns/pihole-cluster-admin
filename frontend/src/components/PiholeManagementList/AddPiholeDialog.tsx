import * as Dialog from '@radix-ui/react-dialog';
import PiholeNodeForm from './PiholeNodeForm';
import { usePiholes } from '../../providers/PiholeProvider';
import { useState } from 'react';
import { PiholeCreateBody } from '../../types/pihole';
import '../../styles/components/PiholeManagementList/pihole-management-modal.scss';

export function AddPiholeDialog({ trigger }: { trigger: React.ReactNode }) {
	const [open, setOpen] = useState(false);
	const [error, setError] = useState<Error | undefined>(undefined);
	const { addNode, addingNode } = usePiholes();

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
		<Dialog.Root open={open} onOpenChange={setOpen}>
			<Dialog.Trigger asChild>{trigger}</Dialog.Trigger>
			<Dialog.Portal>
				<Dialog.Overlay className='modal-overlay' />
				<Dialog.Content className='modal-content'>
					<Dialog.Title>Add Pi-hole</Dialog.Title>
					<PiholeNodeForm
						mode='create'
						submitting={addingNode}
						onCancel={() => setOpen(false)}
						onSubmit={handleSubmit}
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
