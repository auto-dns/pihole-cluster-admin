import { usePiholes } from '../../providers/PiholeProvider';
import { AddPiholeDialog } from './AddPiholeDialog';
import '../../styles/components/PiholeManagementList/pihole-management-list.scss';
import PiholeTable from './PiholeTable';

export default function PiholeManagementList() {
	const { piholeNodes } = usePiholes();

	return (
		<div className='pihole-management-list'>
			{!piholeNodes?.length && (
				<div className='empty-state'>
					<h2>No Pi-hole instances yet</h2>
					<p>You’ll need at least one to get started.</p>
					<AddPiholeDialog
						trigger={<button className='primary'>Add first node</button>}
					/>
				</div>
			)}

			{!!piholeNodes?.length && (
				<>
					<div className='header'>
						<h2>Here are your nodes</h2>
						<div className='toolbar'>
							<AddPiholeDialog
								trigger={<button className='primary add-btn'>Add node</button>}
							/>
						</div>
					</div>
					<div className='table-wrap'>
						<PiholeTable nodes={piholeNodes} onRowClick={(node) => console.log(node)} />
					</div>
				</>
			)}
		</div>
	);
}
