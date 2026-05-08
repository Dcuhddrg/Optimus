import { PageHeader } from '../../components/page/PageHeader';
import { EmptyState } from '../../components/state/EmptyState';
import { useI18n } from '../../locale/useI18n';

import styles from './DashboardPage.module.css';

export function DashboardPage() {
  const { t } = useI18n();

  return (
    <div className={styles.page}>
      <PageHeader
        eyebrow={t('dashboard.eyebrow')}
        title={t('dashboard.title')}
        description={t('dashboard.description')}
      />

      <EmptyState
        eyebrow={t('dashboard.emptyEyebrow')}
        title={t('dashboard.emptyTitle')}
        description={t('dashboard.emptyDescription')}
        aside={
          <div className={styles.aside}>
            <p className={styles.asideLabel}>{t('dashboard.plannedUseLabel')}</p>
            <ul className={styles.asideList}>
              <li>{t('dashboard.plannedUseStatusSummary')}</li>
              <li>{t('dashboard.plannedUsePriorityContext')}</li>
              <li>{t('dashboard.plannedUseOverviewModules')}</li>
            </ul>
          </div>
        }
      />
    </div>
  );
}
