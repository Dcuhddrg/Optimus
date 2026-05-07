import { NavLink, useLocation } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import styles from './SideNav.module.css';

export function SideNav() {
  const location = useLocation();
  const { hasMenu } = useAuth();
  const systemExpanded = location.pathname.startsWith('/system');
  const showUsers = hasMenu('system.users.menu');
  const showRoles = hasMenu('system.roles.menu');
  const showLDAP = hasMenu('system.ldap.menu');
  const showAudit = hasMenu('system.audit.menu');

  return (
    <aside className={styles.sidebar}>
      <div className={styles.brand}>
        <div className={styles.brandMark}>O</div>
        <div>
          <p className={styles.brandEyebrow}>Platform</p>
          <h2 className={styles.brandTitle}>Optimus</h2>
        </div>
      </div>

      <nav className={styles.nav} aria-label="Primary">
        <NavLink
          to="/"
          end
          className={({ isActive }) =>
            isActive ? styles.navItemActive : styles.navItem
          }
        >
          Dashboard
        </NavLink>

        <div className={styles.group}>
          <div className={systemExpanded ? styles.groupLabelActive : styles.groupLabel}>
            System
          </div>
          <div className={styles.groupChildren}>
            {showUsers ? (
              <NavLink
                to="/system/users"
                className={({ isActive }) =>
                  isActive ? styles.childItemActive : styles.childItem
                }
              >
                Users
              </NavLink>
            ) : null}
            {showRoles ? (
              <NavLink
                to="/system/roles"
                className={({ isActive }) =>
                  isActive ? styles.childItemActive : styles.childItem
                }
              >
                Roles
              </NavLink>
            ) : null}
            {showLDAP ? (
              <NavLink
                to="/system/ldap"
                className={({ isActive }) =>
                  isActive ? styles.childItemActive : styles.childItem
                }
              >
                LDAP
              </NavLink>
            ) : null}
            {showAudit ? (
              <NavLink
                to="/system/audit-logs"
                className={({ isActive }) =>
                  isActive ? styles.childItemActive : styles.childItem
                }
              >
                Audit Logs
              </NavLink>
            ) : null}
          </div>
        </div>
      </nav>
    </aside>
  );
}
