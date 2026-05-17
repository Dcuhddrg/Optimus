import { FormEvent, useEffect, useState } from 'react';

import { createCloudAccount, deleteCloudAccount, listCloudAccounts, syncCloudAccount } from '../../api/cloud';
import { useI18n } from '../../locale/useI18n';
import type { CloudAccount } from '../../types/cloud';
import styles from './CloudAccountsPage.module.css';

interface Draft {
  name: string;
  accountId: string;
  accessKeyId: string;
  accessKeySecret: string;
}

const emptyDraft: Draft = { name: '', accountId: '', accessKeyId: '', accessKeySecret: '' };

export function CloudAccountsPage() {
  const { t } = useI18n();
  const [accounts, setAccounts] = useState<CloudAccount[]>([]);
  const [draft, setDraft] = useState<Draft>(emptyDraft);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<number | null>(null);

  async function load() {
    try {
      const result = await listCloudAccounts();
      setAccounts(result.items);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.loadFailed'));
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function handleCreate(event: FormEvent) {
    event.preventDefault();
    setError(null);
    try {
      await createCloudAccount({
        name: draft.name,
        provider: 'aliyun',
        accountId: draft.accountId,
        credentialType: 'ak_sk',
        credential: {
          access_key_id: draft.accessKeyId,
          access_key_secret: draft.accessKeySecret,
        },
      });
      setDraft(emptyDraft);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.saveFailed'));
    }
  }

  async function handleDelete(account: CloudAccount) {
    if (!window.confirm(t('cloud.accounts.deleteConfirm'))) {
      return;
    }
    setBusyId(account.id);
    setError(null);
    try {
      await deleteCloudAccount(account.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.deleteFailed'));
    } finally {
      setBusyId(null);
    }
  }

  async function handleSync(account: CloudAccount) {
    setBusyId(account.id);
    setError(null);
    try {
      await syncCloudAccount(account.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.syncFailed'));
    } finally {
      setBusyId(null);
    }
  }

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <div>
          <p className="page-eyebrow">{t('cloud.accounts.eyebrow')}</p>
          <h1 className={styles.title}>{t('cloud.accounts.title')}</h1>
          <p className={styles.sub}>{t('cloud.accounts.sub')}</p>
        </div>
      </header>

      {error ? <div className={styles.error}>{error}</div> : null}

      <form className={styles.createBar} onSubmit={(event) => void handleCreate(event)}>
        <label>
          <span>{t('cloud.accounts.name')}</span>
          <input
            value={draft.name}
            onChange={(event) => setDraft((value) => ({ ...value, name: event.target.value }))}
            required
          />
        </label>
        <label>
          <span>{t('cloud.accounts.accountId')}</span>
          <input
            value={draft.accountId}
            onChange={(event) => setDraft((value) => ({ ...value, accountId: event.target.value }))}
            required
          />
        </label>
        <label>
          <span>{t('cloud.accounts.accessKeyId')}</span>
          <input
            value={draft.accessKeyId}
            onChange={(event) => setDraft((value) => ({ ...value, accessKeyId: event.target.value }))}
            required
          />
        </label>
        <label>
          <span>{t('cloud.accounts.accessKeySecret')}</span>
          <input
            type="password"
            value={draft.accessKeySecret}
            onChange={(event) => setDraft((value) => ({ ...value, accessKeySecret: event.target.value }))}
            required
          />
        </label>
        <button type="submit">{t('cloud.accounts.create')}</button>
      </form>

      <div className={styles.list}>
        {accounts.map((account) => (
          <article className={styles.row} key={account.id}>
            <div className={styles.identity}>
              <strong>{account.name}</strong>
              <span>
                {account.provider} · {account.accountId}
              </span>
              <em>{account.status}</em>
            </div>
            <div className={styles.actions}>
              <button type="button" onClick={() => void handleSync(account)} disabled={busyId === account.id}>
                {t('cloud.accounts.sync')}
              </button>
              <button
                type="button"
                className={styles.danger}
                onClick={() => void handleDelete(account)}
                disabled={busyId === account.id}
              >
                {t('cloud.accounts.delete')}
              </button>
            </div>
          </article>
        ))}
        {accounts.length === 0 ? <p className={styles.empty}>{t('cloud.accounts.empty')}</p> : null}
      </div>
    </section>
  );
}
