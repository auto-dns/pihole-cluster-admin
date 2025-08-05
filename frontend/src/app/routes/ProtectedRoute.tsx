import { Navigate, Outlet, useLocation } from "react-router";
import { useAuth } from "../AuthProvider";
import { useInitializationStatus } from "../InitializationStatusProvider";
import { isFullyInitialized } from "../..//utils/initHelpers";

interface ProtectedRouteProps {
  /** Require the system to be fully initialized before allowing access */
    requireFullInit?: boolean;
}

/**
 * Protects routes that require authentication.
 * Optionally also requires full initialization.
 */
export function ProtectedRoute({ requireFullInit = false }: ProtectedRouteProps) {
    const { user, loading: authLoading } = useAuth();
    const { publicStatus, fullStatus, loading: initLoading } = useInitializationStatus();
    const location = useLocation();

    if (authLoading || initLoading) {
        return <div>Loading...</div>;
    }

    // --- Case: Unauthenticated user ---
    if (!user) {
        if (!publicStatus) {
            return <Navigate to="/setup/user" replace state={{ from: location }} />;
        }
        return <Navigate to="/login" replace state={{ from: location }} />;
    }

    // --- Case: Authenticated but not fully initialized ---
    if (requireFullInit && !isFullyInitialized(fullStatus)) {
        return <Navigate to="/setup/piholes" replace />;
    }

    return <Outlet />;
}

export function ProtectedRouteFullInit() {
    return <ProtectedRoute requireFullInit={true}/>
}
