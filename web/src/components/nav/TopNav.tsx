import { useEffect, useRef, useState } from 'react';
import { NavLink, useLocation } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import { useI18n } from '../../locale/useI18n';
import styles from './TopNav.module.css';

export function TopNav() {
  const { t } = useI18n();
  const { hasMenu } = useAuth();
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const showCloudAccounts = hasMenu('cloud.accounts.menu');
  const showCloudAssets = hasMenu('cloud.assets.menu');
  const hasCloud = showCloudAccounts || showCloudAssets;
  const isCloudActive = location.pathname.startsWith('/cloud');

  useEffect(() => {
    if (!open) {
      return;
    }
    function onMouseDown(event: MouseEvent) {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }
    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', onMouseDown);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onMouseDown);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  useEffect(() => {
    setOpen(false);
  }, [location.pathname]);

  if (!hasCloud) {
    return null;
  }

  return (
    <nav className={styles.topNav} aria-label={t('shell.moduleNav')}>
      <div className={styles.item} ref={containerRef}>
        <button
          type="button"
          className={[styles.tab, isCloudActive ? styles.tabActive : '']
            .filter(Boolean)
            .join(' ')}
          onClick={() => setOpen((value) => !value)}
          aria-haspopup="menu"
          aria-expanded={open}
        >
          <span className={styles.tabMark} aria-hidden="true" />
          <span className={styles.tabLabel}>{t('shell.cloud')}</span>
          <span
            className={[styles.chevron, open ? styles.chevronOpen : ''].filter(Boolean).join(' ')}
            aria-hidden="true"
          >
            ▾
          </span>
        </button>
        {open ? (
          <div className={styles.dropdown} role="menu">
            {showCloudAccounts ? (
              <NavLink
                to="/cloud/accounts"
                className={({ isActive }) =>
                  [styles.dropdownItem, isActive ? styles.dropdownItemActive : '']
                    .filter(Boolean)
                    .join(' ')
                }
                role="menuitem"
              >
                {t('shell.cloudAccounts')}
              </NavLink>
            ) : null}
            {showCloudAssets ? (
              <NavLink
                to="/cloud/assets"
                className={({ isActive }) =>
                  [styles.dropdownItem, isActive ? styles.dropdownItemActive : '']
                    .filter(Boolean)
                    .join(' ')
                }
                role="menuitem"
              >
                {t('shell.cloudAssets')}
              </NavLink>
            ) : null}
          </div>
        ) : null}
      </div>
    </nav>
  );
}
