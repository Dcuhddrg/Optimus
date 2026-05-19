import { NavLink } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import { useI18n } from '../../locale/useI18n';
import styles from './SideNav.module.css';

type NavItem = {
  to: string;
  label: string;
};

function isNavItem(item: NavItem | null): item is NavItem {
  return item !== null;
}

export function SideNav() {
  const { hasMenu } = useAuth();
  const { t } = useI18n();
  const showUsers = hasMenu('system.users.menu');
  const showRoles = hasMenu('system.roles.menu');
  const showLDAP = hasMenu('system.ldap.menu');
  const showAudit = hasMenu('system.audit.menu');
  const showProjects = hasMenu('project.menu');

  const platformItems = [
    showProjects ? { to: '/projects', label: t('shell.projects') } : null,
    showUsers ? { to: '/system/users', label: t('shell.users') } : null,
    showRoles ? { to: '/system/roles', label: t('shell.roles') } : null,
    showLDAP ? { to: '/system/ldap', label: t('shell.ldap') } : null,
    showAudit ? { to: '/system/audit-logs', label: t('shell.auditLogs') } : null,
  ].filter(isNavItem);
  const hasPlatformItems = platformItems.length > 0;

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
          {t('shell.overview')}
        </NavLink>

        <section className={styles.group}>
          <p className={styles.groupLabel}>{t('shell.opsCenter')}</p>
          <span className={styles.navEntryDisabled} title={t('shell.comingSoon')}>
            {t('shell.servicePortal')}
          </span>
          <span className={styles.navEntryDisabled} title={t('shell.comingSoon')}>
            {t('shell.tickets')}
          </span>
          <span className={styles.navEntryDisabled} title={t('shell.comingSoon')}>
            {t('shell.tasks')}
          </span>
          <span className={styles.navEntryDisabled} title={t('shell.comingSoon')}>
            {t('shell.myDuty')}
          </span>
          <span className={styles.navEntryDisabled} title={t('shell.comingSoon')}>
            {t('shell.bigScreen')}
          </span>
        </section>

        {hasPlatformItems ? (
          <section className={styles.group}>
            <p className={styles.groupLabel}>{t('shell.platformSettings')}</p>
            {platformItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  [styles.navEntry, isActive ? styles.navEntryActive : '']
                    .filter(Boolean)
                    .join(' ')
                }
              >
                {item.label}
              </NavLink>
            ))}
          </section>
        ) : null}
      </nav>
    </aside>
  );
}
