import { Navigate } from 'react-router';
import { useInitializationStatus } from '../app/InitializationStatusProvider';

export default function Setup() {
	const init = useInitializationStatus();

	if (!init.fullStatus) {
		return <div>Loading...</div>;
	}

	if (init.fullStatus?.piholeStatus === 'UNINITIALIZED') {
		return <Navigate to="/setup/piholes" replace />;
	}

	return <Navigate to="/" replace />;
}
