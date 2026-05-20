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
  const serviceCategories = [
    t('dashboard.servicePortal.categoryPopular'),
    t('dashboard.servicePortal.categoryAccount'),
    t('dashboard.servicePortal.categoryCloud'),
    t('dashboard.servicePortal.categoryNetwork'),
    t('dashboard.servicePortal.categorySupport'),
  ];
  const serviceCards = [
    {
      title: t('dashboard.servicePortal.accountTitle'),
      description: t('dashboard.servicePortal.accountDescription'),
      tags: [t('dashboard.servicePortal.tagApproval'), t('dashboard.servicePortal.tagOneToTwoDays')],
    },
    {
      title: t('dashboard.servicePortal.cloudTitle'),
      description: t('dashboard.servicePortal.cloudDescription'),
      tags: [t('dashboard.servicePortal.tagProjectScope'), t('dashboard.servicePortal.tagCostReview')],
    },
    {
      title: t('dashboard.servicePortal.networkTitle'),
      description: t('dashboard.servicePortal.networkDescription'),
      tags: [t('dashboard.servicePortal.tagSecurityReview'), t('dashboard.servicePortal.tagFrequent')],
    },
    {
      title: t('dashboard.servicePortal.supportTitle'),
      description: t('dashboard.servicePortal.supportDescription'),
      tags: [t('dashboard.servicePortal.tagFastResponse'), t('dashboard.servicePortal.tagTrackable')],
    },
  ];
  const serviceTickets = [
    {
      title: t('dashboard.servicePortal.ticketProjectAccess'),
      meta: t('dashboard.servicePortal.ticketProjectAccessMeta'),
      status: t('dashboard.servicePortal.statusApproving'),
    },
    {
      title: t('dashboard.servicePortal.ticketVpn'),
      meta: t('dashboard.servicePortal.ticketVpnMeta'),
      status: t('dashboard.servicePortal.statusNeedInput'),
      tone: 'warning',
    },
    {
      title: t('dashboard.servicePortal.ticketCloudResize'),
      meta: t('dashboard.servicePortal.ticketCloudResizeMeta'),
      status: t('dashboard.servicePortal.statusProcessing'),
    },
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

      <section className={styles.servicePortal}>
        <div className={styles.serviceHeader}>
          <div>
            <p className={styles.serviceEyebrow}>{t('dashboard.servicePortal.eyebrow')}</p>
            <h2>{t('dashboard.servicePortal.title')}</h2>
            <p>{t('dashboard.servicePortal.description')}</p>
          </div>
          <div className={styles.serviceActions}>
            <button className={styles.secondaryAction} type="button">
              {t('dashboard.servicePortal.viewAll')}
            </button>
            <button className={styles.primaryAction} type="button">
              {t('dashboard.servicePortal.createRequest')}
            </button>
          </div>
        </div>

        <div className={styles.serviceBody}>
          <div className={styles.catalogArea}>
            <div className={styles.categoryTabs} aria-label={t('dashboard.servicePortal.categoryLabel')}>
              {serviceCategories.map((category, index) => (
                <button
                  className={index === 0 ? `${styles.categoryTab} ${styles.categoryTabActive}` : styles.categoryTab}
                  type="button"
                  key={category}
                >
                  {category}
                </button>
              ))}
            </div>

            <div className={styles.serviceGrid}>
              {serviceCards.map((service) => (
                <article className={styles.serviceCard} key={service.title}>
                  <span className={styles.serviceIcon} aria-hidden="true" />
                  <h3>{service.title}</h3>
                  <p>{service.description}</p>
                  <div className={styles.serviceTags}>
                    {service.tags.map((tag) => (
                      <span className={styles.serviceTag} key={tag}>
                        {tag}
                      </span>
                    ))}
                  </div>
                </article>
              ))}
            </div>
          </div>

          <aside className={styles.ticketArea}>
            <h3>{t('dashboard.servicePortal.myTickets')}</h3>
            <p>{t('dashboard.servicePortal.myTicketsDescription')}</p>
            <div className={styles.ticketList}>
              {serviceTickets.map((ticket) => (
                <article className={styles.ticketItem} key={ticket.title}>
                  <div className={styles.ticketTopline}>
                    <strong>{ticket.title}</strong>
                    <span
                      className={
                        ticket.tone === 'warning' ? `${styles.ticketStatus} ${styles.ticketStatusWarning}` : styles.ticketStatus
                      }
                    >
                      {ticket.status}
                    </span>
                  </div>
                  <p>{ticket.meta}</p>
                </article>
              ))}
            </div>
          </aside>
        </div>
      </section>
    </div>
  );
}
