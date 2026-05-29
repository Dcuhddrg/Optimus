// TicketTimeline.tsx
import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { Ticket } from '../ticketTypes';
import styles from './TicketTimeline.module.css';

interface Props {
  ticket: Ticket;
}

export function TicketTimeline({ ticket }: Props) {
  const { t } = useI18n();
  return (
    <ol className={styles.timeline}>
      {ticket.timeline.map((event, index) => {
        const statusLabel = t(`servicePortal.ticket.status.${event.status}` as MessageKey);
        const note = event.i18nNoteKey
          ? t(event.i18nNoteKey as MessageKey)
              .replace('{applicant}', ticket.applicant.name || ticket.applicant.email)
              .replace('{status}', statusLabel)
          : statusLabel;
        return (
          <li key={`${event.at}-${index}`} className={styles.item}>
            <span className={styles.timestamp}>{new Date(event.at).toLocaleString()}</span>
            <p className={styles.text}>{note}</p>
          </li>
        );
      })}
    </ol>
  );
}
