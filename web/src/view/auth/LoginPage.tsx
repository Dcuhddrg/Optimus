import { FormEvent, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import { LanguageSwitcher } from '../../components/locale/LanguageSwitcher';
import { useI18n } from '../../locale/useI18n';
import styles from './LoginPage.module.css';

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuth();
  const { t } = useI18n();
  const [username, setUsername] = useState('admin');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      await login(username, password);
      const destination = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname || '/';
      navigate(destination, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : t('auth.loginFailed'));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className={styles.page}>
      <section className={styles.panel}>
        <div className={styles.panelHeader}>
          <LanguageSwitcher />
        </div>
        <div>
          <p className={styles.eyebrow}>{t('auth.eyebrow')}</p>
          <h1 className={styles.title}>{t('auth.signInTitle')}</h1>
          <p className={styles.copy}>{t('auth.signInDescription')}</p>
        </div>

        <form className={styles.form} onSubmit={handleSubmit}>
          <label className={styles.field}>
            <span>{t('auth.usernameLabel')}</span>
            <input value={username} onChange={(event) => setUsername(event.target.value)} />
          </label>
          <label className={styles.field}>
            <span>{t('auth.passwordLabel')}</span>
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
          </label>
          {error ? <p className={styles.error}>{error}</p> : null}
          <button className={styles.submit} type="submit" disabled={submitting}>
            {submitting ? t('auth.signingIn') : t('auth.signIn')}
          </button>
        </form>
      </section>
    </main>
  );
}
