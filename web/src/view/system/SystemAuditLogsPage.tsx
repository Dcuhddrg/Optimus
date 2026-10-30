import { useEffect, useRef, useState, type FormEvent } from 'react';
import { ApiError } from '../../api/client';
import { exportAuditLogs, listAuditLogs } from '../../api/system';
import { AUDIT_ACTION_CODES, getAuditActionLabel, getAuditResourceTypeLabel, getAuditResultLabel } from '../../locale/systemLabels';
import { useI18n } from '../../locale/useI18n';
import type { AuditLog, AuditLogFilter, AuditLogPage } from '../../types/system';
import styles from './SystemAuditLogsPage.module.css';

const emptyFilters = { actorUserId: '', action: '', resourceType: '', result: '', startTime: '', endTime: '', keyword: '' };
type FilterFields = typeof emptyFilters;
type Query = AuditLogFilter & { page: number; pageSize: number };
function detailText(value: string) {
  try { return JSON.stringify(JSON.parse(value), null, 2); } catch { return value; }
}

export function SystemAuditLogsPage() {
  const { locale, t } = useI18n();
  const [draft, setDraft] = useState<FilterFields>(emptyFilters);
  const [query, setQuery] = useState<Query>({ page: 1, pageSize: 20 });
  const [data, setData] = useState<AuditLogPage | null>(null);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [validation, setValidation] = useState<'invalidRange' | 'invalidActor' | 'invalidTime' | null>(null);
  const [refresh, setRefresh] = useState(0);
  const [selected, setSelected] = useState<AuditLog | null>(null);
  const [exporting, setExporting] = useState(false);
  const [exportError, setExportError] = useState<'exportFailed' | 'exportForbidden' | null>(null);
  const download = useRef<AbortController | null>(null);
  const totalPages = Math.max(1, Math.ceil((data?.total ?? 0) / query.pageSize));

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true); setFailed(false); setData(null); setSelected(null);
    download.current?.abort(); setExporting(false); setExportError(null);
    void listAuditLogs(query, controller.signal).then((result) => {
      if (controller.signal.aborted) return;
      if (!Array.isArray(result?.list) || !Number.isFinite(result.total) || result.total < 0) {
        throw new ApiError('Invalid audit page', 200, 'protocol');
      }
      const lastPage = Math.max(1, Math.ceil(result.total / query.pageSize));
      if (query.page > lastPage) {
        setQuery(current => ({ ...current, page: lastPage }));
        return;
      }
      setData(result);
    }).catch(() => {
      if (!controller.signal.aborted) setFailed(true);
    }).finally(() => {
      if (!controller.signal.aborted) setLoading(false);
    });
    return () => controller.abort();
  }, [query, refresh]);
  useEffect(() => () => download.current?.abort(), []);

  function apply(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const actor = draft.actorUserId ? Number(draft.actorUserId) : undefined;
    if (actor !== undefined && (!Number.isSafeInteger(actor) || actor <= 0)) { setValidation('invalidActor'); return; }
    const start = draft.startTime ? new Date(draft.startTime) : undefined;
    const end = draft.endTime ? new Date(draft.endTime) : undefined;
    if ((start && !Number.isFinite(start.getTime())) || (end && !Number.isFinite(end.getTime()))) { setValidation('invalidTime'); return; }
    if (start && end && start > end) { setValidation('invalidRange'); return; }
    setValidation(null);
    setQuery({ actorUserId: actor, action: draft.action, resourceType: draft.resourceType, result: draft.result,
      startTime: start?.toISOString(), endTime: end?.toISOString(), keyword: draft.keyword.trim(), page: 1, pageSize: query.pageSize });
  }

  async function exportCSV() {
    download.current?.abort();
    const controller = new AbortController(); download.current = controller;
    setExporting(true); setExportError(null);
    try {
      const blob = await exportAuditLogs(query, controller.signal);
      if (controller.signal.aborted) return;
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      try {
        link.href = url; link.download = 'audit-logs.csv'; document.body.appendChild(link); link.click();
      } finally {
        link.remove(); URL.revokeObjectURL(url);
      }
    } catch (error) {
      if (!controller.signal.aborted) setExportError(error instanceof ApiError && error.status === 403 ? 'exportForbidden' : 'exportFailed');
    } finally {
      if (!controller.signal.aborted) setExporting(false);
    }
  }

  const actorName = (log: AuditLog) => log.actorUsername || (log.actorUserId ? `#${log.actorUserId}` : t('system.audit.unknownActor'));
  const timestamp = (value: string) => {
    const date = new Date(value);
    return Number.isFinite(date.getTime()) ? date.toLocaleString(locale, { hour12: false }) : value || t('system.audit.notRecorded');
  };
  const input = (key: keyof FilterFields, type = 'text') => <input type={type} value={draft[key]}
    min={key === 'actorUserId' ? '1' : undefined} step={key === 'actorUserId' ? '1' : undefined}
    onChange={event => setDraft(value => ({ ...value, [key]: event.target.value }))} />;
  const select = (key: 'action' | 'result' | 'resourceType', values: readonly string[], label: (value: string) => string) =>
    <select value={draft[key]} onChange={event => setDraft(value => ({ ...value, [key]: event.target.value }))}>
      <option value="">{t('system.audit.all')}</option>
      {values.map(value => <option key={value} value={value}>{label(value)}</option>)}
    </select>;

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <div><p className="page-eyebrow">{t('system.audit.eyebrow')}</p><h1 className={styles.title}>{t('system.audit.title')}</h1>
          <p className={styles.muted}>{t('system.audit.sub')}</p></div>
        <button type="button" disabled={loading || failed || exporting} onClick={() => void exportCSV()}>
          {t(exporting ? 'system.audit.exporting' : 'system.audit.exportCSV')}
        </button>
      </header>
      <p className={styles.muted}>{t('system.audit.exportHint')}</p>
      {exportError && <p className={styles.error} role="alert">{t(`system.audit.${exportError}`)}</p>}
      <form onSubmit={apply} className={styles.filters}>
        <label>{t('system.audit.startTime')}{input('startTime', 'datetime-local')}</label>
        <label>{t('system.audit.endTime')}{input('endTime', 'datetime-local')}</label>
        <label>{t('system.audit.actorId')}{input('actorUserId', 'number')}</label>
        <label>{t('system.audit.action')}{select('action', AUDIT_ACTION_CODES, value => getAuditActionLabel(value, locale))}</label>
        <label>{t('system.audit.resourceType')}{select('resourceType', ['user', 'role', 'ldap'], value => getAuditResourceTypeLabel(value, locale))}</label>
        <label>{t('system.audit.result')}{select('result', ['success', 'failure'], value => getAuditResultLabel(value, locale))}</label>
        <label>{t('system.audit.keyword')}{input('keyword')}</label>
        <div className={styles.actions}>
          <button type="submit">{t('system.audit.search')}</button>
          <button type="button" onClick={() => { setDraft(emptyFilters); setValidation(null); setQuery({ page: 1, pageSize: query.pageSize }); }}>{t('system.audit.reset')}</button>
        </div>
        {validation && <p className={styles.error} role="alert">{t(`system.audit.${validation}`)}</p>}
      </form>
      {loading ? <p role="status">{t('system.audit.loading')}</p> : failed ?
        <div className={styles.error} role="alert"><p>{t('system.audit.loadFailed')}</p><button type="button" onClick={() => setRefresh(value => value + 1)}>{t('system.audit.retry')}</button></div> :
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <caption className={styles.srOnly}>{t('system.audit.title')}</caption>
            <thead><tr>{(['time', 'actor', 'action', 'resource', 'result', 'ip', 'details'] as const).map(key => <th scope="col" key={key}>{t(`system.audit.${key}`)}</th>)}</tr></thead>
            <tbody>{data?.list.map(log => <tr key={log.id}>
              <td><time dateTime={log.createdAt}>{timestamp(log.createdAt)}</time></td>
              <td>{actorName(log)}</td><td>{getAuditActionLabel(log.action, locale)}</td>
              <td>{[getAuditResourceTypeLabel(log.resourceType, locale), log.resourceId].filter(Boolean).join(' ') || t('system.audit.notRecorded')}</td>
              <td><span className={log.result === 'failure' ? styles.failure : log.result === 'success' ? styles.success : undefined}>{getAuditResultLabel(log.result, locale)}</span></td>
              <td>{log.ip || t('system.audit.notRecorded')}</td>
              <td><button type="button" aria-expanded={selected?.id === log.id} aria-controls="audit-log-detail" onClick={() => setSelected(selected?.id === log.id ? null : log)}>{t('system.audit.details')}</button></td>
            </tr>)}</tbody>
          </table>
          {data?.list.length === 0 && <p className={styles.empty}>{t('system.audit.empty')}</p>}
        </div>}
      <div className={styles.pagination}>
        <span>{t('system.audit.total')}: {data?.total ?? '—'}</span>
        <div className={styles.actions}>
          <button type="button" disabled={loading || failed || query.page <= 1} onClick={() => setQuery(value => ({ ...value, page: value.page - 1 }))}>{t('system.audit.prev')}</button>
          <span aria-live="polite">{t('system.audit.page')} {query.page} / {data ? totalPages : '—'}</span>
          <button type="button" disabled={loading || failed || query.page >= totalPages} onClick={() => setQuery(value => ({ ...value, page: value.page + 1 }))}>{t('system.audit.next')}</button>
          <label>{t('system.audit.rowsPerPage')}<select value={query.pageSize} onChange={event => setQuery(value => ({ ...value, page: 1, pageSize: Number(event.target.value) }))}>
            {[20, 50, 100].map(size => <option key={size} value={size}>{size}</option>)}
          </select></label>
        </div>
      </div>
      {selected && <section className={styles.detail} id="audit-log-detail" aria-labelledby="audit-detail-title">
        <header className={styles.header}><h2 id="audit-detail-title">{t('system.audit.detailTitle')} #{selected.id}</h2>
          <button type="button" onClick={() => setSelected(null)}>{t('system.audit.closeDetails')}</button></header>
        <dl>{[
          [t('system.audit.time'), timestamp(selected.createdAt)], [t('system.audit.actor'), actorName(selected)],
          [t('system.audit.actorId'), selected.actorUserId || t('system.audit.notRecorded')],
          [t('system.audit.rawAction'), selected.action], [t('system.audit.resourceType'), selected.resourceType],
          [t('system.audit.resource'), selected.resourceId], [t('system.audit.result'), getAuditResultLabel(selected.result, locale)],
          [t('system.audit.ip'), selected.ip], [t('system.audit.userAgent'), selected.userAgent],
        ].map(([key, value]) => <div key={key}><dt>{key}</dt><dd>{value || t('system.audit.notRecorded')}</dd></div>)}</dl>
        <h3>{t('system.audit.detail')}</h3><pre>{selected.detail ? detailText(selected.detail) : t('system.audit.notRecorded')}</pre>
      </section>}
    </section>
  );
}
