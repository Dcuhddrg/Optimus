// TicketTimeline.tsx
import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { Ticket } from '../ticketTypes';
import styles from './TicketTimeline.module.css';
import { formatDateTime } from '../../../utils/format';

interface Props {
  ticket: Ticket;
}

export function TicketTimeline({ ticket }: Props) {
  const { t, locale } = useI18n();
  return (
    <div className={styles.wrap}>
      <svg
        className={styles.spine}
        preserveAspectRatio="none"
        viewBox="0 0 2 100"
        aria-hidden="true"
      >
        <path d="M1 0 V100" pathLength={100} stroke="rgba(86, 70, 55, 0.22)" strokeWidth={2} />
      </svg>
      <ol className={styles.timeline}>
      {ticket.timeline.map((event, index) => {
        const statusLabel = t(`servicePortal.ticket.status.${event.status}` as MessageKey);
        const note = event.i18nNoteKey
          ? t(event.i18nNoteKey as MessageKey)
              .replace('{applicant}', ticket.applicant.name || ticket.applicant.email)
              .replace('{status}', statusLabel)
              .replace('{actor}', event.actor ?? '')
              .replace('{role}', event.roleI18nKey ? t(event.roleI18nKey as MessageKey) : '')
          : statusLabel;
        return (
          <li key={`${event.at}-${index}`} className={styles.item}>
            <span className={styles.timestamp}>{formatDateTime(event.at, locale)}</span>
            <p className={styles.text}>{note}</p>
          </li>
        );
      })}
      </ol>
    </div>
  );
}
