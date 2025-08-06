import { usePiholes } from '../providers/PiholeProvider';
import '../styles/pages/pihole-setup.scss';

export default function SetupPiholes() {
	const { piholeNodes } = usePiholes();
	return (
		<div className='pihole-creation-setup'>
			<div className='setup-card'>
				<h1>Add one or more Pihole instances to get started</h1>
			</div>
		</div>
	);
}
