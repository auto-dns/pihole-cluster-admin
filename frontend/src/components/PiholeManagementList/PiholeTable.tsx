import { PiholeNode } from '../../types/pihole';
import { useClusterHealth } from '../../hooks/useClusterHealth';
import PiholeStatusLight from '../PiholeStatusLight';
import styles from './PiholeTable.module.scss';
import classNames from 'classnames';

type Props = {
	nodes: PiholeNode[];
	onRowClick: (node: PiholeNode) => void;
};

export default function PiholeTable({ nodes, onRowClick }: Props) {
	const { nodeHealthById, nodeHealthIsFresh } = useClusterHealth();

	return (
		<div className={styles.tableCard}>
			<table className={styles.table}>
				<caption className='sr-only'>Configured Pi-hole nodes</caption>
				<thead>
					<tr>
						<th style={{ width: 40 }} aria-label='Status' />
						<th>Name</th>
						<th>URL</th>
						<th>Description</th>
					</tr>
				</thead>
				<tbody>
					{nodes.map((node) => {
						const nodeHealth = nodeHealthById.get(node.id);
						return (
							<tr
								key={node.id}
								tabIndex={0}
								role='button'
								onClick={() => onRowClick(node)}
								onKeyDown={(e) =>
									(e.key === 'Enter' || e.key === ' ') && onRowClick(node)
								}
								className={styles.clickable}
								aria-label={`Edit ${node.name}`}
								title='Click to edit'
							>
								<td>
									<PiholeStatusLight
										node={node}
										health={nodeHealth}
										fresh={nodeHealthIsFresh}
									/>
								</td>
								<td className='truncate'>{node.name}</td>
								<td className={classNames(styles.mono, 'truncate')}>
									{`${node.scheme}://${node.host}:${node.port}`}
								</td>
								<td className='truncate'>{node.description || '-'}</td>
							</tr>
						);
					})}
				</tbody>
			</table>
		</div>
	);
}
