import { PiholeInitStatus } from '../types/initialization';
import PiholeManagementList from '../components/PiholeManagementList';
import { useInitializationStatus } from '../providers/InitializationStatusProvider';
import { usePiholes } from '../providers/PiholeProvider';
import '../styles/pages/pihole-setup.scss';

export default function SetupPiholes() {
	const { updatePiholeInitStatus } = useInitializationStatus();
	const { piholeNodes } = usePiholes();

	function handleClick() {
		updateInitStatus();
	}

	async function updateInitStatus() {
		if (piholeNodes.length) {
			await updatePiholeInitStatus(PiholeInitStatus.ADDED, true);
		} else {
			await updatePiholeInitStatus(PiholeInitStatus.SKIPPED, true);
		}
	}

	return (
		<div className='pihole-creation-setup'>
			<div className='setup-card'>
				<h2 className='setup-step-title'>
					Add one or more Pihole instances to get started
				</h2>
				<PiholeManagementList />
				<button onClick={handleClick}>
					{piholeNodes?.length ? 'Finish Setup' : 'Skip'}
				</button>
			</div>
		</div>
	);
}
