export type AppLocale = 'zh-CN' | 'en';

export const SUPPORTED_LOCALES = ['zh-CN', 'en'] as const;

export const LOCALE_STORAGE_KEY = 'optimus.locale';

export const PLACEHOLDER_TRANSLATION_KEYS = [
  'common.locale.english',
  'common.locale.simplifiedChinese',
] as const;

export type TranslationKey = (typeof PLACEHOLDER_TRANSLATION_KEYS)[number];
