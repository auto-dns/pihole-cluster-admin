import { useState } from 'react';
import { usePiholes } from '@/providers/PiholeProvider';
import { useClusterHealth } from '@/hooks/useClusterHealth';
import { PiholeNode } from '@/types/pihole';
import { PiholeDialogAdd } from '@/components/PiholeManagementList/PiholeDialog/PiholeDialogAdd';
import { PiholeDialogEdit } from '@/components/PiholeManagementList/PiholeDialog/PiholeDialogEdit';
import { PiholeTable } from '@/components/PiholeManagementList/PiholeTable/PiholeTable';
import { PiholeCardList } from '@/components/PiholeManagementList/PiholeCardList';
import styles from './PiholeManagementList.module.scss';

export function PiholeManagementList() {
	const { piholeNodes } = usePiholes();
	const { nodeHealthById, isFresh } = useClusterHealth();
	const [editing, setEditing] = useState<PiholeNode | undefined>(undefined);

	return (
		<div className={styles.managementList}>
			{!piholeNodes?.length && (
				<div className={styles.emptyState}>
					<h2>No Pi-hole instances yet</h2>
					<p>You’ll need at least one to get started.</p>
					<PiholeDialogAdd
						trigger={<button className={styles.primary}>Add first node</button>}
					/>
				</div>
			)}

			{!!piholeNodes?.length && (
				<>
					<div className={styles.header}>
						<h2>Here are your nodes</h2>
						<div className={styles.toolbar}>
							<PiholeDialogAdd
								trigger={<button className={styles.primary}>Add node</button>}
							/>
						</div>
					</div>

					{/* Desktop table */}
					<div className={styles.tableWrap}>
						<PiholeTable
							nodes={piholeNodes}
							nodeHealthById={nodeHealthById}
							isFresh={isFresh}
							onRowClick={(node) => setEditing(node)}
						/>
					</div>

					{/* Mobile cards */}
					<div className={styles.mobileOnly}>
						<PiholeCardList
							nodes={piholeNodes}
							nodeHealthById={nodeHealthById}
							isFresh={isFresh}
							onCardClick={(node) => setEditing(node)}
						/>
					</div>

					{editing && (
						<PiholeDialogEdit
							open={!!editing}
							node={editing}
							onOpenChange={(next) => {
								if (!next) setEditing(undefined);
							}}
						/>
					)}
				</>
			)}
		</div>
	);
}
