import type { ReactNode } from "react";

import styles from "./EmptyState.module.css";

interface EmptyStateProps {
  eyebrow?: string;
  title: string;
  description: string;
  actions?: ReactNode;
  aside?: ReactNode;
}

export function EmptyState({
  eyebrow,
  title,
  description,
  actions,
  aside,
}: EmptyStateProps) {
  return (
    <section className={styles.card}>
      <div className={styles.content}>
        {eyebrow ? <p className={styles.eyebrow}>{eyebrow}</p> : null}
        <h2 className={styles.title}>{title}</h2>
        <p className={styles.description}>{description}</p>
        {actions ? <div className={styles.actions}>{actions}</div> : null}
      </div>
      <div className={styles.visual} aria-hidden="true">
        <div className={styles.orbit}>
          <span className={styles.dotPrimary} />
          <span className={styles.dotSecondary} />
          <span className={styles.ring} />
        </div>
        {aside ? <div className={styles.aside}>{aside}</div> : null}
      </div>
    </section>
  );
}
