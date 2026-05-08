import { createBrowserRouter, Outlet } from 'react-router-dom';

import { AuthProvider } from '../auth/AuthContext';
import { RequireAuth } from '../auth/RequireAuth';
import { AppShell } from '../layout/AppShell';
import { LanguageProvider } from '../locale/LanguageProvider';
import { LoginPage } from '../view/auth/LoginPage';
import { ForbiddenPage } from '../view/system/ForbiddenPage';
import { SystemAuditLogsPage } from '../view/system/SystemAuditLogsPage';
import { SystemLdapPage } from '../view/system/SystemLdapPage';
import { SystemRolesPage } from '../view/system/SystemRolesPage';
import { SystemUsersPage } from '../view/system/SystemUsersPage';

function DashboardPlaceholder() {
  return (
    <section className="page-card">
      <p className="page-eyebrow">Dashboard</p>
      <h1 className="page-title">Ops workspace is being prepared.</h1>
      <p className="page-description">
        The first dashboard experience will be added after the shell and LDAP
        management page are finalized.
      </p>
    </section>
  );
}

function AuthRoot() {
  return (
    <LanguageProvider>
      <AuthProvider>
        <Outlet />
      </AuthProvider>
    </LanguageProvider>
  );
}

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AuthRoot />,
    children: [
      {
        path: 'login',
        element: <LoginPage />,
      },
      {
        element: <RequireAuth />,
        children: [
          {
            element: <AppShell />,
            children: [
              {
                index: true,
                element: <DashboardPlaceholder />,
              },
              {
                path: 'forbidden',
                element: <ForbiddenPage />,
              },
              {
                path: 'system',
                children: [
                  {
                    path: 'users',
                    element: <SystemUsersPage />,
                  },
                  {
                    path: 'roles',
                    element: <SystemRolesPage />,
                  },
                  {
                    path: 'ldap',
                    element: <SystemLdapPage />,
                  },
                  {
                    path: 'audit-logs',
                    element: <SystemAuditLogsPage />,
                  },
                ],
              },
            ],
          },
        ],
      },
    ],
  },
]);
