import { NavLink, Outlet } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import { useI18n } from '../../locale/useI18n';
import styles from './CloudLayout.module.css';

export function CloudLayout() {
  const { t } = useI18n();
  const { hasMenu } = useAuth();
  const showCloudAccounts = hasMenu('cloud.accounts.menu');
  const showCloudAssets = hasMenu('cloud.assets.menu');

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <p className="page-eyebrow">{t('cloud.module.eyebrow')}</p>
        <h1 className={styles.title}>{t('cloud.module.title')}</h1>
        <p className={styles.sub}>{t('cloud.module.sub')}</p>
      </header>

      <nav className={styles.tabs} aria-label={t('cloud.module.tabsLabel')}>
        {showCloudAccounts ? (
          <NavLink
            to="/cloud/accounts"
            className={({ isActive }) =>
              [styles.tab, isActive ? styles.tabActive : ''].filter(Boolean).join(' ')
            }
          >
            {t('shell.cloudAccounts')}
          </NavLink>
        ) : null}
        {showCloudAssets ? (
          <NavLink
            to="/cloud/assets"
            className={({ isActive }) =>
              [styles.tab, isActive ? styles.tabActive : ''].filter(Boolean).join(' ')
            }
          >
            {t('shell.cloudAssets')}
          </NavLink>
        ) : null}
      </nav>

      <div className={styles.body}>
        <Outlet />
      </div>
    </section>
  );
}
