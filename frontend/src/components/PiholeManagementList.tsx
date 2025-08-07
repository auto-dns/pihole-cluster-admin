import { usePiholes } from '../providers/PiholeProvider';
import '../styles/components/pihole-management-list.scss';

export default function PiholeManagementList() {
	const { piholeNodes } = usePiholes();
	return (
		<div className='pihole-management-list'>
			{!piholeNodes?.length && (
				<div className='empty-state'>
					<h2>No Pi-hole instances yet</h2>
					<p>You’ll need at least one to get started.</p>
					<button className='primary'>Add First Node</button>
				</div>
			)}
			{!!piholeNodes?.length && (
				<div>
					<h2>Here are your nodes</h2>
				</div>
			)}
		</div>
	);
}
