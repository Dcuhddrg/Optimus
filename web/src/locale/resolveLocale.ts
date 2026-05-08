import type { AppLocale } from './types';

// 'zh-CN' -> 'zh-CN'
// 'zh-TW' -> 'zh-CN'
// 'en-US' -> 'en'
// undefined -> 'en'
export function resolveLocale(input?: string | null): AppLocale {
  const normalizedInput = input?.trim().toLowerCase();

  if (!normalizedInput) {
    return 'en';
  }

  if (normalizedInput.startsWith('zh')) {
    return 'zh-CN';
  }

  return 'en';
}
