// Formatting helpers shared by pages, as gin-vue-admin's web/src/utils/format.

const DATE_TIME_OPTIONS: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
  hour12: false,
};

/**
 * Formats a timestamp in the app locale with a 24-hour clock. Empty values
 * return `fallback`; values that do not parse as a date are returned as-is.
 */
export function formatDateTime(value: string | number | Date | null | undefined, locale?: string, fallback = '-'): string {
  if (value === null || value === undefined || value === '') {
    return fallback;
  }
  const date = value instanceof Date ? value : new Date(value);
  if (!Number.isFinite(date.getTime())) {
    return String(value);
  }
  return date.toLocaleString(locale, DATE_TIME_OPTIONS);
}
