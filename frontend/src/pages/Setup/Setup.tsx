import { Outlet } from 'react-router';
import { useInitializationStatus } from '../../providers/InitializationStatusProvider';

export default function Setup() {
	const { fullStatus } = useInitializationStatus();

	if (!fullStatus) {
		return <div>Loading...</div>;
	}

	return <Outlet />;
}
