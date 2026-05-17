import { useState } from 'react';

import { useI18n } from '../../locale/useI18n';
import { TagChip } from './TagChip';
import styles from './TagEditor.module.css';

const TAG_KEY_REGEX = /^[a-z][a-z0-9_-]{0,62}$/;
const MAX_TAGS = 32;
const SUGGESTIONS = ['environment=prod', 'environment=dev', 'environment=qa', 'environment=orphan'];

interface TagEditorProps {
  initialTags: Record<string, string>;
  onSubmit: (tags: Record<string, string>) => Promise<void>;
  onCancel: () => void;
}

export function TagEditor({ initialTags, onSubmit, onCancel }: TagEditorProps) {
  const { t } = useI18n();
  const [tags, setTags] = useState<Record<string, string>>(initialTags);
  const [keyInput, setKeyInput] = useState('');
  const [valueInput, setValueInput] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  function addTag(rawKey: string, rawValue: string) {
    const key = rawKey.trim();
    const value = rawValue.trim();
    if (!TAG_KEY_REGEX.test(key)) {
      setError(t('cloud.tags.invalidKey'));
      return;
    }
    if (value.length === 0 || value.length > 128) {
      setError(t('cloud.tags.invalidValue'));
      return;
    }
    if (!(key in tags) && Object.keys(tags).length >= MAX_TAGS) {
      setError(t('cloud.tags.tooMany'));
      return;
    }
    setError(null);
    setTags((prev) => ({ ...prev, [key]: value }));
    setKeyInput('');
    setValueInput('');
  }

  function removeTag(key: string) {
    setTags((prev) => {
      const next = { ...prev };
      delete next[key];
      return next;
    });
  }

  function applySuggestion(suggestion: string) {
    const [key, value] = suggestion.split('=', 2);
    addTag(key, value);
  }

  async function handleSubmit() {
    setSubmitting(true);
    try {
      await onSubmit(tags);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.tags.saveFailed'));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className={styles.editor}>
      <div className={styles.tagList}>
        {Object.entries(tags).length === 0 ? (
          <span className={styles.empty}>{t('cloud.tags.empty')}</span>
        ) : (
          Object.entries(tags).map(([key, value]) => (
            <TagChip key={key} tagKey={key} value={value} onRemove={() => removeTag(key)} />
          ))
        )}
      </div>

      <div className={styles.addRow}>
        <input
          className={styles.keyInput}
          placeholder={t('cloud.tags.keyPlaceholder')}
          value={keyInput}
          onChange={(event) => setKeyInput(event.target.value)}
        />
        <input
          className={styles.valueInput}
          placeholder={t('cloud.tags.valuePlaceholder')}
          value={valueInput}
          onChange={(event) => setValueInput(event.target.value)}
        />
        <button type="button" onClick={() => addTag(keyInput, valueInput)} disabled={submitting}>
          {t('cloud.tags.add')}
        </button>
      </div>

      {error ? <p className={styles.error}>{error}</p> : null}

      <p className={styles.hint}>{t('cloud.tags.hint')}</p>
      <div className={styles.suggestions}>
        {SUGGESTIONS.map((suggestion) => (
          <button
            key={suggestion}
            type="button"
            className={styles.suggestion}
            onClick={() => applySuggestion(suggestion)}
          >
            {suggestion}
          </button>
        ))}
      </div>

      <div className={styles.actions}>
        <button type="button" className={styles.cancel} onClick={onCancel} disabled={submitting}>
          {t('cloud.tags.cancel')}
        </button>
        <button type="button" className={styles.save} onClick={() => void handleSubmit()} disabled={submitting}>
          {t('cloud.tags.save')}
        </button>
      </div>
    </div>
  );
}
