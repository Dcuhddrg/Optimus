import { useEffect, useState } from 'react';

import { listUsers, updateUserStatus } from '../../api/system';
import type { SystemUser } from '../../types/system';
import styles from './SystemUsersPage.module.css';

export function SystemUsersPage() {
  const [users, setUsers] = useState<SystemUser[]>([]);

  async function loadUsers() {
    const result = await listUsers();
    setUsers(result.items);
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
        <p className="page-eyebrow">System / Users</p>
        <h1 className={styles.title}>Users</h1>
      </header>
      <div className={styles.table}>
        {users.map((user) => (
          <div className={styles.row} key={user.id}>
            <div>
              <strong>{user.displayName || user.username}</strong>
              <p>{user.username} · {user.source}</p>
            </div>
            <div>{user.roles?.join(', ') || 'No roles'}</div>
            <button type="button" onClick={() => void toggleStatus(user)}>
              {user.status}
            </button>
          </div>
        ))}
      </div>
    </section>
  );
}
