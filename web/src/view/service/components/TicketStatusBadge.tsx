// TicketStatusBadge.tsx

import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { TicketStatus } from '../ticketTypes';
import styles from './TicketStatusBadge.module.css';

interface Props {
  status: TicketStatus;
}

const STATUS_I18N: Record<TicketStatus, MessageKey> = {
  submitted: 'servicePortal.ticket.status.submitted',
  approving: 'servicePortal.ticket.status.approving',
  processing: 'servicePortal.ticket.status.processing',
  done: 'servicePortal.ticket.status.done',
  rejected: 'servicePortal.ticket.status.rejected',
};

export function TicketStatusBadge({ status }: Props) {
  const { t } = useI18n();
  return <span className={`${styles.badge} ${styles[status]}`}>{t(STATUS_I18N[status])}</span>;
}
