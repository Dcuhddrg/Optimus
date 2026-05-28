import { PageHeader } from '../../components/page/PageHeader';
import { useI18n } from '../../locale/useI18n';

import styles from './ServicePortalPage.module.css';

export function ServicePortalPage() {
  const { t } = useI18n();
  const categories = [
    t('servicePortal.category.popular'),
    t('servicePortal.category.account'),
    t('servicePortal.category.cloud'),
    t('servicePortal.category.network'),
    t('servicePortal.category.support'),
  ];
  const services = [
    {
      title: t('servicePortal.catalog.accountRoleGrant.title'),
      description: t('servicePortal.catalog.accountRoleGrant.description'),
      tags: [t('servicePortal.tag.approval'), t('servicePortal.tag.oneToTwoDays')],
    },
    {
      title: t('servicePortal.catalog.accountRoleGrant.title'),
      description: t('servicePortal.catalog.accountRoleGrant.description'),
      tags: [t('servicePortal.tag.projectScope'), t('servicePortal.tag.costReview')],
    },
    {
      title: t('servicePortal.catalog.accountRoleGrant.title'),
      description: t('servicePortal.catalog.accountRoleGrant.description'),
      tags: [t('servicePortal.tag.securityReview'), t('servicePortal.tag.frequent')],
    },
    {
      title: t('servicePortal.catalog.accountRoleGrant.title'),
      description: t('servicePortal.catalog.accountRoleGrant.description'),
      tags: [t('servicePortal.tag.fastResponse'), t('servicePortal.tag.trackable')],
    },
  ];
  const tickets = [
    {
      title: t('servicePortal.catalog.accountRoleGrant.title'),
      meta: t('servicePortal.catalog.accountRoleGrant.description'),
      status: t('servicePortal.ticket.status.approving'),
    },
    {
      title: t('servicePortal.catalog.accountRoleGrant.title'),
      meta: t('servicePortal.catalog.accountRoleGrant.description'),
      status: t('servicePortal.ticket.status.approving'),
      tone: 'warning',
    },
    {
      title: t('servicePortal.catalog.accountRoleGrant.title'),
      meta: t('servicePortal.catalog.accountRoleGrant.description'),
      status: t('servicePortal.ticket.status.processing'),
    },
  ];

  return (
    <div className={styles.page}>
      <PageHeader
        eyebrow={t('servicePortal.eyebrow')}
        title={t('servicePortal.title')}
        description={t('servicePortal.description')}
      />

      <section className={styles.portalPanel}>
        <div className={styles.panelHeader}>
          <div>
            <h2>{t('servicePortal.title')}</h2>
            <p>{t('servicePortal.description')}</p>
          </div>
          <div className={styles.actions}>
            <button className={styles.secondaryAction} type="button">
              {t('servicePortal.viewAll')}
            </button>
            <button className={styles.primaryAction} type="button">
              {t('servicePortal.createRequest')}
            </button>
          </div>
        </div>

        <div className={styles.panelBody}>
          <div className={styles.catalogArea}>
            <div className={styles.categoryTabs} aria-label={t('servicePortal.category.label')}>
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
            <h3>{t('servicePortal.ticket.myAside.title')}</h3>
            <p>{t('servicePortal.ticket.myAside.description')}</p>
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
