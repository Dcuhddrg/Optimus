import { useEffect, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import { useI18n } from '../../locale/useI18n';
import styles from './TopNav.module.css';

type Board = {
  code: string;
  name: string;
  sub: string;
  to: string;
  glyph: string;
  matches: (path: string) => boolean;
};

export function TopNav() {
  const { t } = useI18n();
  const { hasMenu } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const hasCloud = hasMenu('cloud.accounts.menu') || hasMenu('cloud.assets.menu');

  const boards: Board[] = [
    {
      code: 'workspace',
      name: t('shell.boardWorkspace'),
      sub: t('shell.boardWorkspaceSub'),
      to: '/',
      glyph: 'W',
      matches: (path) => !path.startsWith('/cloud'),
    },
    ...(hasCloud
      ? [
          {
            code: 'cloud',
            name: t('shell.cloud'),
            sub: t('shell.boardCloudSub'),
            to: '/cloud',
            glyph: '☁',
            matches: (path) => path.startsWith('/cloud'),
          } satisfies Board,
        ]
      : []),
  ];

  const activeBoard = boards.find((b) => b.matches(location.pathname)) ?? boards[0];

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

  function handleBoardClick(board: Board) {
    setOpen(false);
    if (board.code !== activeBoard?.code) {
      navigate(board.to);
    }
  }

  return (
    <nav className={styles.topNav} aria-label={t('shell.platformNav')}>
      <div className={styles.item} ref={containerRef}>
        <button
          type="button"
          className={styles.tab}
          onClick={() => setOpen((value) => !value)}
          aria-haspopup="dialog"
          aria-expanded={open}
        >
          <span className={styles.tabMark} aria-hidden="true" />
          <span className={styles.tabLabel}>{t('shell.platformNav')}</span>
          <span className={styles.tabBoard}>{activeBoard?.name}</span>
          <span
            className={[styles.chevron, open ? styles.chevronOpen : ''].filter(Boolean).join(' ')}
            aria-hidden="true"
          >
            ▾
          </span>
        </button>
        {open ? (
          <div className={styles.panel} role="dialog" aria-label={t('shell.platformNav')}>
            <p className={styles.panelHint}>{t('shell.platformNav')}</p>
            <div className={styles.boards}>
              {boards.map((board) => {
                const active = board.code === activeBoard?.code;
                return (
                  <button
                    key={board.code}
                    type="button"
                    className={[styles.boardCard, active ? styles.boardCardActive : '']
                      .filter(Boolean)
                      .join(' ')}
                    onClick={() => handleBoardClick(board)}
                  >
                    <span className={styles.boardGlyph} aria-hidden="true">
                      {board.glyph}
                    </span>
                    <span className={styles.boardText}>
                      <span className={styles.boardName}>{board.name}</span>
                      <span className={styles.boardSub}>{board.sub}</span>
                    </span>
                    {active ? <span className={styles.boardActiveDot} aria-hidden="true" /> : null}
                  </button>
                );
              })}
            </div>
          </div>
        ) : null}
      </div>
    </nav>
  );
}
