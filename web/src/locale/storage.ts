import { LOCALE_STORAGE_KEY, SUPPORTED_LOCALES, type AppLocale } from './types';

function isSupportedLocale(value: string | null): value is AppLocale {
  return SUPPORTED_LOCALES.includes(value as AppLocale);
}

export function getStoredLocale(): AppLocale | null {
  try {
    const storedLocale = window.localStorage.getItem(LOCALE_STORAGE_KEY);
    return isSupportedLocale(storedLocale) ? storedLocale : null;
  } catch {
    return null;
  }
}

export function setStoredLocale(locale: AppLocale): void {
  try {
    window.localStorage.setItem(LOCALE_STORAGE_KEY, locale);
  } catch {
    // Ignore storage failures so restricted browser contexts still render.
  }
}
