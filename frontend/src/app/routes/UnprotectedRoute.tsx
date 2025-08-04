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
	const { user } = useAuth();
	const { publicStatus, fullStatus } = useInitializationStatus();

  	// Authenticated user → send to home
	if (user) {
		// If user is authenticated but setup not complete, redirect to pihole setup
		if (!isFullyInitialized(fullStatus)) {
			return <Navigate to="/setup/piholes" replace />;
		}
		return <Navigate to="/" replace />;
	}

	if (!publicStatus && !onlyWhenUninitialized) {
		return <Navigate to='/setup/user' replace/>;
	}

  	// Unauthenticated but system already initialized and route is only for uninitialized state
	if (onlyWhenUninitialized && publicStatus) {
		return <Navigate to="/login" replace />;
	}

	return <Outlet />;
}

export function UnprotectedRouteUninitialized() {
	return <UnprotectedRoute onlyWhenUninitialized={true}/>
}
