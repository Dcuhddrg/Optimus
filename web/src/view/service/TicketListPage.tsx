// TicketListPage.tsx
import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { useI18n } from '../../locale/useI18n';
import type { MessageKey } from '../../locale/messages';
import { PageHeader } from '../../components/page/PageHeader';
import { TicketStatusBadge } from './components/TicketStatusBadge';
import { useTicketStore } from './data/ticketStore';
import type { TicketStatus } from './ticketTypes';
import styles from './TicketListPage.module.css';

const ALL_STATUSES: TicketStatus[] = ['submitted', 'approving', 'processing', 'done', 'rejected'];

export function TicketListPage() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const { tickets } = useTicketStore();
  const [search, setSearch] = useState('');
  const [activeStatuses, setActiveStatuses] = useState<Set<TicketStatus>>(new Set());

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase();
    return tickets.filter((ticket) => {
      if (activeStatuses.size > 0 && !activeStatuses.has(ticket.status)) return false;
      if (query) {
        const hay = `${ticket.id} ${ticket.title}`.toLowerCase();
        if (!hay.includes(query)) return false;
      }
      return true;
    });
  }, [tickets, search, activeStatuses]);

  function toggleStatus(status: TicketStatus) {
    setActiveStatuses((prev) => {
      const next = new Set(prev);
      if (next.has(status)) next.delete(status);
      else next.add(status);
      return next;
    });
  }

  return (
    <div className={styles.page}>
      <PageHeader
        eyebrow={t('servicePortal.ticket.list.eyebrow')}
        title={t('servicePortal.ticket.list.title')}
        description={t('servicePortal.ticket.list.description')}
      />

      <div className={styles.controls}>
        <input
          className={styles.search}
          placeholder={t('servicePortal.ticket.list.searchPlaceholder')}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className={styles.chips} role="group" aria-label={t('servicePortal.ticket.list.filterLabel')}>
          {ALL_STATUSES.map((status) => {
            const active = activeStatuses.has(status);
            return (
              <button
                key={status}
                type="button"
                className={active ? `${styles.chip} ${styles.chipActive}` : styles.chip}
                onClick={() => toggleStatus(status)}
              >
                {t(`servicePortal.ticket.status.${status}` as MessageKey)}
              </button>
            );
          })}
        </div>
      </div>

      <table className={styles.table}>
        <thead>
          <tr>
            <th>{t('servicePortal.ticket.list.columnId')}</th>
            <th>{t('servicePortal.ticket.list.columnTitle')}</th>
            <th>{t('servicePortal.ticket.list.columnStatus')}</th>
            <th>{t('servicePortal.ticket.list.columnCreated')}</th>
          </tr>
        </thead>
        <tbody>
          {filtered.length === 0 ? (
            <tr>
              <td colSpan={4} className={styles.empty}>
                {t('servicePortal.ticket.list.empty')}
              </td>
            </tr>
          ) : (
            filtered.map((ticket) => (
              <tr
                key={ticket.id}
                className={styles.row}
                onClick={() => navigate(`/service-portal/tickets/${ticket.id}`)}
              >
                <td>{ticket.id}</td>
                <td>{ticket.title}</td>
                <td>
                  <TicketStatusBadge status={ticket.status} />
                </td>
                <td>{new Date(ticket.createdAt).toLocaleString()}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
