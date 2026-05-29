// MyTicketsAside.tsx
import { Link } from 'react-router-dom';

import { useI18n } from '../../../locale/useI18n';
import { useTicketStore } from '../data/ticketStore';
import { TicketStatusBadge } from './TicketStatusBadge';
import styles from './MyTicketsAside.module.css';

const PREVIEW_COUNT = 5;

export function MyTicketsAside() {
  const { t } = useI18n();
  const { tickets } = useTicketStore();
  const preview = tickets.slice(0, PREVIEW_COUNT);

  return (
    <aside className={styles.aside}>
      <h3 className={styles.heading}>{t('servicePortal.ticket.myAside.title')}</h3>
      <p className={styles.description}>{t('servicePortal.ticket.myAside.description')}</p>

      {preview.length === 0 ? (
        <p className={styles.empty}>{t('servicePortal.ticket.myAside.empty')}</p>
      ) : (
        <div className={styles.list}>
          {preview.map((ticket) => (
            <Link key={ticket.id} to={`/service-portal/tickets/${ticket.id}`} className={styles.item}>
              <div className={styles.topline}>
                <strong className={styles.title}>{ticket.title}</strong>
                <TicketStatusBadge status={ticket.status} />
              </div>
              <p className={styles.meta}>
                {ticket.id} · {new Date(ticket.createdAt).toLocaleString()}
              </p>
            </Link>
          ))}
        </div>
      )}

      <Link to="/service-portal/tickets" className={styles.viewAll}>
        {t('servicePortal.ticket.myAside.viewAll')}
      </Link>
    </aside>
  );
}
