import { afterEach, expect, test, vi } from 'vitest';
import { ApiError } from './client';
import { listProjects } from './project';

afterEach(() => vi.unstubAllGlobals());

test('project list unwraps the GVA response', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
    code: 0,
    data: { list: [{ id: 1, code: 'demo', name: 'Demo' }] },
    msg: '成功',
  }), { status: 200 })));

  await expect(listProjects()).resolves.toMatchObject({ list: [{ code: 'demo' }] });
});

test('project business failure rejects even with HTTP 200', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
    code: 7,
    data: {},
    msg: 'project not found',
  }), { status: 200 })));

  await expect(listProjects()).rejects.toMatchObject<ApiError>({
    message: 'project not found',
    kind: 'business',
    code: 7,
  });
});
