import { useEffect, useState } from 'react';
import { NavLink, useLocation } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import { useI18n } from '../../locale/useI18n';
import styles from './SideNav.module.css';

type SystemItem = {
  to: string;
  label: string;
};

function isSystemItem(item: SystemItem | null): item is SystemItem {
  return item !== null;
}

export function SideNav() {
  const location = useLocation();
  const { hasMenu } = useAuth();
  const { t } = useI18n();
  const onSystemPage = location.pathname.startsWith('/system');
  const [collapsed, setCollapsed] = useState(() => !onSystemPage);
  const showUsers = hasMenu('system.users.menu');
  const showRoles = hasMenu('system.roles.menu');
  const showLDAP = hasMenu('system.ldap.menu');
  const showAudit = hasMenu('system.audit.menu');

  useEffect(() => {
    if (onSystemPage) {
      setCollapsed(false);
    }
  }, [location.pathname, onSystemPage]);

  const expanded = !collapsed;
  const systemItems = [
    showUsers ? { to: '/system/users', label: t('shell.users') } : null,
    showRoles ? { to: '/system/roles', label: t('shell.roles') } : null,
    showLDAP ? { to: '/system/ldap', label: t('shell.ldap') } : null,
    showAudit ? { to: '/system/audit-logs', label: t('shell.auditLogs') } : null,
  ].filter(isSystemItem);
  const hasSystemItems = systemItems.length > 0;
  const systemSectionClassName = [
    styles.section,
    expanded ? styles.sectionOpen : '',
    onSystemPage ? styles.sectionActive : '',
  ]
    .filter(Boolean)
    .join(' ');

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
            [styles.navEntry, isActive ? styles.navEntryActive : '']
              .filter(Boolean)
              .join(' ')
          }
        >
          {t('shell.dashboard')}
        </NavLink>

        {hasSystemItems ? (
          <section className={systemSectionClassName}>
            <button
              type="button"
              className={[styles.navEntry, styles.sectionButton].join(' ')}
              onClick={() => {
                if (onSystemPage) {
                  setCollapsed(false);
                  return;
                }

                setCollapsed((c) => !c);
              }}
              aria-expanded={expanded}
            >
              <span
                className={[
                  styles.chevron,
                  expanded ? styles.chevronOpen : '',
                ]
                  .filter(Boolean)
                  .join(' ')}
                aria-hidden="true"
              />
              {t('shell.system')}
            </button>
            {expanded ? (
              <div className={styles.sectionChildren}>
                {systemItems.map((item) => (
                  <NavLink
                    key={item.to}
                    to={item.to}
                    className={({ isActive }) =>
                      [styles.childItem, isActive ? styles.childItemActive : '']
                        .filter(Boolean)
                        .join(' ')
                    }
                  >
                    {item.label}
                  </NavLink>
                ))}
              </div>
            ) : null}
          </section>
        ) : null}
      </nav>
    </aside>
  );
}
