import { PageHeader } from '../../components/page/PageHeader';
import { useI18n } from '../../locale/useI18n';

import styles from './ServicePortalPage.module.css';

export function ServicePortalPage() {
  const { t } = useI18n();
  const categories = [
    t('dashboard.servicePortal.categoryPopular'),
    t('dashboard.servicePortal.categoryAccount'),
    t('dashboard.servicePortal.categoryCloud'),
    t('dashboard.servicePortal.categoryNetwork'),
    t('dashboard.servicePortal.categorySupport'),
  ];
  const services = [
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
  const tickets = [
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
      <PageHeader
        eyebrow={t('dashboard.servicePortal.eyebrow')}
        title={t('dashboard.servicePortal.title')}
        description={t('dashboard.servicePortal.description')}
      />

      <section className={styles.portalPanel}>
        <div className={styles.panelHeader}>
          <div>
            <h2>{t('dashboard.servicePortal.title')}</h2>
            <p>{t('dashboard.servicePortal.description')}</p>
          </div>
          <div className={styles.actions}>
            <button className={styles.secondaryAction} type="button">
              {t('dashboard.servicePortal.viewAll')}
            </button>
            <button className={styles.primaryAction} type="button">
              {t('dashboard.servicePortal.createRequest')}
            </button>
          </div>
        </div>

        <div className={styles.panelBody}>
          <div className={styles.catalogArea}>
            <div className={styles.categoryTabs} aria-label={t('dashboard.servicePortal.categoryLabel')}>
              {categories.map((category, index) => (
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
              {services.map((service) => (
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
              {tickets.map((ticket) => (
                <article className={styles.ticketItem} key={ticket.title}>
                  <div className={styles.ticketTopline}>
                    <strong>{ticket.title}</strong>
                    <span
                      className={
                        ticket.tone === 'warning'
                          ? `${styles.ticketStatus} ${styles.ticketStatusWarning}`
                          : styles.ticketStatus
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
