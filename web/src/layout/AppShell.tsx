import { Outlet } from 'react-router-dom';

import { useAuth } from '../auth/AuthContext';
import { LanguageSwitcher } from '../components/locale/LanguageSwitcher';
import { SideNav } from '../components/nav/SideNav';
import { ProjectSwitcher } from '../components/project/ProjectSwitcher';
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
          <div className={styles.topbarActions}>
            <ProjectSwitcher />
            <LanguageSwitcher />
            <div className={styles.session}>
              <span className={styles.badge}>{current?.user.displayName || current?.user.username}</span>
              <button className={styles.logout} type="button" onClick={() => void logout()}>
                {t('shell.signOut')}
              </button>
            </div>
          </div>
        </header>
        <main className={styles.main}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
