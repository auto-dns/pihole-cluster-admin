import { PiholeNode } from '../../types/pihole';
import '../../styles/components/PiholeManagementList/pihole-table.scss';

type Props = {
	nodes: PiholeNode[];
	onRowClick: (node: PiholeNode) => void;
};

export default function PiholeTable({ nodes, onRowClick }: Props) {
	return (
		<div className='table-card'>
			<table className='app-table'>
				<caption className='sr-only'>Configured Pi-hole nodes</caption>
				<thead>
					<tr>
						<th>Name</th>
						<th>URL</th>
						<th>Description</th>
					</tr>
				</thead>
				<tbody>
					{nodes.map((node) => (
						<tr
							key={node.id}
							tabIndex={0}
							role='button'
							onClick={() => onRowClick(node)}
							onKeyDown={(e) =>
								(e.key === 'Enter' || e.key === ' ') && onRowClick(node)
							}
							className='clickable'
							aria-label={`Edit ${node.name}`}
							title='Click to edit'
						>
							<td className='truncate'>{node.name}</td>
							<td className='mono truncate'>
								{`${node.scheme}://${node.host}:${node.port}`}
							</td>
							<td className='truncate'>{node.description || '-'}</td>
						</tr>
					))}
				</tbody>
			</table>
		</div>
	);
}
