import { Outlet } from 'react-router-dom';

import { useAuth } from '../auth/AuthContext';
import { SideNav } from '../components/nav/SideNav';
import styles from './AppShell.module.css';

export function AppShell() {
  const { current, logout } = useAuth();

  return (
    <div className={styles.shell}>
      <SideNav />
      <div className={styles.content}>
        <header className={styles.topbar}>
          <div>
            <p className={styles.eyebrow}>Optimus Ops</p>
            <h1 className={styles.title}>Operations workspace</h1>
          </div>
          <div className={styles.session}>
            <span className={styles.badge}>{current?.user.displayName || current?.user.username}</span>
            <button className={styles.logout} type="button" onClick={() => void logout()}>
              Sign out
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
