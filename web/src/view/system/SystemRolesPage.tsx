import { useEffect, useState } from 'react';

import { listRoles, permissionTree, saveRolePermissions } from '../../api/system';
import type { SystemPermission, SystemRole } from '../../types/system';
import styles from './SystemRolesPage.module.css';

export function SystemRolesPage() {
  const [roles, setRoles] = useState<SystemRole[]>([]);
  const [permissions, setPermissions] = useState<SystemPermission[]>([]);

  useEffect(() => {
    void Promise.all([listRoles(), permissionTree()]).then(([roleResult, permissionResult]) => {
      setRoles(roleResult.items);
      setPermissions(permissionResult.items);
    });
  }, []);

  async function grantCore(role: SystemRole) {
    await saveRolePermissions(role.id, permissions.slice(0, 8).map((permission) => permission.code));
  }

  return (
    <section className={styles.page}>
      <header>
        <p className="page-eyebrow">System / Roles</p>
        <h1 className={styles.title}>Roles</h1>
      </header>
      <div className={styles.grid}>
        {roles.map((role) => (
          <article className={styles.card} key={role.id}>
            <strong>{role.name}</strong>
            <p>{role.code} · {role.status}</p>
            <button type="button" onClick={() => void grantCore(role)}>
              Grant core permissions
            </button>
          </article>
        ))}
      </div>
    </section>
  );
}
