import { type ChangeEvent, useEffect, useMemo, useState } from 'react';

import { listCloudAccounts, listCloudAssets } from '../../api/cloud';
import { TagChip } from '../../components/tag/TagChip';
import { TagFilter } from '../../components/tag/TagFilter';
import { useI18n } from '../../locale/useI18n';
import type { CloudAccount, CloudAsset } from '../../types/cloud';
import styles from './CloudAssetsPage.module.css';

interface BasicFilters {
  accountId: string;
  provider: string;
  region: string;
  resourceType: string;
  status: string;
  keyword: string;
}

const emptyBasicFilters: BasicFilters = {
  accountId: '',
  provider: '',
  region: '',
  resourceType: '',
  status: '',
  keyword: '',
};

export function CloudAssetsPage() {
  const { t } = useI18n();
  const [assets, setAssets] = useState<CloudAsset[]>([]);
  const [total, setTotal] = useState(0);
  const [accountsById, setAccountsById] = useState<Record<number, CloudAccount>>({});
  const [accountsList, setAccountsList] = useState<CloudAccount[]>([]);
  const [tagFilters, setTagFilters] = useState<string[]>([]);
  const [basic, setBasic] = useState<BasicFilters>(emptyBasicFilters);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [error, setError] = useState<string | null>(null);

  const providers = useMemo(
    () => Array.from(new Set(accountsList.map((account) => account.provider))).sort(),
    [accountsList],
  );
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  async function load() {
    try {
      const [assetsResult, accountsResult] = await Promise.all([
        listCloudAssets({
          accountId: basic.accountId ? Number(basic.accountId) : undefined,
          provider: basic.provider || undefined,
          region: basic.region || undefined,
          resourceType: basic.resourceType || undefined,
          status: basic.status || undefined,
          keyword: basic.keyword || undefined,
          tag: tagFilters.length > 0 ? tagFilters : undefined,
          page,
          pageSize,
        }),
        listCloudAccounts(),
      ]);
      setAssets(assetsResult.items);
      setTotal(assetsResult.total);
      setAccountsList(accountsResult.items);
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
  }, [tagFilters, basic, page, pageSize]);

  function updateBasic(key: keyof BasicFilters, event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) {
    const value = event.target.value;
    setBasic((prev) => ({ ...prev, [key]: value }));
    setPage(1);
  }

  function handleTagFiltersChange(next: string[]) {
    setTagFilters(next);
    setPage(1);
  }

  return (
    <section className={styles.page}>
      {error ? <div className={styles.error}>{error}</div> : null}

      <TagFilter active={tagFilters} onChange={handleTagFiltersChange} totalCount={total} />

      <div className={styles.filterRow}>
        <select value={basic.accountId} onChange={(event) => updateBasic('accountId', event)}>
          <option value="">{t('cloud.assets.allAccounts')}</option>
          {accountsList.map((account) => (
            <option key={account.id} value={String(account.id)}>
              {account.name}
            </option>
          ))}
        </select>
        <select value={basic.provider} onChange={(event) => updateBasic('provider', event)}>
          <option value="">{t('cloud.assets.allProviders')}</option>
          {providers.map((provider) => (
            <option key={provider} value={provider}>
              {provider}
            </option>
          ))}
        </select>
        <input
          value={basic.region}
          onChange={(event) => updateBasic('region', event)}
          placeholder={t('cloud.assets.regionPlaceholder')}
        />
        <input
          value={basic.resourceType}
          onChange={(event) => updateBasic('resourceType', event)}
          placeholder={t('cloud.assets.resourceTypePlaceholder')}
        />
        <input
          value={basic.status}
          onChange={(event) => updateBasic('status', event)}
          placeholder={t('cloud.assets.statusPlaceholder')}
        />
        <input
          value={basic.keyword}
          onChange={(event) => updateBasic('keyword', event)}
          placeholder={t('cloud.assets.keywordPlaceholder')}
        />
      </div>

      <div className={styles.list}>
        <div className={styles.tableHeader} aria-hidden="true">
          <span>{t('cloud.assets.name')}</span>
          <span>{t('cloud.assets.resourceId')}</span>
          <span>{t('cloud.assets.type')}</span>
          <span>{t('cloud.assets.region')}</span>
          <span>{t('cloud.assets.status')}</span>
          <span>{t('cloud.assets.account')}</span>
          <span>{t('cloud.assets.lastSeen')}</span>
          <span>{t('cloud.assets.inheritedTags')}</span>
          <span>{t('cloud.assets.nativeTags')}</span>
        </div>
        {assets.map((asset) => {
          const account = accountsById[asset.accountId];
          const accountTags = account?.tags ?? {};
          const nativeTags = parseNativeTags(asset.tagsJson);
          const isDeleted = asset.deletedAt !== null;
          return (
            <article
              className={[styles.row, isDeleted ? styles.rowDeleted : ''].filter(Boolean).join(' ')}
              key={asset.id}
              title={isDeleted ? t('cloud.assets.deletedTitle') : undefined}
            >
              <strong>
                {asset.name}
                {isDeleted ? <span className={styles.deletedBadge}>{t('cloud.assets.deletedBadge')}</span> : null}
              </strong>
              <code>{asset.resourceId}</code>
              <span>{asset.resourceType}</span>
              <span>{asset.region}</span>
              <span className={styles.status}>{asset.status || '-'}</span>
              <span>{account?.name ?? `#${asset.accountId}`}</span>
              <span className={styles.timestamp}>{formatTimestamp(asset.lastSeenAt)}</span>
              <div className={styles.tags}>
                {Object.entries(accountTags).length === 0 ? (
                  <span className={styles.untagged}>{t('cloud.accounts.untagged')}</span>
                ) : (
                  Object.entries(accountTags).map(([key, value]) => (
                    <TagChip key={key} tagKey={key} value={value} />
                  ))
                )}
              </div>
              <div className={styles.tags}>
                {Object.entries(nativeTags).length === 0 ? (
                  <span className={styles.untagged}>{t('cloud.assets.noNativeTags')}</span>
                ) : (
                  Object.entries(nativeTags).map(([key, value]) => (
                    <TagChip key={key} tagKey={key} value={value} />
                  ))
                )}
              </div>
            </article>
          );
        })}
        {assets.length === 0 ? <p className={styles.empty}>{t('cloud.assets.empty')}</p> : null}
      </div>

      <div className={styles.pagination}>
        <span className={styles.paginationSummary}>
          {t('cloud.assets.totalLabel')} {total} {t('cloud.assets.entriesLabel')}
        </span>
        <div className={styles.paginationControls}>
          <button type="button" disabled={page <= 1} onClick={() => setPage((prev) => Math.max(1, prev - 1))}>
            {t('cloud.assets.prevPage')}
          </button>
          <span className={styles.paginationPosition}>
            {page} / {totalPages}
          </span>
          <button
            type="button"
            disabled={page >= totalPages}
            onClick={() => setPage((prev) => Math.min(totalPages, prev + 1))}
          >
            {t('cloud.assets.nextPage')}
          </button>
          <select
            value={String(pageSize)}
            onChange={(event) => {
              setPageSize(Number(event.target.value));
              setPage(1);
            }}
          >
            <option value="20">20 / {t('cloud.assets.perPage')}</option>
            <option value="50">50 / {t('cloud.assets.perPage')}</option>
            <option value="100">100 / {t('cloud.assets.perPage')}</option>
          </select>
        </div>
      </div>
    </section>
  );
}

function formatTimestamp(iso: string): string {
  if (!iso) return '-';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return iso;
  }
  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function parseNativeTags(tagsJson: string): Record<string, string> {
  if (!tagsJson) return {};
  try {
    const parsed = JSON.parse(tagsJson);
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      const out: Record<string, string> = {};
      for (const [key, value] of Object.entries(parsed)) {
        if (typeof value === 'string') {
          out[key] = value;
        } else if (value != null) {
          out[key] = String(value);
        }
      }
      return out;
    }
  } catch {
    // Ignore invalid provider tag payloads and render the asset as untagged.
  }
  return {};
}
