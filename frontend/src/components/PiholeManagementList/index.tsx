import { useState } from 'react';
import { usePiholes } from '../../providers/PiholeProvider';
import { PiholeNode } from '../../types/pihole';
import { PiholeDialogAdd } from './PiholeDialogAdd';
import { PiholeDialogEdit } from './PiholeDialogEdit';
import PiholeTable from './PiholeTable';
import PiholeCardList from './PiholeCardList';
import styles from './index.module.scss';

export default function PiholeManagementList() {
	const { piholeNodes } = usePiholes();
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
						<PiholeTable nodes={piholeNodes} onRowClick={(node) => setEditing(node)} />
					</div>

					{/* Mobile cards */}
					<div className={styles.mobileOnly}>
						<PiholeCardList
							nodes={piholeNodes}
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
