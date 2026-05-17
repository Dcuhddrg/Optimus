import { useEffect, useState } from 'react';

import { listCloudAccounts, listCloudAssets } from '../../api/cloud';
import { TagChip } from '../../components/tag/TagChip';
import { TagFilter } from '../../components/tag/TagFilter';
import { useI18n } from '../../locale/useI18n';
import type { CloudAccount, CloudAsset } from '../../types/cloud';
import styles from './CloudAssetsPage.module.css';

export function CloudAssetsPage() {
  const { t } = useI18n();
  const [assets, setAssets] = useState<CloudAsset[]>([]);
  const [total, setTotal] = useState(0);
  const [accountsById, setAccountsById] = useState<Record<number, CloudAccount>>({});
  const [tagFilters, setTagFilters] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);

  async function load() {
    try {
      const [assetsResult, accountsResult] = await Promise.all([
        listCloudAssets(tagFilters),
        listCloudAccounts(),
      ]);
      setAssets(assetsResult.items);
      setTotal(assetsResult.total);
      const indexed: Record<number, CloudAccount> = {};
      for (const account of accountsResult.items) {
        indexed[account.id] = account;
      }
      setAccountsById(indexed);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.assets.loadFailed'));
    }
  }

  useEffect(() => {
    void load();
  }, [tagFilters]);

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <div>
          <p className="page-eyebrow">{t('cloud.assets.eyebrow')}</p>
          <h1 className={styles.title}>{t('cloud.assets.title')}</h1>
          <p className={styles.sub}>{t('cloud.assets.sub')}</p>
        </div>
      </header>

      {error ? <div className={styles.error}>{error}</div> : null}

      <TagFilter active={tagFilters} onChange={setTagFilters} totalCount={total} />

      <div className={styles.list}>
        <div className={styles.tableHeader} aria-hidden="true">
          <span>{t('cloud.assets.name')}</span>
          <span>{t('cloud.assets.resourceId')}</span>
          <span>{t('cloud.assets.type')}</span>
          <span>{t('cloud.assets.account')}</span>
          <span>{t('cloud.assets.inheritedTags')}</span>
        </div>
        {assets.map((asset) => {
          const account = accountsById[asset.accountId];
          const accountTags = account?.tags ?? {};
          return (
            <article className={styles.row} key={asset.id}>
              <strong>{asset.name}</strong>
              <code>{asset.resourceId}</code>
              <span>{asset.resourceType}</span>
              <span>{account?.name ?? `#${asset.accountId}`}</span>
              <div className={styles.tags}>
                {Object.entries(accountTags).length === 0 ? (
                  <span className={styles.untagged}>{t('cloud.accounts.untagged')}</span>
                ) : (
                  Object.entries(accountTags).map(([key, value]) => (
                    <TagChip key={key} tagKey={key} value={value} />
                  ))
                )}
              </div>
            </article>
          );
        })}
        {assets.length === 0 ? <p className={styles.empty}>{t('cloud.assets.empty')}</p> : null}
      </div>
    </section>
  );
}
