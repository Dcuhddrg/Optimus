import { useI18n } from '../../locale/useI18n';
import styles from './LanguageSwitcher.module.css';

export function LanguageSwitcher() {
  const { locale, setLocale, t } = useI18n();

  return (
    <div
      className={styles.switcher}
      role="group"
      aria-label={t('common.locale.switcherLabel')}
    >
      <button
        type="button"
        className={locale === 'zh-CN' ? styles.optionActive : styles.option}
        onClick={() => setLocale('zh-CN')}
        aria-pressed={locale === 'zh-CN'}
      >
        {t('common.locale.simplifiedChinese')}
      </button>
      <button
        type="button"
        className={locale === 'en' ? styles.optionActive : styles.option}
        onClick={() => setLocale('en')}
        aria-pressed={locale === 'en'}
      >
        {t('common.locale.english')}
      </button>
    </div>
  );
}
