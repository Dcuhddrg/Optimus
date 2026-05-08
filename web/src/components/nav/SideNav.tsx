import { useEffect, useState } from 'react';
import { NavLink, useLocation } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import { useI18n } from '../../locale/useI18n';
import styles from './SideNav.module.css';

export function SideNav() {
  const location = useLocation();
  const { hasMenu } = useAuth();
  const { t } = useI18n();
  const onSystemPage = location.pathname.startsWith('/system');
  const [collapsed, setCollapsed] = useState(false);
  const showUsers = hasMenu('system.users.menu');
  const showRoles = hasMenu('system.roles.menu');
  const showLDAP = hasMenu('system.ldap.menu');
  const showAudit = hasMenu('system.audit.menu');

  useEffect(() => {
    if (onSystemPage) {
      setCollapsed(false);
    }
  }, [location.pathname, onSystemPage]);

  const expanded = onSystemPage && !collapsed;

  return (
    <aside className={styles.sidebar}>
      <div className={styles.brand}>
        <div className={styles.brandMark}>O</div>
        <div>
          <p className={styles.brandEyebrow}>{t('shell.brandEyebrow')}</p>
          <h2 className={styles.brandTitle}>{t('shell.brandTitle')}</h2>
        </div>
      </div>

      <nav className={styles.nav} aria-label={t('shell.navLabel')}>
        <NavLink
          to="/"
          end
          className={({ isActive }) =>
            isActive ? styles.navItemActive : styles.navItem
          }
        >
          {t('shell.dashboard')}
        </NavLink>

        <div className={styles.group}>
          <button
            type="button"
            className={expanded ? styles.groupLabelActive : styles.groupLabel}
            onClick={() => setCollapsed((c) => !c)}
            aria-expanded={expanded}
          >
            <span
              className={expanded ? styles.chevronOpen : styles.chevron}
              aria-hidden="true"
            />
            {t('shell.system')}
          </button>
          {expanded ? (
            <div className={styles.groupChildren}>
              {showUsers ? (
                <NavLink
                  to="/system/users"
                  className={({ isActive }) =>
                    isActive ? styles.childItemActive : styles.childItem
                  }
                >
                  {t('shell.users')}
                </NavLink>
              ) : null}
              {showRoles ? (
                <NavLink
                  to="/system/roles"
                  className={({ isActive }) =>
                    isActive ? styles.childItemActive : styles.childItem
                  }
                >
                  {t('shell.roles')}
                </NavLink>
              ) : null}
              {showLDAP ? (
                <NavLink
                  to="/system/ldap"
                  className={({ isActive }) =>
                    isActive ? styles.childItemActive : styles.childItem
                  }
                >
                  {t('shell.ldap')}
                </NavLink>
              ) : null}
              {showAudit ? (
                <NavLink
                  to="/system/audit-logs"
                  className={({ isActive }) =>
                    isActive ? styles.childItemActive : styles.childItem
                  }
                >
                  {t('shell.auditLogs')}
                </NavLink>
              ) : null}
            </div>
          ) : null}
        </div>
      </nav>
    </aside>
  );
}
