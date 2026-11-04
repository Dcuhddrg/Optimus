import { afterEach, expect, test, vi } from 'vitest';
import { apiDownloadCSV } from './client';

afterEach(() => vi.unstubAllGlobals());

test('CSV download reports GVA business failure from JSON body', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
    code: 7,
    data: {},
    msg: 'export failed',
  }), { status: 200, headers: { 'Content-Type': 'application/json' } })));

  await expect(apiDownloadCSV('/system/audit-logs/export')).rejects.toMatchObject({
    kind: 'business', code: 7, message: 'export failed',
  });
});
