import { useEffect, useState } from 'react';

import { listAuditLogs } from '../../api/system';
import { getAuditResourceTypeLabel, getAuditResultLabel } from '../../locale/systemLabels';
import { useI18n } from '../../locale/useI18n';
import type { AuditLog } from '../../types/system';
import styles from './SystemAuditLogsPage.module.css';

export function SystemAuditLogsPage() {
  const { locale, t } = useI18n();
  const [logs, setLogs] = useState<AuditLog[]>([]);

  useEffect(() => {
    void listAuditLogs().then((result) => setLogs(result.items));
  }, []);

  return (
    <section className={styles.page}>
      <header>
        <p className="page-eyebrow">{t('system.audit.eyebrow')}</p>
        <h1 className={styles.title}>{t('system.audit.title')}</h1>
      </header>
      <div className={styles.list}>
        {logs.map((log) => (
          <article className={styles.item} key={log.id}>
            <strong>{log.action}</strong>
            <span>{getAuditResultLabel(log.result, locale)}</span>
            <p>{[getAuditResourceTypeLabel(log.resourceType, locale), log.resourceId].filter(Boolean).join(' ')}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
