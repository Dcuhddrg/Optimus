import { FormEvent, useEffect, useState } from 'react';

import {
  createCloudAccount,
  deleteCloudAccount,
  listCloudAccounts,
  syncCloudAccount,
  updateCloudAccountTags,
} from '../../api/cloud';
import { TagChip } from '../../components/tag/TagChip';
import { TagEditor } from '../../components/tag/TagEditor';
import { TagFilter } from '../../components/tag/TagFilter';
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
  const [editingTagsId, setEditingTagsId] = useState<number | null>(null);
  const [tagFilters, setTagFilters] = useState<string[]>([]);

  async function load() {
    try {
      const result = await listCloudAccounts(tagFilters);
      setAccounts(result.items);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.loadFailed'));
    }
  }

  useEffect(() => {
    void load();
  }, [tagFilters]);

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

  async function handleTagsSubmit(account: CloudAccount, tags: Record<string, string>) {
    await updateCloudAccountTags(account.id, { tags });
    setEditingTagsId(null);
    await load();
  }

  return (
    <section className={styles.page}>
      {error ? <div className={styles.error}>{error}</div> : null}

      <TagFilter active={tagFilters} onChange={setTagFilters} totalCount={accounts.length} />

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
          <article className={styles.rowWrapper} key={account.id}>
            <div className={styles.row}>
              <div className={styles.identity}>
                <strong>{account.name}</strong>
                <span>
                  {account.provider} · {account.accountId}
                </span>
                <em>{account.status}</em>
              </div>
              <div className={styles.tags}>
                {Object.entries(account.tags).length === 0 ? (
                  <span className={styles.untagged}>{t('cloud.accounts.untagged')}</span>
                ) : (
                  Object.entries(account.tags).map(([key, value]) => (
                    <TagChip key={key} tagKey={key} value={value} />
                  ))
                )}
              </div>
              <div className={styles.actions}>
                <button
                  type="button"
                  onClick={() => setEditingTagsId(editingTagsId === account.id ? null : account.id)}
                >
                  {editingTagsId === account.id ? t('cloud.accounts.closeTags') : t('cloud.accounts.editTags')}
                </button>
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
            </div>
            {editingTagsId === account.id ? (
              <div className={styles.editorRow}>
                <TagEditor
                  initialTags={account.tags}
                  onSubmit={(tags) => handleTagsSubmit(account, tags)}
                  onCancel={() => setEditingTagsId(null)}
                />
              </div>
            ) : null}
          </article>
        ))}
        {accounts.length === 0 ? <p className={styles.empty}>{t('cloud.accounts.empty')}</p> : null}
      </div>
    </section>
  );
}
