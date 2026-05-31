// ServiceCard.tsx
import { ShieldCheck, Cloud, ShareNetwork, StackSimple, Headset } from '@phosphor-icons/react';

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
    <button type="button" className={item.hot ? `${styles.card} ${styles.cardWide}` : styles.card} onClick={() => onSelect(item)}>
      <span className={styles.icon} aria-hidden="true">
        <ServiceTokenIcon token={item.iconToken} />
      </span>
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

function ServiceTokenIcon({ token }: { token: ServiceCatalogItem['iconToken'] }) {
  const size = 20;
  const weight = 'duotone' as const;
  switch (token) {
    case 'shield':
      return <ShieldCheck size={size} weight={weight} />;
    case 'cloud':
      return <Cloud size={size} weight={weight} />;
    case 'network':
      return <ShareNetwork size={size} weight={weight} />;
    case 'env':
      return <StackSimple size={size} weight={weight} />;
    case 'support':
      return <Headset size={size} weight={weight} />;
    default:
      return null;
  }
}
