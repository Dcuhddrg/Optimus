import { useEffect, useState } from 'react';

import { listAuditLogs } from '../../api/system';
import type { AuditLog } from '../../types/system';
import styles from './SystemAuditLogsPage.module.css';

export function SystemAuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);

  useEffect(() => {
    void listAuditLogs().then((result) => setLogs(result.items));
  }, []);

  return (
    <section className={styles.page}>
      <header>
        <p className="page-eyebrow">System / Audit</p>
        <h1 className={styles.title}>Audit logs</h1>
      </header>
      <div className={styles.list}>
        {logs.map((log) => (
          <article className={styles.item} key={log.id}>
            <strong>{log.action}</strong>
            <span>{log.result}</span>
            <p>{log.resourceType} {log.resourceId}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
