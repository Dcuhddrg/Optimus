import { NavLink, useLocation } from 'react-router-dom';

import styles from './SideNav.module.css';

export function SideNav() {
  const location = useLocation();
  const systemExpanded = location.pathname.startsWith('/system');

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
            <NavLink
              to="/system/ldap"
              className={({ isActive }) =>
                isActive ? styles.childItemActive : styles.childItem
              }
            >
              LDAP
            </NavLink>
          </div>
        </div>
      </nav>
    </aside>
  );
}
