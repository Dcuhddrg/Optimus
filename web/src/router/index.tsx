import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom';

import { AuthProvider } from '../auth/AuthContext';
import { ProjectProvider } from '../auth/ProjectContext';
import { RequireAuth } from '../auth/RequireAuth';
import { AppShell } from '../layout/AppShell';
import { LanguageProvider } from '../locale/LanguageProvider';
import { LoginPage } from '../view/auth/LoginPage';
import { CloudAccountsPage } from '../view/cloud/CloudAccountsPage';
import { CloudAssetsPage } from '../view/cloud/CloudAssetsPage';
import { DashboardPage } from '../view/dashboard/DashboardPage';
import { ProjectMembersPage } from '../view/project/ProjectMembersPage';
import { ProjectsPage } from '../view/project/ProjectsPage';
import { ServicePortalPage } from '../view/service/ServicePortalPage';
import { TicketDetailPage } from '../view/service/TicketDetailPage';
import { TicketListPage } from '../view/service/TicketListPage';
import { ForbiddenPage } from '../view/system/ForbiddenPage';
import { SystemAuditLogsPage } from '../view/system/SystemAuditLogsPage';
import { SystemLdapPage } from '../view/system/SystemLdapPage';
import { SystemRolesPage } from '../view/system/SystemRolesPage';
import { SystemUsersPage } from '../view/system/SystemUsersPage';

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
            element: (
              <ProjectProvider>
                <AppShell />
              </ProjectProvider>
            ),
            children: [
              {
                index: true,
                element: <DashboardPage />,
              },
              {
                path: 'forbidden',
                element: <ForbiddenPage />,
              },
              {
                path: 'projects',
                element: <ProjectsPage />,
              },
              {
                path: 'service-portal',
                element: <ServicePortalPage />,
              },
              {
                path: 'service-portal/tickets',
                element: <TicketListPage />,
              },
              {
                path: 'service-portal/tickets/:id',
                element: <TicketDetailPage />,
              },
              {
                path: 'projects/:code/members',
                element: <ProjectMembersPage />,
              },
              {
                path: 'cloud',
                element: <Navigate to="/cloud/accounts" replace />,
              },
              {
                path: 'cloud/accounts',
                element: <CloudAccountsPage />,
              },
              {
                path: 'cloud/assets',
                element: <CloudAssetsPage />,
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
