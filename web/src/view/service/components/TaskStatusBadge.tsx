// TaskStatusBadge.tsx
import { AnimatePresence, motion } from 'framer-motion';

import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { TaskStatus } from '../data/taskTypes';
import styles from './TaskStatusBadge.module.css';

interface Props {
  status: TaskStatus;
}

const STATUS_I18N: Record<TaskStatus, MessageKey> = {
  pending: 'servicePortal.task.status.pending',
  in_progress: 'servicePortal.task.status.in_progress',
  done: 'servicePortal.task.status.done',
};

export function TaskStatusBadge({ status }: Props) {
  const { t } = useI18n();
  return (
    <AnimatePresence mode="wait" initial={false}>
      <motion.span
        key={status}
        className={`${styles.badge} ${styles[status]}`}
        initial={{ scale: 0.85, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.85, opacity: 0 }}
        transition={{ type: 'spring', stiffness: 380, damping: 24 }}
      >
        {t(STATUS_I18N[status])}
      </motion.span>
    </AnimatePresence>
  );
}
