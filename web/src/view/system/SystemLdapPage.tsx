import { FormEvent, useEffect, useState } from 'react';

import { getLDAPConfig, saveLDAPConfig, testLDAPConfig } from '../../api/system';
import { PageHeader } from '../../components/page/PageHeader';
import { useI18n } from '../../locale/useI18n';
import type { LDAPConfig } from '../../types/system';
import styles from './SystemLdapPage.module.css';

const emptyConfig: LDAPConfig = {
  enabled: false,
  url: '',
  baseDn: '',
  bindDn: '',
  bindPassword: '',
  hasBindPassword: false,
  userFilter: '(uid={username})',
  usernameAttr: 'uid',
  displayNameAttr: 'cn',
  emailAttr: 'mail',
};

export function SystemLdapPage() {
  const { t } = useI18n();
  const [config, setConfig] = useState<LDAPConfig>(emptyConfig);
  const [message, setMessage] = useState('');
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);

  useEffect(() => {
    void getLDAPConfig().then((result) => setConfig({ ...emptyConfig, ...result, bindPassword: '' }));
  }, []);

  function update<K extends keyof LDAPConfig>(key: K, value: LDAPConfig[K]) {
    setConfig((current) => ({ ...current, [key]: value }));
  }

  async function handleSave(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setMessage('');
    try {
      const saved = await saveLDAPConfig(config);
      setConfig({ ...emptyConfig, ...saved, bindPassword: '' });
      setMessage(t('system.ldap.saveSuccess'));
    } catch (err) {
      setMessage(err instanceof Error ? err.message : t('system.ldap.saveFailed'));
    } finally {
      setSaving(false);
    }
  }

  async function handleTest() {
    setTesting(true);
    setMessage('');
    try {
      const result = await testLDAPConfig({
        ...config,
        useSavedBindPassword: !config.bindPassword && config.hasBindPassword,
      });
      setMessage(result.message);
    } catch (err) {
      setMessage(err instanceof Error ? err.message : t('system.ldap.testFailed'));
    } finally {
      setTesting(false);
    }
  }

  return (
    <div className={styles.page}>
      <PageHeader
        eyebrow={t('system.ldap.eyebrow')}
        title={t('system.ldap.title')}
        description={t('system.ldap.description')}
        actions={
          <button className={styles.primaryAction} type="button" onClick={() => void handleTest()}>
            {testing ? t('system.ldap.testing') : t('system.ldap.testConnection')}
          </button>
        }
      />

      <form className={styles.panel} onSubmit={handleSave}>
        <label className={styles.toggle}>
          <input
            type="checkbox"
            checked={config.enabled}
            onChange={(event) => update('enabled', event.target.checked)}
          />
          <span>{t('system.ldap.enableLogin')}</span>
        </label>

        <div className={styles.formGrid}>
          <label className={styles.field}>
            <span>{t('system.ldap.urlLabel')}</span>
            <input value={config.url} onChange={(event) => update('url', event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>{t('system.ldap.baseDnLabel')}</span>
            <input value={config.baseDn} onChange={(event) => update('baseDn', event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>{t('system.ldap.bindDnLabel')}</span>
            <input value={config.bindDn} onChange={(event) => update('bindDn', event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>{t('system.ldap.bindPasswordLabel')}</span>
            <input
              type="password"
              value={config.bindPassword || ''}
              placeholder={config.hasBindPassword ? t('system.ldap.savedPasswordRetained') : ''}
              onChange={(event) => update('bindPassword', event.target.value)}
            />
          </label>
          <label className={styles.field}>
            <span>{t('system.ldap.userFilterLabel')}</span>
            <input
              value={config.userFilter}
              onChange={(event) => update('userFilter', event.target.value)}
            />
          </label>
          <label className={styles.field}>
            <span>{t('system.ldap.usernameAttributeLabel')}</span>
            <input
              value={config.usernameAttr}
              onChange={(event) => update('usernameAttr', event.target.value)}
            />
          </label>
          <label className={styles.field}>
            <span>{t('system.ldap.displayNameAttributeLabel')}</span>
            <input
              value={config.displayNameAttr}
              onChange={(event) => update('displayNameAttr', event.target.value)}
            />
          </label>
          <label className={styles.field}>
            <span>{t('system.ldap.emailAttributeLabel')}</span>
            <input value={config.emailAttr} onChange={(event) => update('emailAttr', event.target.value)} />
          </label>
        </div>

        {message ? <p className={styles.message}>{message}</p> : null}

        <div className={styles.actions}>
          <button className={styles.primaryAction} type="submit" disabled={saving}>
            {saving ? t('system.ldap.saving') : t('system.ldap.saveConfiguration')}
          </button>
        </div>
      </form>
    </div>
  );
}
