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
  const props = {
    width: 20,
    height: 20,
    viewBox: '0 0 24 24',
    fill: 'none' as const,
    stroke: 'currentColor',
    strokeWidth: 1.5,
    strokeLinecap: 'round' as const,
    strokeLinejoin: 'round' as const,
  };
  switch (token) {
    case 'shield':
      return (
        <svg {...props}>
          <path d="M12 3 4 6v6c0 4.5 3.4 8.4 8 9 4.6-.6 8-4.5 8-9V6l-8-3Z" />
          <path d="m9 12 2 2 4-4" />
        </svg>
      );
    case 'cloud':
      return (
        <svg {...props}>
          <path d="M7 17a5 5 0 0 1 0-10 6 6 0 0 1 11.4 1A4 4 0 0 1 18 17H7Z" />
        </svg>
      );
    case 'network':
      return (
        <svg {...props}>
          <circle cx="6" cy="12" r="2.5" />
          <circle cx="18" cy="6" r="2.5" />
          <circle cx="18" cy="18" r="2.5" />
          <path d="m8 11 8-4M8 13l8 4" />
        </svg>
      );
    case 'env':
      return (
        <svg {...props}>
          <path d="m12 3 9 5-9 5-9-5 9-5Z" />
          <path d="m3 13 9 5 9-5" />
          <path d="m3 17 9 5 9-5" />
        </svg>
      );
    case 'support':
      return (
        <svg {...props}>
          <path d="M4 13v-1a8 8 0 0 1 16 0v1" />
          <rect x="3" y="13" width="4" height="6" rx="1.5" />
          <rect x="17" y="13" width="4" height="6" rx="1.5" />
          <path d="M20 19a4 4 0 0 1-4 4h-2" />
        </svg>
      );
    default:
      return null;
  }
}
