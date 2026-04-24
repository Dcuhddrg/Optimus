import { Outlet } from 'react-router-dom';

import { SideNav } from '../components/nav/SideNav';
import styles from './AppShell.module.css';

export function AppShell() {
  return (
    <div className={styles.shell}>
      <SideNav />
      <div className={styles.content}>
        <header className={styles.topbar}>
          <div>
            <p className={styles.eyebrow}>Optimus Ops</p>
            <h1 className={styles.title}>Operations workspace</h1>
          </div>
          <div className={styles.badge}>Preview</div>
        </header>
        <main className={styles.main}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
