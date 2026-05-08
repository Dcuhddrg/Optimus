import { createContext, type ReactNode, useState } from 'react';

import { getMessage, type MessageKey } from './messages';
import { resolveLocale } from './resolveLocale';
import { getStoredLocale, setStoredLocale } from './storage';
import type { AppLocale } from './types';

export interface I18nContextValue {
  locale: AppLocale;
  setLocale: (locale: AppLocale) => void;
  t: (key: MessageKey) => string;
}

export const I18nContext = createContext<I18nContextValue | null>(null);

function getInitialLocale(): AppLocale {
  return getStoredLocale() ?? resolveLocale(window.navigator.language);
}

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<AppLocale>(() => getInitialLocale());

  function setLocale(locale: AppLocale) {
    setLocaleState(locale);
    setStoredLocale(locale);
  }

  function t(key: MessageKey) {
    return getMessage(locale, key);
  }

  return (
    <I18nContext.Provider value={{ locale, setLocale, t }}>
      {children}
    </I18nContext.Provider>
  );
}
