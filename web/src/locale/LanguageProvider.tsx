import { createContext, type ReactNode, useEffect, useState } from 'react';

import { resolveLocale } from './resolveLocale';
import { getStoredLocale, setStoredLocale } from './storage';
import type { AppLocale, TranslationKey } from './types';

export interface I18nContextValue {
  locale: AppLocale;
  setLocale: (locale: AppLocale) => void;
  t: (key: TranslationKey) => string;
}

export const I18nContext = createContext<I18nContextValue | null>(null);

const PLACEHOLDER_MESSAGES: Record<AppLocale, Record<TranslationKey, string>> = {
  'zh-CN': {
    'common.locale.english': 'English',
    'common.locale.simplifiedChinese': '简体中文',
  },
  en: {
    'common.locale.english': 'English',
    'common.locale.simplifiedChinese': 'Simplified Chinese',
  },
};

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState<AppLocale>('en');
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    const initialLocale =
      getStoredLocale() ?? resolveLocale(window.navigator.language);

    setLocale(initialLocale);
    setHydrated(true);
  }, []);

  useEffect(() => {
    if (!hydrated) {
      return;
    }

    setStoredLocale(locale);
  }, [hydrated, locale]);

  function t(key: TranslationKey) {
    return PLACEHOLDER_MESSAGES[locale][key] ?? PLACEHOLDER_MESSAGES.en[key] ?? key;
  }

  return (
    <I18nContext.Provider value={{ locale, setLocale, t }}>
      {children}
    </I18nContext.Provider>
  );
}
