import PiholeManagementList from '../components/PiholeManagementList';
import { usePiholes } from '../providers/PiholeProvider';
import '../styles/pages/pihole-setup.scss';

export default function SetupPiholes() {
	const { piholeNodes } = usePiholes();

	return (
		<div className='pihole-creation-setup'>
			<div className='setup-card'>
				<h2 className='setup-step-title'>
					Add one or more Pihole instances to get started
				</h2>
				<PiholeManagementList />
				<button>{piholeNodes?.length ? 'Finish Setup' : 'Skip'}</button>
			</div>
		</div>
	);
}
