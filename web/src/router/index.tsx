import { createBrowserRouter } from 'react-router-dom';

import { AppShell } from '../layout/AppShell';

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

function SystemLdapPlaceholder() {
  return (
    <section className="page-card">
      <p className="page-eyebrow">System / LDAP</p>
      <h1 className="page-title">LDAP configuration shell</h1>
      <p className="page-description">
        This route is reserved for the static LDAP settings page in the next
        implementation tasks.
      </p>
    </section>
  );
}

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppShell />,
    children: [
      {
        index: true,
        element: <DashboardPlaceholder />,
      },
      {
        path: 'system',
        children: [
          {
            path: 'ldap',
            element: <SystemLdapPlaceholder />,
          },
        ],
      },
    ],
  },
]);
