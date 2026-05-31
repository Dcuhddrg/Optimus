// ServicePortalPage.tsx
import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { AnimatePresence, motion } from 'framer-motion';

import { useAuth } from '../../auth/AuthContext';
import { useI18n } from '../../locale/useI18n';
import type { MessageKey } from '../../locale/messages';
import { PageHeader } from '../../components/page/PageHeader';
import { CategoryTabs, type CategoryFilter } from './components/CategoryTabs';
import { MyTicketsAside } from './components/MyTicketsAside';
import { RequestDrawer } from './components/RequestDrawer';
import { RequestForm } from './components/RequestForm';
import { ServiceCard } from './components/ServiceCard';
import { listByCategory } from './data/serviceCatalog';
import { getFormSchema } from './data/serviceFormSchemas';
import type { ServiceCatalogItem } from './ticketTypes';
import styles from './ServicePortalPage.module.css';

type DrawerSelection =
  | { kind: 'service'; item: ServiceCatalogItem }
  | { kind: 'generic' }
  | null;

export function ServicePortalPage() {
  const { t } = useI18n();
  const { current } = useAuth();
  const [activeCategory, setActiveCategory] = useState<CategoryFilter>('popular');
  const [drawer, setDrawer] = useState<DrawerSelection>(null);

  const visibleItems = useMemo(() => listByCategory(activeCategory), [activeCategory]);

  const applicant = {
    name: current?.user.displayName ?? current?.user.username ?? '',
    email: current?.user.email ?? '',
  };

  function openService(item: ServiceCatalogItem) {
    setDrawer({ kind: 'service', item });
  }

  function openGeneric() {
    setDrawer({ kind: 'generic' });
  }

  function closeDrawer() {
    setDrawer(null);
  }

  const drawerProps = useMemo(() => {
    if (!drawer) return null;
    if (drawer.kind === 'service') {
      const titleKey = `${drawer.item.i18nKey}.title` as MessageKey;
      const descKey = `${drawer.item.i18nKey}.description` as MessageKey;
      return {
        serviceId: drawer.item.id,
        serviceTitle: t(titleKey),
        description: t(descKey),
        schema: getFormSchema(drawer.item.id),
      };
    }
    return {
      serviceId: null,
      serviceTitle: t('servicePortal.createRequest'),
      description: undefined as string | undefined,
      schema: getFormSchema(null),
    };
  }, [drawer, t]);

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
            <Link className={styles.secondaryAction} to="/service-portal/tickets">
              {t('servicePortal.viewAll')}
            </Link>
            <button className={styles.primaryAction} type="button" onClick={openGeneric}>
              {t('servicePortal.createRequest')}
            </button>
          </div>
        </div>

        <div className={styles.panelBody}>
          <div className={styles.catalogArea}>
            <CategoryTabs active={activeCategory} onChange={setActiveCategory} />
            <AnimatePresence mode="wait" initial={false}>
              <motion.div
                key={activeCategory}
                className={styles.serviceGrid}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                transition={{ duration: 0.22 }}
              >
                {visibleItems.map((item) => (
                  <ServiceCard key={item.id} item={item} onSelect={openService} />
                ))}
              </motion.div>
            </AnimatePresence>
          </div>

          <MyTicketsAside />
        </div>
      </section>

      <RequestDrawer
        open={drawer !== null}
        onClose={closeDrawer}
        eyebrow={t('servicePortal.eyebrow')}
        title={drawerProps?.serviceTitle ?? ''}
        description={drawerProps?.description}
      >
        {drawerProps ? (
          <RequestForm
            serviceId={drawerProps.serviceId}
            serviceTitle={drawerProps.serviceTitle}
            schema={drawerProps.schema}
            applicant={applicant}
            onSubmitted={closeDrawer}
            onCancel={closeDrawer}
          />
        ) : null}
      </RequestDrawer>
    </div>
  );
}
