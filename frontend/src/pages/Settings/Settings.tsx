import { PiholeManagementList } from '@/components/PiholeManagementList';
import styles from './Settings.module.scss';

export function Settings() {
	return (
		<div className={styles.settingsPage}>
			<section>
				<h2 className={styles.sectionTitle}>Pi-hole Nodes</h2>
				<PiholeManagementList />
			</section>
		</div>
	);
}
