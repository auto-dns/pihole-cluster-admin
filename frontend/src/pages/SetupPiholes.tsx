import { usePiholes } from '../providers/PiholeProvider';
import '../styles/pages/pihole-setup.scss';

export default function SetupPiholes() {
	const { piholeNodes } = usePiholes();

	function handleClick() {
		if (piholeNodes.length) {
			console.log('Finish setup');
		} else {
			console.log('Skip pihole setup');
		}
	}

	return (
		<div className='pihole-creation-setup'>
			<div className='setup-card'>
				<h1>Add one or more Pihole instances to get started</h1>
				<button onClick={handleClick}>
					{piholeNodes.length ? 'Finish Setup' : 'Skip'}
				</button>
			</div>
		</div>
	);
}
