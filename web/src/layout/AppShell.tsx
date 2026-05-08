import { Outlet } from 'react-router-dom';

import { useAuth } from '../auth/AuthContext';
import { LanguageSwitcher } from '../components/locale/LanguageSwitcher';
import { SideNav } from '../components/nav/SideNav';
import { useI18n } from '../locale/useI18n';
import styles from './AppShell.module.css';

export function AppShell() {
  const { current, logout } = useAuth();
  const { t } = useI18n();

  return (
    <div className={styles.shell}>
      <SideNav />
      <div className={styles.content}>
        <header className={styles.topbar}>
          <div>
            <p className={styles.eyebrow}>{t('shell.topbarEyebrow')}</p>
            <h1 className={styles.title}>{t('shell.topbarTitle')}</h1>
          </div>
          <div className={styles.session}>
            <LanguageSwitcher />
            <span className={styles.badge}>{current?.user.displayName || current?.user.username}</span>
            <button className={styles.logout} type="button" onClick={() => void logout()}>
              {t('shell.signOut')}
            </button>
          </div>
        </header>
        <main className={styles.main}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
