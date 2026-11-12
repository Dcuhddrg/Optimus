import { useEffect, useState } from 'react';

import { listUsers, updateUserExternal, updateUserStatus } from '../../api/system';
import {
  getRoleCodeLabel,
  getStatusLabel,
  getUserSourceLabel,
} from '../../locale/systemLabels';
import { useI18n } from '../../locale/useI18n';
import type { SystemUser } from '../../types/system';
import { LdapSyncPanel } from './LdapSyncPanel';
import styles from './SystemUsersPage.module.css';

function errorMessage(err: unknown) {
  return err instanceof Error ? err.message : String(err);
}

// ExternalEditor sets an external account's last valid day and sponsor
// (design 3.6); the sponsor must be an enabled employee.
function ExternalEditor({ user, sponsors, onSaved, onCancel }: { user: SystemUser; sponsors: SystemUser[]; onSaved: () => void; onCancel: () => void }) {
  const { t } = useI18n();
  const [expiresOn, setExpiresOn] = useState(user.expiresOn ?? '');
  const [sponsorId, setSponsorId] = useState(user.sponsorId ? String(user.sponsorId) : '');
  const [error, setError] = useState<string | null>(null);

  async function save() {
    setError(null);
    try {
      await updateUserExternal(user.id, { expiresOn: expiresOn || null, sponsorId: sponsorId ? Number(sponsorId) : null });
      onSaved();
    } catch (err) {
      setError(errorMessage(err));
    }
  }

  return (
    <div className={styles.editor} aria-label={t('system.users.editExternal')}>
      <label>
        {t('system.users.expiresOnLabel')}
        <input type="date" value={expiresOn} onChange={(event) => setExpiresOn(event.target.value)} />
      </label>
      <label>
        {t('system.users.sponsorLabel')}
        <select value={sponsorId} onChange={(event) => setSponsorId(event.target.value)}>
          <option value="">—</option>
          {sponsors.map((s) => (
            <option key={s.id} value={s.id}>{s.displayName || s.username} ({s.username})</option>
          ))}
        </select>
      </label>
      <button type="button" onClick={() => void save()}>{t('system.users.save')}</button>
      <button type="button" onClick={onCancel}>{t('system.users.cancel')}</button>
      {error ? <p role="alert">{error}</p> : null}
    </div>
  );
}

export function SystemUsersPage() {
  const { locale, t } = useI18n();
  const [users, setUsers] = useState<SystemUser[]>([]);
  const [editing, setEditing] = useState<number | null>(null);

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

  const sponsors = users.filter((u) => u.status === 'enabled' && (u.personType ?? 'employee') === 'employee');

  function externalLine(user: SystemUser) {
    const parts = [
      user.expiresOn ? t('system.users.expiresOn').replace('{date}', user.expiresOn) : t('system.users.noExpiry'),
      user.sponsor ? t('system.users.sponsor').replace('{name}', user.sponsor) : t('system.users.noSponsor'),
    ];
    if (user.expired) parts.push(t('system.users.expired'));
    if (user.sponsorInvalid) parts.push(t('system.users.sponsorInvalid'));
    if (!user.expiresOn || !user.sponsorId) parts.push(t('system.users.missingControls'));
    return parts.join(' · ');
  }

  return (
    <section className={styles.page}>
      <header>
        <p className="page-eyebrow">{t('system.users.eyebrow')}</p>
        <h1 className={styles.title}>{t('system.users.title')}</h1>
      </header>
      <LdapSyncPanel onDone={() => void loadUsers()} />
      <div className={styles.table}>
        {users.map((user) => (
          <div className={styles.row} key={user.id} data-username={user.username}>
            <div>
              <strong>{user.displayName || user.username}</strong>
              <p>
                {user.username} · {getUserSourceLabel(user.source, locale)}
                {user.personType === 'external' ? ` · ${t('system.users.external')}` : null}
                {user.department ? ` · ${user.department}` : null}
                {user.source === 'ldap' && !user.bound ? ` · ${t('system.users.pendingBinding')}` : null}
              </p>
              {user.personType === 'external' ? (
                <p className={user.expired || user.sponsorInvalid ? styles.warning : undefined}>
                  {externalLine(user)}{' '}
                  {editing === user.id ? null : (
                    <button type="button" className={styles.link} onClick={() => setEditing(user.id)}>
                      {t('system.users.editExternal')}
                    </button>
                  )}
                </p>
              ) : null}
              {editing === user.id ? (
                <ExternalEditor
                  user={user}
                  sponsors={sponsors}
                  onCancel={() => setEditing(null)}
                  onSaved={() => {
                    setEditing(null);
                    void loadUsers();
                  }}
                />
              ) : null}
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
