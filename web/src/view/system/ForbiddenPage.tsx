import { useI18n } from '../../locale/useI18n';

export function ForbiddenPage() {
  const { t } = useI18n();

  return (
    <section className="page-card">
      <p className="page-eyebrow">{t('system.forbidden.eyebrow')}</p>
      <h1 className="page-title">{t('system.forbidden.title')}</h1>
      <p className="page-description">{t('system.forbidden.description')}</p>
    </section>
  );
}
