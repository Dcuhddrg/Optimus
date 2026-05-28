import { PageHeader } from '../../components/page/PageHeader';
import { useI18n } from '../../locale/useI18n';

import styles from './DashboardPage.module.css';

export function DashboardPage() {
  const { t } = useI18n();
  const plannedModules = [
    t('dashboard.plannedUseStatusSummary'),
    t('dashboard.plannedUsePriorityContext'),
    t('dashboard.plannedUseOverviewModules'),
  ];

  return (
    <div className={styles.page}>
      <section className={styles.hero}>
        <div className={styles.heroCopy}>
          <PageHeader
            eyebrow={t('dashboard.eyebrow')}
            title={t('dashboard.title')}
            description={t('dashboard.description')}
          />
          <div className={styles.notice}>
            <span className={styles.noticeKicker}>{t('dashboard.emptyEyebrow')}</span>
            <p>{t('dashboard.emptyTitle')}</p>
          </div>
        </div>

        <div className={styles.commandSurface} aria-hidden="true">
          <div className={styles.commandTopline}>
            <span />
            <span />
            <span />
          </div>
          <div className={styles.commandGrid}>
            <span />
            <span />
            <span />
            <span />
            <span />
            <span />
          </div>
          <div className={styles.commandStack}>
            <span />
            <span />
            <span />
          </div>
        </div>
      </section>

      <section className={styles.overviewPanel}>
        <div className={styles.panelIntro}>
          <p className={styles.panelEyebrow}>{t('dashboard.emptyEyebrow')}</p>
          <h2>{t('dashboard.emptyTitle')}</h2>
          <p>{t('dashboard.emptyDescription')}</p>
        </div>

        <div className={styles.moduleRail}>
          <p className={styles.railLabel}>{t('dashboard.plannedUseLabel')}</p>
          <div className={styles.moduleList}>
            {plannedModules.map((moduleName, index) => (
              <article className={styles.moduleItem} key={moduleName}>
                <span className={styles.moduleIndex}>{String(index + 1).padStart(2, '0')}</span>
                <span className={styles.moduleName}>{moduleName}</span>
                <span className={styles.moduleLine} />
              </article>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
