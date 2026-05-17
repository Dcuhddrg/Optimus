import { FormEvent, useState } from 'react';

import { useI18n } from '../../locale/useI18n';
import { TagChip } from './TagChip';
import styles from './TagFilter.module.css';

interface TagFilterProps {
  active: string[];
  onChange: (tokens: string[]) => void;
  environmentCounts?: Record<string, number>;
  totalCount?: number;
}

const ENV_VALUES = ['prod', 'dev', 'qa', 'orphan'] as const;

export function TagFilter({ active, onChange, environmentCounts, totalCount }: TagFilterProps) {
  const { t } = useI18n();
  const [input, setInput] = useState('');

  function isActiveEnv(value: string | null): boolean {
    const target = value === null ? '__none__:environment' : `environment=${value}`;
    return active.includes(target);
  }

  function toggleEnvChip(value: string | null) {
    const target = value === null ? '__none__:environment' : `environment=${value}`;
    const cleared = active.filter((token) => !token.startsWith('environment=') && !token.startsWith('__none__:environment'));
    if (active.includes(target)) {
      onChange(cleared);
      return;
    }
    onChange([...cleared, target]);
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const trimmed = input.trim();
    if (!trimmed) {
      return;
    }
    if (active.includes(trimmed)) {
      setInput('');
      return;
    }
    onChange([...active, trimmed]);
    setInput('');
  }

  function removeToken(token: string) {
    onChange(active.filter((item) => item !== token));
  }

  const customTokens = active.filter((token) => !token.startsWith('environment=') && !token.startsWith('__none__:environment'));
  const hasEnvironmentFilter = ENV_VALUES.some((value) => isActiveEnv(value)) || isActiveEnv(null);

  return (
    <div className={styles.filter}>
      <div className={styles.chipRow}>
        <span className={styles.label}>{t('cloud.tags.environmentLabel')}:</span>
        <button
          type="button"
          className={[styles.chip, !hasEnvironmentFilter ? styles.chipActive : ''].filter(Boolean).join(' ')}
          onClick={() => onChange(customTokens)}
        >
          {t('cloud.tags.all')}
          {totalCount !== undefined ? <strong>{totalCount}</strong> : null}
        </button>
        {ENV_VALUES.map((value) => (
          <button
            key={value}
            type="button"
            className={[styles.chip, isActiveEnv(value) ? styles.chipActive : ''].filter(Boolean).join(' ')}
            onClick={() => toggleEnvChip(value)}
          >
            <TagChip tagKey="env" value={value} />
            {environmentCounts && environmentCounts[value] !== undefined ? <strong>{environmentCounts[value]}</strong> : null}
          </button>
        ))}
        <button
          type="button"
          className={[styles.chip, isActiveEnv(null) ? styles.chipActive : ''].filter(Boolean).join(' ')}
          onClick={() => toggleEnvChip(null)}
        >
          {t('cloud.tags.untagged')}
          {environmentCounts && environmentCounts.__none__ !== undefined ? <strong>{environmentCounts.__none__}</strong> : null}
        </button>
      </div>

      <form className={styles.inputRow} onSubmit={handleSubmit}>
        <input
          value={input}
          onChange={(event) => setInput(event.target.value)}
          placeholder={t('cloud.tags.filterPlaceholder')}
          aria-label={t('cloud.tags.filterPlaceholder')}
        />
        <button type="submit">{t('cloud.tags.addFilter')}</button>
        {customTokens.map((token) => (
          <button key={token} type="button" className={styles.activeToken} onClick={() => removeToken(token)}>
            {token} x
          </button>
        ))}
      </form>
    </div>
  );
}
