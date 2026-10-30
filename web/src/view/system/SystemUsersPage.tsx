import { useEffect, useState } from 'react';

import { listUsers, updateUserStatus } from '../../api/system';
import {
  getRoleCodeLabel,
  getStatusLabel,
  getUserSourceLabel,
} from '../../locale/systemLabels';
import { useI18n } from '../../locale/useI18n';
import type { SystemUser } from '../../types/system';
import styles from './SystemUsersPage.module.css';

export function SystemUsersPage() {
  const { locale, t } = useI18n();
  const [users, setUsers] = useState<SystemUser[]>([]);

  async function loadUsers() {
    const result = await listUsers();
    setUsers(result.list);
  }

  useEffect(() => {
    void loadUsers();
  }, []);

  async function toggleStatus(user: SystemUser) {
    const next = user.status === 'enabled' ? 'disabled' : 'enabled';
    await updateUserStatus(user.id, next);
    await loadUsers();
  }

  return (
    <section className={styles.page}>
      <header>
        <p className="page-eyebrow">{t('system.users.eyebrow')}</p>
        <h1 className={styles.title}>{t('system.users.title')}</h1>
      </header>
      <div className={styles.table}>
        {users.map((user) => (
          <div className={styles.row} key={user.id}>
            <div>
              <strong>{user.displayName || user.username}</strong>
              <p>{user.username} · {getUserSourceLabel(user.source, locale)}</p>
            </div>
            <div>
              {user.roles?.length
                ? user.roles.map((role) => getRoleCodeLabel(role, locale)).join(', ')
                : t('system.users.emptyRoles')}
            </div>
            <button type="button" onClick={() => void toggleStatus(user)}>
              {getStatusLabel(user.status, locale)}
            </button>
          </div>
        ))}
      </div>
    </section>
  );
}
