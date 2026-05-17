import styles from './TagChip.module.css';

interface TagChipProps {
  tagKey: string;
  value: string;
  onRemove?: () => void;
}

const ENV_VALUES: Record<string, string> = {
  prod: styles.envProd,
  dev: styles.envDev,
  qa: styles.envQa,
  orphan: styles.envOrphan,
};

export function TagChip({ tagKey, value, onRemove }: TagChipProps) {
  const envClass = tagKey === 'environment' ? ENV_VALUES[value] : undefined;
  return (
    <span className={[styles.chip, envClass].filter(Boolean).join(' ')}>
      <span className={styles.key}>{tagKey}</span>
      <span className={styles.eq}>=</span>
      <span>{value}</span>
      {onRemove ? (
        <button type="button" className={styles.remove} onClick={onRemove} aria-label="remove tag">
          x
        </button>
      ) : null}
    </span>
  );
}
