import { Navigate, Outlet, useLocation } from 'react-router';
import { useAuth } from '../../providers/AuthProvider';
import { useInitializationStatus } from '../../providers/InitializationStatusProvider';
import { isFullyInitialized } from '../..//utils/initHelpers';

interface ProtectedRouteProps {
	/** Require the system to be fully initialized before allowing access */
	allowUninitialized?: boolean;
	requireUninitialized?: boolean;
}

/**
 * Protects routes that require authentication.
 * Optionally also requires full initialization.
 */
export function ProtectedRoute({
	allowUninitialized = false,
	requireUninitialized = true,
}: ProtectedRouteProps) {
	const { user, loadingUser } = useAuth();
	const { publicStatus, fullStatus, fullLoading } = useInitializationStatus();
	const location = useLocation();

	if (loadingUser || fullLoading) {
		return <div>Loading...</div>;
	}

	// --- Case: Unauthenticated user ---
	if (!user) {
		if (!publicStatus) {
			return <Navigate to='/setup/user' replace state={{ from: location }} />;
		}
		return <Navigate to='/login' replace state={{ from: location }} />;
	}

	if (!allowUninitialized && !isFullyInitialized(fullStatus)) {
		return <Navigate to='/setup/piholes' replace />;
	}
	if (requireUninitialized && isFullyInitialized(fullStatus)) {
		return <Navigate to='/' replace />;
	}

	return <Outlet />;
}

export function ProtectedRouteFullInit() {
	return <ProtectedRoute allowUninitialized={true} requireUninitialized={false} />;
}

export function ProtectedRouteUninitialized() {
	return <ProtectedRoute allowUninitialized={true} requireUninitialized={true} />;
}
