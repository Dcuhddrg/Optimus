import { FormEvent, useEffect, useState } from 'react';

import { getLDAPConfig, saveLDAPConfig, testLDAPConfig } from '../../api/system';
import { PageHeader } from '../../components/page/PageHeader';
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
      setMessage('LDAP configuration saved');
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'Failed to save LDAP configuration');
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
      setMessage(err instanceof Error ? err.message : 'LDAP connection test failed');
    } finally {
      setTesting(false);
    }
  }

  return (
    <div className={styles.page}>
      <PageHeader
        eyebrow="System / LDAP"
        title="Directory access"
        description="Manage LDAP login configuration for non-admin users."
        actions={
          <button className={styles.primaryAction} type="button" onClick={() => void handleTest()}>
            {testing ? 'Testing...' : 'Test Connection'}
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
          <span>Enable LDAP login</span>
        </label>

        <div className={styles.formGrid}>
          <label className={styles.field}>
            <span>LDAP URL</span>
            <input value={config.url} onChange={(event) => update('url', event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>Base DN</span>
            <input value={config.baseDn} onChange={(event) => update('baseDn', event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>Bind DN</span>
            <input value={config.bindDn} onChange={(event) => update('bindDn', event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>Bind password</span>
            <input
              type="password"
              value={config.bindPassword || ''}
              placeholder={config.hasBindPassword ? 'Saved password retained' : ''}
              onChange={(event) => update('bindPassword', event.target.value)}
            />
          </label>
          <label className={styles.field}>
            <span>User filter</span>
            <input
              value={config.userFilter}
              onChange={(event) => update('userFilter', event.target.value)}
            />
          </label>
          <label className={styles.field}>
            <span>Username attribute</span>
            <input
              value={config.usernameAttr}
              onChange={(event) => update('usernameAttr', event.target.value)}
            />
          </label>
          <label className={styles.field}>
            <span>Display name attribute</span>
            <input
              value={config.displayNameAttr}
              onChange={(event) => update('displayNameAttr', event.target.value)}
            />
          </label>
          <label className={styles.field}>
            <span>Email attribute</span>
            <input value={config.emailAttr} onChange={(event) => update('emailAttr', event.target.value)} />
          </label>
        </div>

        {message ? <p className={styles.message}>{message}</p> : null}

        <div className={styles.actions}>
          <button className={styles.primaryAction} type="submit" disabled={saving}>
            {saving ? 'Saving...' : 'Save Configuration'}
          </button>
        </div>
      </form>
    </div>
  );
}
