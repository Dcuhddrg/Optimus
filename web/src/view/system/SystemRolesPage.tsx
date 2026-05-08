import { useEffect, useState } from 'react';

import { listRoles, permissionTree, saveRolePermissions } from '../../api/system';
import { getRoleCodeLabel, getStatusLabel } from '../../locale/systemLabels';
import { useI18n } from '../../locale/useI18n';
import type { SystemPermission, SystemRole } from '../../types/system';
import styles from './SystemRolesPage.module.css';

export function SystemRolesPage() {
  const { locale, t } = useI18n();
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
        <p className="page-eyebrow">{t('system.roles.eyebrow')}</p>
        <h1 className={styles.title}>{t('system.roles.title')}</h1>
      </header>
      <div className={styles.grid}>
        {roles.map((role) => (
          <article className={styles.card} key={role.id}>
            <strong>{role.name}</strong>
            <p>{getRoleCodeLabel(role.code, locale)} · {getStatusLabel(role.status, locale)}</p>
            <button type="button" onClick={() => void grantCore(role)}>
              {t('system.roles.grantCorePermissions')}
            </button>
          </article>
        ))}
      </div>
    </section>
  );
}
