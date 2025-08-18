import { PiholeInitStatus } from '../types/initialization';
import PiholeManagementList from '../components/PiholeManagementList';
import { useInitializationStatus } from '../providers/InitializationStatusProvider';
import { usePiholes } from '../providers/PiholeProvider';
import AppCenteredPage from '@/components/Layout/AppCenteredPage';
import AppCard from '@/components/Layout/AppCard';
import styles from './SetupPiholes.module.scss';

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
		<AppCenteredPage className={styles.piholeCreationSetup}>
			<AppCard className={styles.setupCard}>
				<h2 className={styles.title}>Add one or more Pihole instances to get started</h2>
				<PiholeManagementList />
				<button onClick={handleClick}>
					{piholeNodes?.length ? 'Finish Setup' : 'Skip'}
				</button>
			</AppCard>
		</AppCenteredPage>
	);
}
