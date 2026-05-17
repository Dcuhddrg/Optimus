import { apiRequest } from './client';
import type {
  CloudAccount,
  CloudAssetFilters,
  CloudAssetPage,
  CreateCloudAccountPayload,
  UpdateCloudAccountTagsPayload,
} from '../types/cloud';
import type { ListResponse } from '../types/system';

function buildTagQuery(tags: string[]): string {
  if (tags.length === 0) {
    return '';
  }
  return `?${tags.map((tag) => `tag=${encodeURIComponent(tag)}`).join('&')}`;
}

function buildAssetQuery(filters: CloudAssetFilters): string {
  const params = new URLSearchParams();
  if (filters.accountId !== undefined) params.set('accountId', String(filters.accountId));
  if (filters.provider) params.set('provider', filters.provider);
  if (filters.region) params.set('region', filters.region);
  if (filters.resourceType) params.set('resourceType', filters.resourceType);
  if (filters.status) params.set('status', filters.status);
  if (filters.keyword) params.set('keyword', filters.keyword);
  if (filters.page !== undefined) params.set('page', String(filters.page));
  if (filters.pageSize !== undefined) params.set('pageSize', String(filters.pageSize));
  if (filters.tag) {
    for (const tag of filters.tag) {
      params.append('tag', tag);
    }
  }
  const qs = params.toString();
  return qs ? `?${qs}` : '';
}

export function listCloudAccounts(tagFilters: string[] = []) {
  return apiRequest<ListResponse<CloudAccount>>(`/cloud/accounts${buildTagQuery(tagFilters)}`);
}

export function createCloudAccount(payload: CreateCloudAccountPayload) {
  return apiRequest<CloudAccount>('/cloud/accounts', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function deleteCloudAccount(id: number) {
  return apiRequest<{ message: string }>(`/cloud/accounts/${id}`, { method: 'DELETE' });
}

export function syncCloudAccount(id: number) {
  return apiRequest<{ upserted: number; errors: number }>(`/cloud/accounts/${id}/sync`, {
    method: 'POST',
  });
}

export function updateCloudAccountTags(id: number, payload: UpdateCloudAccountTagsPayload) {
  return apiRequest<CloudAccount>(`/cloud/accounts/${id}/tags`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function listCloudAssets(filters: CloudAssetFilters = {}) {
  return apiRequest<CloudAssetPage>(`/cloud/assets${buildAssetQuery(filters)}`);
}
