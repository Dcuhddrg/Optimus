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
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [showFallbackError, setShowFallbackError] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setShowFallbackError(false);
    setSubmitting(true);
    try {
      await login(username, password);
      const destination = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname || '/';
      navigate(destination, { replace: true });
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
        setShowFallbackError(false);
      } else {
        setError(null);
        setShowFallbackError(true);
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className={styles.page}>
      <section className={styles.shell}>
        <div className={styles.grain} aria-hidden="true" />

        <section className={styles.hero}>
          <div className={styles.heroTop}>
            <div className={styles.brandRow}>
              <div className={styles.brandMark}>O</div>
              <span className={styles.brandName}>{t('auth.brandName')}</span>
            </div>

            <div className={styles.heroCopy}>
              <p className={styles.eyebrow}>{t('auth.heroEyebrow')}</p>
              <h1 className={styles.heroTitle}>{t('auth.heroTitle')}</h1>
              <p className={styles.heroDescription}>{t('auth.heroDescription')}</p>
            </div>

            <div className={styles.heroGrid}>
              <article className={styles.heroCard}>
                <p className={styles.heroCardLabel}>{t('auth.heroAccessLabel')}</p>
                <p className={styles.heroCardText}>{t('auth.heroAccessText')}</p>
              </article>
              <article className={styles.heroCard}>
                <p className={styles.heroCardLabel}>{t('auth.heroControlLabel')}</p>
                <p className={styles.heroCardText}>{t('auth.heroControlText')}</p>
              </article>
              <article className={styles.heroCard}>
                <p className={styles.heroCardLabel}>{t('auth.heroAuditLabel')}</p>
                <p className={styles.heroCardText}>{t('auth.heroAuditText')}</p>
              </article>
            </div>
          </div>
        </section>

        <section className={styles.formPane}>
          <div className={styles.panel}>
            <div className={styles.panelHeader}>
              <LanguageSwitcher />
            </div>

            <div className={styles.formIntro}>
              <h2 className={styles.title}>{t('auth.signInTitle')}</h2>
              <p className={styles.copy}>{t('auth.signInDescription')}</p>
            </div>

            <form className={styles.form} onSubmit={handleSubmit}>
              <label className={styles.field}>
                <span>{t('auth.usernameLabel')}</span>
                <input
                  value={username}
                  placeholder="admin"
                  onChange={(event) => setUsername(event.target.value)}
                />
                <small className={styles.hint}>{t('auth.usernameExample')}</small>
              </label>
              <label className={styles.field}>
                <span>{t('auth.passwordLabel')}</span>
                <input
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                />
                <small className={styles.hint}>{t('auth.passwordHint')}</small>
              </label>
              {error || showFallbackError ? (
                <p className={styles.error}>
                  {showFallbackError ? t('auth.loginFailed') : error}
                </p>
              ) : null}
              <button className={styles.submit} type="submit" disabled={submitting}>
                {submitting ? t('auth.signingIn') : t('auth.signIn')}
              </button>
            </form>

            <div className={styles.meta}>
              <span>{t('auth.metaLocale')}</span>
              <span>{t('auth.metaAudit')}</span>
            </div>
          </div>
        </section>
      </section>
    </main>
  );
}
