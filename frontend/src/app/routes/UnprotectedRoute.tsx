import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "../AuthProvider";
import { useInitializationStatus } from "../InitializationStatusProvider";
import { isFullyInitialized } from "../../utils/initHelpers";

interface UnprotectedRouteProps {
	/** If true, only accessible when the app is uninitialized */
	onlyWhenUninitialized?: boolean;
}

/**
 * Allows access to public routes but redirects authenticated users or disallows
 * when system initialization status changes.
 */
export function UnprotectedRoute({ onlyWhenUninitialized = false }: UnprotectedRouteProps) {
	const { user, loading: authLoading } = useAuth();
	const { publicStatus, fullStatus, loading: initLoading } = useInitializationStatus();

	if (authLoading || initLoading) {
		return <div>Loading...</div>;
	}

  	// Authenticated user → send to home
	if (user) {
		// If user is authenticated but setup not complete, redirect to pihole setup
		console.log('user logged in');
		if (!isFullyInitialized(fullStatus)) {
			console.log('redirect to piholes setup');
			return <Navigate to="/setup/piholes" replace/>;
		}
		return <Navigate to="/" replace />;
	}

	if (!publicStatus && !onlyWhenUninitialized) {
		console.log('redirecting to initialization');
		return <Navigate to='/initialization' replace/>;
	}

  	// Unauthenticated but system already initialized and route is only for uninitialized state
	if (onlyWhenUninitialized && publicStatus) {
		console.log('redirecting to login');
		return <Navigate to="/login" replace />;
	}

	console.log('outlet');
	return <Outlet />;
}

export function UnprotectedRouteUninitialized() {
	console.log('unprotected-uninitialized');
	return <UnprotectedRoute onlyWhenUninitialized={true}/>
}
