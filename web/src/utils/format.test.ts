import { describe, expect, it } from 'vitest';
import { formatDateTime } from './format';

describe('formatDateTime', () => {
  it('formats in the given locale with a 24-hour clock', () => {
    const value = new Date(2026, 10, 5, 14, 3, 9);
    expect(formatDateTime(value, 'zh-CN')).toBe(value.toLocaleString('zh-CN', {
      year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false,
    }));
    expect(formatDateTime(value.toISOString(), 'en-US')).toContain('14:03:09');
  });

  it.each([null, undefined, ''])('returns the fallback for %j', (value) => {
    expect(formatDateTime(value)).toBe('-');
    expect(formatDateTime(value, 'en-US', 'not recorded')).toBe('not recorded');
  });

  it('returns unparseable input unchanged', () => {
    expect(formatDateTime('not-a-date', 'en-US')).toBe('not-a-date');
  });
});
