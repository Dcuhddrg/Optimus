import { Navigate, Outlet, useLocation } from 'react-router-dom';

import { useAuth } from './AuthContext';
import { useI18n } from '../locale/useI18n';

export function RequireAuth() {
  const location = useLocation();
  const { token, current, loading, sessionError, retrySession } = useAuth();
  const { t } = useI18n();

  if (loading) {
    return <div className="page-card" role="status">{t('auth.loadingSession')}</div>;
  }

  if (!token) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  if (sessionError) {
    return <div className="page-card" role="alert">
      <p>{t(sessionError === 'forbidden' ? 'auth.sessionForbidden' : 'auth.sessionUnavailable')}</p>
      <button type="button" onClick={retrySession}>{t('auth.retrySession')}</button>
    </div>;
  }

  if (!current) {
    return <div className="page-card" role="status">{t('auth.loadingSession')}</div>;
  }

  return <Outlet />;
}
