// ServiceCard.tsx
import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { ServiceCatalogItem } from '../ticketTypes';
import styles from './ServiceCard.module.css';

interface Props {
  item: ServiceCatalogItem;
  onSelect: (item: ServiceCatalogItem) => void;
}

export function ServiceCard({ item, onSelect }: Props) {
  const { t } = useI18n();
  const titleKey = `${item.i18nKey}.title` as MessageKey;
  const descKey = `${item.i18nKey}.description` as MessageKey;
  return (
    <button type="button" className={styles.card} onClick={() => onSelect(item)}>
      <span className={styles.icon} aria-hidden="true" />
      <h3 className={styles.title}>{t(titleKey)}</h3>
      <p className={styles.description}>{t(descKey)}</p>
      <div className={styles.tags}>
        {item.tags.map((tag) => (
          <span key={tag} className={styles.tag}>
            {t(tag as MessageKey)}
          </span>
        ))}
      </div>
    </button>
  );
}
