import { FormEvent, useCallback, useEffect, useState } from 'react';

import { createRole, deleteRole, listRoles, permissionTree, saveRolePermissions, updateRole } from '../../api/system';
import { getPermissionCodeLabel, getRoleCodeLabel } from '../../locale/systemLabels';
import { useI18n } from '../../locale/useI18n';
import type { SystemPermission, SystemRoleWithPermissions } from '../../types/system';
import styles from './SystemRolesPage.module.css';

// Built-in roles (design 3.2, 3.4): neither can be deleted; admin always keeps
// every system.* permission and is never an approval role.
const ADMIN = 'admin';
const PROTECTED = new Set([ADMIN, 'user']);
const lockedForAdmin = (role: SystemRoleWithPermissions, code: string) => role.code === ADMIN && code.startsWith('system.');

function errorMessage(err: unknown) {
  return err instanceof Error ? err.message : String(err);
}

export function SystemRolesPage() {
  const { locale, t } = useI18n();
  const [roles, setRoles] = useState<SystemRoleWithPermissions[]>([]);
  const [permissions, setPermissions] = useState<SystemPermission[]>([]);
  const [status, setStatus] = useState<'loading' | 'ready' | 'failed'>('loading');
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [editing, setEditing] = useState<number | null>(null);
  const [draft, setDraft] = useState<Set<string>>(new Set());
  const [form, setForm] = useState({ code: '', name: '', description: '', approvable: false });

  const load = useCallback(async () => {
    setStatus('loading');
    try {
      const [roleResult, permissionResult] = await Promise.all([listRoles(), permissionTree()]);
      setRoles(roleResult.list);
      setPermissions(permissionResult.list);
      setStatus('ready');
    } catch {
      setStatus('failed');
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function run(action: () => Promise<unknown>) {
    setError(null);
    setNotice(null);
    try {
      await action();
      setNotice(t('system.roles.saved'));
      await load();
      return true;
    } catch (err) {
      setError(errorMessage(err));
      return false;
    }
  }

  function startEditing(role: SystemRoleWithPermissions) {
    setEditing(role.id);
    setDraft(new Set(role.permissionCodes));
  }

  function toggle(code: string) {
    setDraft((current) => {
      const next = new Set(current);
      if (next.has(code)) next.delete(code);
      else next.add(code);
      return next;
    });
  }

  async function savePermissions(role: SystemRoleWithPermissions) {
    const codes = new Set(draft);
    for (const permission of permissions) if (lockedForAdmin(role, permission.code)) codes.add(permission.code);
    if (await run(() => saveRolePermissions(role.id, [...codes]))) setEditing(null);
  }

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (await run(() => createRole(form))) setForm({ code: '', name: '', description: '', approvable: false });
  }

  async function handleDelete(role: SystemRoleWithPermissions) {
    if (!window.confirm(t('system.roles.confirmDelete'))) return;
    await run(() => deleteRole(role.id));
  }

  const groups: { label: string; items: SystemPermission[] }[] = [
    { label: t('system.roles.menuPermissions'), items: permissions.filter((p) => p.type === 'menu') },
    { label: t('system.roles.apiPermissions'), items: permissions.filter((p) => p.type === 'api') },
  ];

  return (
    <section className={styles.page}>
      <header>
        <p className="page-eyebrow">{t('system.roles.eyebrow')}</p>
        <h1 className={styles.title}>{t('system.roles.title')}</h1>
      </header>

      {error ? <p className={styles.error} role="alert">{error}</p> : null}
      {notice ? <p className={styles.notice} role="status">{notice}</p> : null}

      <form className={styles.card} onSubmit={(event) => void handleCreate(event)} aria-label={t('system.roles.create')}>
        <strong>{t('system.roles.create')}</strong>
        <div className={styles.fields}>
          <label>{t('system.roles.code')}<input required value={form.code} onChange={(e) => setForm({ ...form, code: e.target.value })} /></label>
          <label>{t('system.roles.name')}<input required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></label>
          <label>{t('system.roles.description')}<input value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} /></label>
          <label className={styles.inline}>
            <input type="checkbox" checked={form.approvable} onChange={(e) => setForm({ ...form, approvable: e.target.checked })} />
            {t('system.roles.approvable')}
          </label>
        </div>
        <button type="submit">{t('system.roles.create')}</button>
      </form>

      {status === 'loading' ? <p>{t('system.roles.loading')}</p> : null}
      {status === 'failed' ? (
        <p>
          {t('system.roles.loadFailed')} <button type="button" onClick={() => void load()}>{t('system.roles.retry')}</button>
        </p>
      ) : null}

      <div className={styles.grid}>
        {roles.map((role) => (
          <article className={styles.card} key={role.id} data-role={role.code}>
            <strong>{role.name}</strong>
            <p>
              {getRoleCodeLabel(role.code, locale)}
              {role.approvable ? <span className={styles.badge}>{t('system.roles.approvableBadge')}</span> : null}
              {PROTECTED.has(role.code) ? <span className={styles.badge}>{t('system.roles.protectedRole')}</span> : null}
            </p>
            {role.description ? <p>{role.description}</p> : null}
            {role.code !== ADMIN ? (
              <label className={styles.inline}>
                <input
                  type="checkbox"
                  checked={role.approvable}
                  onChange={(e) => void run(() => updateRole(role.id, { name: role.name, description: role.description, approvable: e.target.checked }))}
                />
                {t('system.roles.approvable')}
              </label>
            ) : null}

            {editing === role.id ? (
              <div className={styles.editor}>
                {role.code === ADMIN ? <p>{t('system.roles.adminLocked')}</p> : null}
                {groups.map((group) => (
                  <fieldset key={group.label}>
                    <legend>{group.label}</legend>
                    {group.items.map((permission) => {
                      const locked = lockedForAdmin(role, permission.code);
                      return (
                        <label className={styles.inline} key={permission.code}>
                          <input
                            type="checkbox"
                            checked={locked || draft.has(permission.code)}
                            disabled={locked}
                            onChange={() => toggle(permission.code)}
                          />
                          {getPermissionCodeLabel(permission.code, locale)}
                        </label>
                      );
                    })}
                  </fieldset>
                ))}
                <div className={styles.actions}>
                  <button type="button" onClick={() => void savePermissions(role)}>{t('system.roles.save')}</button>
                  <button type="button" onClick={() => setEditing(null)}>{t('system.roles.cancel')}</button>
                </div>
              </div>
            ) : (
              <div className={styles.actions}>
                <button type="button" onClick={() => startEditing(role)}>{t('system.roles.editPermissions')}</button>
                {!PROTECTED.has(role.code) ? (
                  <button type="button" onClick={() => void handleDelete(role)}>{t('system.roles.delete')}</button>
                ) : null}
              </div>
            )}
          </article>
        ))}
      </div>
    </section>
  );
}
