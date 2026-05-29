// CategoryTabs.tsx
import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { ServiceCategory } from '../ticketTypes';
import { CATEGORY_I18N_KEYS } from '../data/serviceCatalog';
import styles from './CategoryTabs.module.css';

export type CategoryFilter = ServiceCategory | 'popular';

interface Props {
  active: CategoryFilter;
  onChange: (next: CategoryFilter) => void;
}

const ORDER: CategoryFilter[] = ['popular', 'account', 'cloud', 'network', 'env', 'support'];

export function CategoryTabs({ active, onChange }: Props) {
  const { t } = useI18n();
  return (
    <div className={styles.tabs} role="tablist" aria-label={t('servicePortal.category.label')}>
      {ORDER.map((category) => {
        const isActive = category === active;
        return (
          <button
            key={category}
            type="button"
            role="tab"
            aria-selected={isActive}
            className={isActive ? `${styles.tab} ${styles.tabActive}` : styles.tab}
            onClick={() => onChange(category)}
          >
            {t(CATEGORY_I18N_KEYS[category] as MessageKey)}
          </button>
        );
      })}
    </div>
  );
}
