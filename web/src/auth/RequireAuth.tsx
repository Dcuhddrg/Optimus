import { Navigate, Outlet, useLocation } from 'react-router-dom';

import { useAuth } from './AuthContext';

export function RequireAuth() {
  const location = useLocation();
  const { token, loading } = useAuth();

  if (loading) {
    return <div className="page-card">Loading session...</div>;
  }

  if (!token) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return <Outlet />;
}
