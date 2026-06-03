// TaskListPage.tsx
import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import { useI18n } from '../../locale/useI18n';
import type { MessageKey } from '../../locale/messages';
import { PageHeader } from '../../components/page/PageHeader';
import { TaskStatusBadge } from './components/TaskStatusBadge';
import { useTaskStore } from './data/taskStore';
import type { TaskStatus } from './data/taskTypes';
import styles from './TaskListPage.module.css';

const ALL_STATUSES: TaskStatus[] = ['pending', 'in_progress', 'done'];

export function TaskListPage() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const { current } = useAuth();
  const { tasks, claim, complete, release } = useTaskStore();
  const [search, setSearch] = useState('');
  const [activeStatuses, setActiveStatuses] = useState<Set<TaskStatus>>(new Set());
  const [mineOnly, setMineOnly] = useState(false);

  const me = {
    name: current?.user.displayName ?? current?.user.username ?? '',
    email: current?.user.email ?? '',
  };

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase();
    return tasks.filter((task) => {
      if (activeStatuses.size > 0 && !activeStatuses.has(task.status)) return false;
      if (mineOnly && task.assignee?.email !== me.email) return false;
      if (query) {
        const hay = `${task.id} ${task.title} ${task.ticketId}`.toLowerCase();
        if (!hay.includes(query)) return false;
      }
      return true;
    });
  }, [tasks, search, activeStatuses, mineOnly, me.email]);

  function toggleStatus(status: TaskStatus) {
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
        eyebrow={t('servicePortal.task.list.eyebrow')}
        title={t('servicePortal.task.list.title')}
        description={t('servicePortal.task.list.description')}
      />

      <div className={styles.controls}>
        <input
          className={styles.search}
          placeholder={t('servicePortal.task.list.searchPlaceholder')}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className={styles.chips} role="group" aria-label={t('servicePortal.task.list.filterLabel')}>
          {ALL_STATUSES.map((status) => {
            const active = activeStatuses.has(status);
            return (
              <button
                key={status}
                type="button"
                className={active ? `${styles.chip} ${styles.chipActive}` : styles.chip}
                onClick={() => toggleStatus(status)}
              >
                {t(`servicePortal.task.status.${status}` as MessageKey)}
              </button>
            );
          })}
          <button
            type="button"
            className={mineOnly ? `${styles.chip} ${styles.chipActive}` : styles.chip}
            onClick={() => setMineOnly((v) => !v)}
          >
            {t('servicePortal.task.list.mineOnly')}
          </button>
        </div>
      </div>

      <table className={styles.table}>
        <thead>
          <tr>
            <th>{t('servicePortal.task.list.columnId')}</th>
            <th>{t('servicePortal.task.list.columnTitle')}</th>
            <th>{t('servicePortal.task.list.columnService')}</th>
            <th>{t('servicePortal.task.list.columnStatus')}</th>
            <th>{t('servicePortal.task.list.columnAssignee')}</th>
            <th>{t('servicePortal.task.list.columnActions')}</th>
          </tr>
        </thead>
        <tbody>
          {filtered.length === 0 ? (
            <tr>
              <td colSpan={6} className={styles.empty}>
                {t('servicePortal.task.list.empty')}
              </td>
            </tr>
          ) : (
            filtered.map((task) => {
              const isMine = task.assignee?.email === me.email;
              return (
                <tr key={task.id}>
                  <td className={styles.mono}>{task.id}</td>
                  <td>
                    <div className={styles.titleCell}>
                      <span>{task.title}</span>
                      <span className={styles.mono}>{task.ticketId}</span>
                    </div>
                  </td>
                  <td>{task.serviceId ?? '—'}</td>
                  <td>
                    <TaskStatusBadge status={task.status} />
                  </td>
                  <td>{task.assignee?.name ?? t('servicePortal.task.list.unassigned')}</td>
                  <td>
                    <div className={styles.actions}>
                      {task.status === 'pending' ? (
                        <button type="button" className={styles.action} onClick={() => claim(task.id, me)}>
                          {t('servicePortal.task.list.claim')}
                        </button>
                      ) : null}
                      {task.status === 'in_progress' && isMine ? (
                        <>
                          <button type="button" className={styles.action} onClick={() => complete(task.id)}>
                            {t('servicePortal.task.list.complete')}
                          </button>
                          <button
                            type="button"
                            className={styles.actionGhost}
                            onClick={() => release(task.id)}
                          >
                            {t('servicePortal.task.list.release')}
                          </button>
                        </>
                      ) : null}
                      <button
                        type="button"
                        className={styles.actionGhost}
                        onClick={() => navigate(`/service-portal/tickets/${task.ticketId}`)}
                      >
                        {t('servicePortal.task.list.view')}
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
}
