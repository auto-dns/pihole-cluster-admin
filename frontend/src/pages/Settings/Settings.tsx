import { PiholeManagementList } from '@/components/PiholeManagementList';
import styles from './Settings.module.scss';

export function Settings() {
	return (
		<div className={styles.settingsPage}>
			<PiholeManagementList title='Pi-hole Nodes' />
		</div>
	);
}
