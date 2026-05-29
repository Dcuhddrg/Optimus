// RequestForm.tsx
import { useState, type FormEvent } from 'react';

import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import { useTicketStore } from '../data/ticketStore';
import { ticketMockEngine } from '../data/ticketMockEngine';
import type { RequestFormField, StorageStatus } from '../ticketTypes';
import styles from './RequestForm.module.css';

interface Props {
  serviceId: string | null;
  serviceTitle: string;
  schema: RequestFormField[];
  applicant: { name: string; email: string };
  onSubmitted: () => void;
  onCancel: () => void;
}

type Values = Record<string, string | string[]>;

export function RequestForm({ serviceId, serviceTitle, schema, applicant, onSubmitted, onCancel }: Props) {
  const { t } = useI18n();
  const { createTicket } = useTicketStore();
  const [values, setValues] = useState<Values>(() => initValues(schema));
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [storageBanner, setStorageBanner] = useState<StorageStatus | null>(null);

  function setField(name: string, value: string | string[]) {
    setValues((prev) => ({ ...prev, [name]: value }));
  }

  function validate(): boolean {
    const next: Record<string, string> = {};
    for (const field of schema) {
      const raw = values[field.name];
      const empty =
        raw == null ||
        (typeof raw === 'string' && raw.trim() === '') ||
        (Array.isArray(raw) && raw.length === 0);
      if (field.required && empty) {
        next[field.name] = t('servicePortal.validationRequired');
        continue;
      }
      if (typeof raw === 'string' && field.maxLength && raw.length > field.maxLength) {
        next[field.name] = t('servicePortal.validationTooLong').replace('{max}', String(field.maxLength));
      }
    }
    setErrors(next);
    return Object.keys(next).length === 0;
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (submitting) return;
    if (!validate()) return;
    setSubmitting(true);
    const title = deriveTitle(serviceId, serviceTitle, values);
    const result = createTicket({ serviceId, title, formValues: values, applicant });
    ticketMockEngine.schedule(result.ticket.id);
    if (result.storageStatus === 'memory-only') {
      setStorageBanner('memory-only');
      // 表单保留已提交态，按钮 disabled 不变；不关 Drawer
      return;
    }
    onSubmitted();
  }

  return (
    <form className={styles.form} onSubmit={handleSubmit}>
      {storageBanner === 'memory-only' ? (
        <div className={styles.banner} role="alert">
          {t('servicePortal.storageFailed')}
        </div>
      ) : null}

      {schema.map((field) => (
        <div className={styles.field} key={field.name}>
          <label className={styles.label} htmlFor={`field-${field.name}`}>
            <span>{t(field.i18nLabelKey as MessageKey)}</span>
            {field.required ? null : <span className={styles.optional}>{t('servicePortal.optional')}</span>}
          </label>
          {renderField(field, values[field.name] ?? '', setField, t, `field-${field.name}`)}
          {errors[field.name] ? <span className={styles.error}>{errors[field.name]}</span> : null}
        </div>
      ))}

      <div className={styles.actions}>
        <button type="button" className={styles.cancel} onClick={onCancel} disabled={submitting}>
          {t('servicePortal.cancel')}
        </button>
        <button type="submit" className={styles.submit} disabled={submitting}>
          {submitting ? t('servicePortal.submitting') : t('servicePortal.submit')}
        </button>
      </div>
    </form>
  );
}

function initValues(schema: RequestFormField[]): Values {
  const result: Values = {};
  for (const field of schema) {
    result[field.name] = field.type === 'multiselect' ? [] : '';
  }
  return result;
}

function deriveTitle(serviceId: string | null, serviceTitle: string, values: Values): string {
  if (serviceId) return serviceTitle;
  const raw = values.title;
  if (typeof raw === 'string' && raw.trim()) return raw.trim();
  return serviceTitle; // fallback
}

function renderField(
  field: RequestFormField,
  value: string | string[],
  setField: (name: string, value: string | string[]) => void,
  t: (key: MessageKey) => string,
  inputId: string,
) {
  const placeholder = field.placeholderI18nKey ? t(field.placeholderI18nKey as MessageKey) : undefined;
  switch (field.type) {
    case 'textarea':
      return (
        <textarea
          id={inputId}
          className={styles.textarea}
          value={typeof value === 'string' ? value : ''}
          maxLength={field.maxLength}
          placeholder={placeholder}
          onChange={(e) => setField(field.name, e.target.value)}
        />
      );
    case 'select':
      return (
        <select
          id={inputId}
          className={styles.select}
          value={typeof value === 'string' ? value : ''}
          onChange={(e) => setField(field.name, e.target.value)}
        >
          <option value="">{placeholder ?? ''}</option>
          {field.options?.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {t(opt.i18nLabelKey as MessageKey)}
            </option>
          ))}
        </select>
      );
    case 'multiselect':
      return (
        <select
          id={inputId}
          multiple
          className={styles.select}
          value={Array.isArray(value) ? value : []}
          onChange={(e) => {
            const next = Array.from(e.target.selectedOptions).map((o) => o.value);
            setField(field.name, next);
          }}
        >
          {field.options?.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {t(opt.i18nLabelKey as MessageKey)}
            </option>
          ))}
        </select>
      );
    case 'date':
      return (
        <input
          id={inputId}
          type="date"
          className={styles.input}
          value={typeof value === 'string' ? value : ''}
          onChange={(e) => setField(field.name, e.target.value)}
        />
      );
    case 'text':
    default:
      return (
        <input
          id={inputId}
          type="text"
          className={styles.input}
          value={typeof value === 'string' ? value : ''}
          maxLength={field.maxLength}
          placeholder={placeholder}
          onChange={(e) => setField(field.name, e.target.value)}
        />
      );
  }
}
