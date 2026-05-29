// TicketDetailPage.tsx
import { Link, useParams } from 'react-router-dom';

import { useI18n } from '../../locale/useI18n';
import type { MessageKey } from '../../locale/messages';
import { PageHeader } from '../../components/page/PageHeader';
import { TicketStatusBadge } from './components/TicketStatusBadge';
import { TicketTimeline } from './components/TicketTimeline';
import { getCatalogById } from './data/serviceCatalog';
import { useTicketStore } from './data/ticketStore';
import styles from './TicketDetailPage.module.css';

export function TicketDetailPage() {
  const { id = '' } = useParams<{ id: string }>();
  const { t } = useI18n();
  const { getTicket } = useTicketStore();
  const ticket = getTicket(id);

  if (!ticket) {
    return (
      <div className={styles.page}>
        <div className={styles.breadcrumb}>
          <Link to="/service-portal">{t('servicePortal.breadcrumbHome')}</Link>
          <span>/</span>
          <Link to="/service-portal/tickets">{t('servicePortal.breadcrumbTickets')}</Link>
        </div>
        <div className={styles.emptyState}>
          <h3>{t('servicePortal.notFoundTitle')}</h3>
          <p>{t('servicePortal.notFoundDescription')}</p>
          <Link to="/service-portal/tickets">{t('servicePortal.backToTickets')}</Link>
        </div>
      </div>
    );
  }

  const catalog = ticket.serviceId ? getCatalogById(ticket.serviceId) : null;
  const serviceLabel = catalog
    ? t(`${catalog.i18nKey}.title` as MessageKey)
    : t('servicePortal.ticket.detail.genericServiceLabel');

  return (
    <div className={styles.page}>
      <div className={styles.breadcrumb}>
        <Link to="/service-portal">{t('servicePortal.breadcrumbHome')}</Link>
        <span>/</span>
        <Link to="/service-portal/tickets">{t('servicePortal.breadcrumbTickets')}</Link>
        <span>/</span>
        <span>{ticket.title}</span>
      </div>

      <PageHeader
        eyebrow={t('servicePortal.ticket.detail.eyebrow')}
        title={ticket.title}
        description={`${ticket.id} · ${new Date(ticket.createdAt).toLocaleString()}`}
        actions={<TicketStatusBadge status={ticket.status} />}
      />

      <section className={styles.summary}>
        <div className={styles.field}>
          <span className={styles.fieldLabel}>{t('servicePortal.ticket.detail.serviceLabel')}</span>
          <span className={styles.fieldValue}>{serviceLabel}</span>
        </div>
        <div className={styles.field}>
          <span className={styles.fieldLabel}>{t('servicePortal.ticket.detail.applicantLabel')}</span>
          <span className={styles.fieldValue}>
            {ticket.applicant.name || ticket.applicant.email}
          </span>
        </div>
      </section>

      <section className={styles.section}>
        <h3>{t('servicePortal.ticket.detail.formValuesLabel')}</h3>
        <dl className={styles.formValues}>
          {Object.entries(ticket.formValues).map(([key, value]) => (
            <div key={key}>
              <dt>{key}</dt>
              <dd>{Array.isArray(value) ? value.join(', ') : String(value ?? '')}</dd>
            </div>
          ))}
        </dl>
      </section>

      <section className={styles.section}>
        <h3>{t('servicePortal.ticket.detail.timelineLabel')}</h3>
        <TicketTimeline ticket={ticket} />
      </section>
    </div>
  );
}
