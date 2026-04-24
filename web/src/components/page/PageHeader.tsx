import type { ReactNode } from "react";

import styles from "./PageHeader.module.css";

interface PageHeaderProps {
  eyebrow?: string;
  title: string;
  description?: string;
  actions?: ReactNode;
}

export function PageHeader({
  eyebrow,
  title,
  description,
  actions,
}: PageHeaderProps) {
  return (
    <header className={styles.header}>
      <div className={styles.copy}>
        {eyebrow ? <p className={styles.eyebrow}>{eyebrow}</p> : null}
        <div className={styles.titleRow}>
          <h1 className={styles.title}>{title}</h1>
          {actions ? <div className={styles.actions}>{actions}</div> : null}
        </div>
        {description ? <p className={styles.description}>{description}</p> : null}
      </div>
    </header>
  );
}
