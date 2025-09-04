import { Navigate } from 'react-router';

export function UnhandledRoute() {
	return <Navigate to='/' replace />;
}
