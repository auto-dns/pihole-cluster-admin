import { Navigate, Outlet } from "react-router";
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
		if (!isFullyInitialized(fullStatus)) {
			return <Navigate to="/setup/piholes" replace/>;
		}
		return <Navigate to="/" replace />;
	}

	// Unauthenticated

	// For mode "only when uninitialized"
	if (onlyWhenUninitialized) {
		// If user has been created, redirect to login
		if (publicStatus) {
			return <Navigate to="/login" replace />;
		}
		// When uninitialized, allow through
		return <Outlet/>
	} else if (!publicStatus) {
		return <Navigate to="/setup/user"/>
	}

	// Else, it doesn't matter what the initialization status is - allow through
	return <Outlet />;
}

export function UnprotectedRouteUninitialized() {
	return <UnprotectedRoute onlyWhenUninitialized={true}/>
}
