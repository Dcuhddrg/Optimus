import { apiRequest } from './client';
import type {
  CloudAccount,
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

export function listCloudAssets(tagFilters: string[] = []) {
  return apiRequest<CloudAssetPage>(`/cloud/assets${buildTagQuery(tagFilters)}`);
}
