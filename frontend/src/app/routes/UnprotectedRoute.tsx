import { Navigate, Outlet } from 'react-router';
import { useAuth } from '../AuthProvider';
import { useInitializationStatus } from '../InitializationStatusProvider';
import { isFullyInitialized } from '../../utils/initHelpers';

interface UnprotectedRouteProps {
  /** If true, only accessible when the app is uninitialized */
  onlyWhenUninitialized?: boolean;
}

export function UnprotectedRoute({ onlyWhenUninitialized = false }: UnprotectedRouteProps) {
  const { user, loading: authLoading } = useAuth();
  const { publicStatus, fullStatus, loading: initLoading } = useInitializationStatus();

  const isLoading = authLoading || initLoading;

  // Redirect authenticated users
  if (user) {
    if (!isFullyInitialized(fullStatus)) {
      return <Navigate to="/setup/piholes" replace />;
    }
    return <Navigate to="/" replace />;
  }

  // Redirect initialized apps away from uninitialized-only routes
  if (onlyWhenUninitialized) {
    if (publicStatus) {
      return <Navigate to="/login" replace />;
    }
    return (
      <div className="route-container">
        {isLoading && <div className="loading-overlay">Loading...</div>}
        <Outlet />
      </div>
    );
  } else if (!publicStatus) {
    return <Navigate to="/setup/user" replace />;
  }

  // Default: allow access
  return (
    <div className="route-container">
      {isLoading && <div className="loading-overlay">Loading...</div>}
      <Outlet />
    </div>
  );
}

export function UnprotectedRouteUninitialized() {
  return <UnprotectedRoute onlyWhenUninitialized={true} />;
}
