import { useState } from 'react';

import { executeLDAPSync, previewLDAPSync } from '../../api/system';
import type { MessageKey } from '../../locale/messages';
import { useI18n } from '../../locale/useI18n';
import type { LDAPSyncCategory, LDAPSyncItem, LDAPSyncPreview, LDAPSyncResult } from '../../types/system';
import styles from './LdapSyncPanel.module.css';

// Display order; the first group needs no decision.
const CATEGORIES: LDAPSyncCategory[] = [
  'unidentifiable',
  'conflict',
  'pendingBinding',
  'outOfScope',
  'unconfirmable',
  'autoDisable',
  'create',
  'update',
  'disabledInfo',
];

type BindingChoice = 'leave' | 'bind' | 'separate';

function errorMessage(err: unknown) {
  return err instanceof Error ? err.message : String(err);
}

function itemKey(item: LDAPSyncItem) {
  return item.account ? `a${item.account.id}` : `e${item.entry?.externalId ?? ''}${item.entry?.dn ?? ''}`;
}

function itemName(item: LDAPSyncItem) {
  return item.account?.username ?? item.entry?.username ?? item.entry?.dn ?? '';
}

function formatTime(value: string | null | undefined) {
  return value ? new Date(value).toLocaleString() : '—';
}

// LdapSyncPanel previews a full LDAP sync and executes it with the
// administrator's choices (design 3.3 D6). Nothing is written until Execute.
export function LdapSyncPanel({ onDone }: { onDone: () => void }) {
  const { t } = useI18n();
  const [preview, setPreview] = useState<LDAPSyncPreview | null>(null);
  const [result, setResult] = useState<LDAPSyncResult | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [disable, setDisable] = useState<Set<number>>(new Set());
  const [binding, setBinding] = useState<Record<number, BindingChoice>>({});
  const [separate, setSeparate] = useState<Set<number>>(new Set());

  async function runPreview() {
    setBusy(true);
    setError(null);
    setResult(null);
    try {
      setPreview(await previewLDAPSync());
      setDisable(new Set());
      setBinding({});
      setSeparate(new Set());
    } catch (err) {
      setPreview(null);
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  async function runExecute() {
    if (!preview) return;
    const bind: number[] = [];
    const separated = new Set(separate);
    for (const [id, choice] of Object.entries(binding)) {
      if (choice === 'bind') bind.push(Number(id));
      if (choice === 'separate') separated.add(Number(id));
    }
    setBusy(true);
    setError(null);
    try {
      setResult(await executeLDAPSync(preview.previewId, { disable: [...disable], bind, separate: [...separated] }));
      setPreview(null);
      onDone();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  function toggle(set: Set<number>, update: (next: Set<number>) => void, id: number) {
    const next = new Set(set);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    update(next);
  }

  // Bulk confirmation covers only bindings whose timeline looks normal.
  function bindAllPlausible() {
    if (!preview) return;
    const next = { ...binding };
    for (const item of preview.items) {
      if (item.category === 'pendingBinding' && item.account && !item.suspicious) next[item.account.id] = 'bind';
    }
    setBinding(next);
  }

  function controls(item: LDAPSyncItem) {
    const id = item.account?.id;
    if (id === undefined) return null;
    switch (item.category) {
      case 'outOfScope':
      case 'unconfirmable':
        return (
          <label className={styles.choice}>
            <input type="checkbox" disabled={preview?.disablesBlocked} checked={disable.has(id)} onChange={() => toggle(disable, setDisable, id)} />
            {t('system.ldapSync.choiceDisable')}
          </label>
        );
      case 'conflict':
        // Local accounts (the built-in admin) are never retired by a sync.
        if (item.reason === 'local_account') return null;
        return (
          <label className={styles.choice}>
            <input type="checkbox" checked={separate.has(id)} onChange={() => toggle(separate, setSeparate, id)} />
            {t('system.ldapSync.choiceSeparate')}
          </label>
        );
      case 'pendingBinding':
        return (
          <select
            aria-label={t('system.ldapSync.bindingChoice')}
            value={binding[id] ?? 'leave'}
            onChange={(event) => setBinding({ ...binding, [id]: event.target.value as BindingChoice })}
          >
            <option value="leave">{t('system.ldapSync.choiceLeave')}</option>
            <option value="bind">{t('system.ldapSync.choiceBind')}</option>
            <option value="separate">{t('system.ldapSync.choiceSeparate')}</option>
          </select>
        );
      default:
        return null;
    }
  }

  function details(item: LDAPSyncItem) {
    const parts: string[] = [];
    if (item.reason) parts.push(t(`system.ldapSync.reason.${item.reason}` as MessageKey));
    if (item.changes?.length) parts.push(item.changes.map((c) => `${c.field}: ${c.before || '—'} → ${c.after || '—'}`).join('; '));
    if (item.restrictedRoles?.length) parts.push(t('system.ldapSync.restrictedRoles').replace('{roles}', item.restrictedRoles.join(', ')));
    if (item.ownedProjects?.length) parts.push(t('system.ldapSync.ownedProjects').replace('{projects}', item.ownedProjects.join(', ')));
    if (item.category === 'pendingBinding' || item.category === 'conflict') {
      const a = item.account;
      const e = item.entry;
      if (a) parts.push(`${t('system.ldapSync.optimusSide')}: ${a.displayName} · ${a.email || '—'} · ${formatTime(a.createdAt)} · ${(a.roles ?? []).join(', ') || '—'}${a.externalId ? ` · ${a.externalId}` : ''}`);
      if (e) parts.push(`${t('system.ldapSync.directorySide')}: ${e.displayName} · ${e.email || '—'} · ${e.externalId} · ${formatTime(e.createdAt)}`);
    }
    return parts;
  }

  const grouped = new Map<LDAPSyncCategory, LDAPSyncItem[]>();
  for (const item of preview?.items ?? []) grouped.set(item.category, [...(grouped.get(item.category) ?? []), item]);

  return (
    <section className={styles.panel} aria-label={t('system.ldapSync.title')}>
      <div className={styles.header}>
        <strong>{t('system.ldapSync.title')}</strong>
        <button type="button" disabled={busy} onClick={() => void runPreview()}>
          {preview ? t('system.ldapSync.previewAgain') : t('system.ldapSync.preview')}
        </button>
      </div>
      {error ? <p className={styles.error} role="alert">{error}</p> : null}

      {preview ? (
        <>
          {preview.disablesBlocked ? (
            <p className={styles.warning} role="alert">
              {t('system.ldapSync.disablesBlocked').replace('{count}', String(grouped.get('unidentifiable')?.length ?? 0))}
            </p>
          ) : null}
          {preview.configChanged ? <p className={styles.warning}>{t('system.ldapSync.configChanged')}</p> : null}
          {preview.items.length === 0 ? <p>{t('system.ldapSync.nothingToDo')}</p> : null}

          {CATEGORIES.filter((c) => grouped.has(c)).map((category) => (
            <div className={styles.group} key={category} data-category={category}>
              <div className={styles.groupHeader}>
                <h3>
                  {t(`system.ldapSync.category.${category}` as MessageKey)} ({grouped.get(category)!.length})
                </h3>
                {category === 'pendingBinding' ? (
                  <button type="button" onClick={bindAllPlausible}>{t('system.ldapSync.bindAllPlausible')}</button>
                ) : null}
              </div>
              <p className={styles.hint}>{t(`system.ldapSync.categoryHint.${category}` as MessageKey)}</p>
              <ul className={styles.items}>
                {grouped.get(category)!.map((item) => (
                  <li key={itemKey(item)} className={item.suspicious ? styles.suspicious : undefined} data-username={itemName(item)}>
                    <div>
                      <strong>{itemName(item)}</strong>
                      {item.suspicious ? <span className={styles.flag}>{t('system.ldapSync.suspicious')}</span> : null}
                      {details(item).map((line) => (
                        <p key={line}>{line}</p>
                      ))}
                    </div>
                    {controls(item)}
                  </li>
                ))}
              </ul>
            </div>
          ))}

          <div className={styles.footer}>
            <span>{t('system.ldapSync.expiresAt').replace('{time}', formatTime(preview.expiresAt))}</span>
            <button type="button" className={styles.primary} disabled={busy} onClick={() => void runExecute()}>
              {t('system.ldapSync.execute')}
            </button>
          </div>
        </>
      ) : null}

      {result ? (
        <div className={styles.result} role="status">
          <strong>{t(result.status === 'completed' ? 'system.ldapSync.completed' : 'system.ldapSync.partial')}</strong>
          <ul className={styles.items}>
            {result.outcomes.map((o, index) => (
              <li key={`${o.action}-${o.accountId ?? o.username}-${index}`} data-outcome={o.outcome}>
                <span>{o.username}</span>
                <span>{t(`system.ldapSync.action.${o.action}` as MessageKey)}</span>
                <span>{t(`system.ldapSync.outcome.${o.outcome}` as MessageKey)}</span>
                {o.reason ? <span>{o.reason}</span> : null}
                {o.restrictedRoles?.length ? <span>{t('system.ldapSync.restrictedRoles').replace('{roles}', o.restrictedRoles.join(', '))}</span> : null}
                {o.ownedProjects?.length ? <span>{t('system.ldapSync.ownedProjects').replace('{projects}', o.ownedProjects.join(', '))}</span> : null}
              </li>
            ))}
          </ul>
        </div>
      ) : null}
    </section>
  );
}
