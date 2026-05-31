// RequestDrawer.tsx
import { useEffect, type ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { X } from '@phosphor-icons/react';

import styles from './RequestDrawer.module.css';

interface Props {
  open: boolean;
  onClose: () => void;
  eyebrow?: string;
  title: string;
  description?: string;
  children: ReactNode;
}

export function RequestDrawer({ open, onClose, eyebrow, title, description, children }: Props) {
  useEffect(() => {
    if (!open) return;
    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, onClose]);

  if (typeof document === 'undefined') return null;

  return createPortal(
    <>
      <div className={styles.overlay} data-open={open} onClick={onClose} aria-hidden="true" />
      <aside
        className={styles.drawer}
        data-open={open}
        role="dialog"
        aria-modal="true"
        aria-label={title}
      >
        <header className={styles.header}>
          <div>
            {eyebrow ? <p className={styles.eyebrow}>{eyebrow}</p> : null}
            <h2 className={styles.title}>{title}</h2>
            {description ? <p className={styles.description}>{description}</p> : null}
          </div>
          <button type="button" className={styles.close} onClick={onClose} aria-label="Close">
            <X size={18} weight="bold" />
          </button>
        </header>
        <div className={styles.body}>{children}</div>
      </aside>
    </>,
    document.body,
  );
}
