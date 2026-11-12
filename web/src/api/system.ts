import { ApiError, apiDownloadCSV, apiRequest } from '../utils/request';
import type { AuditLogFilter, AuditLogPage, CurrentUserResponse, LDAPConfig, LDAPSyncChoices, LDAPSyncPreview, LDAPSyncResult, LoginResponse, SystemPermission, SystemRole, SystemUser, SystemRoleWithPermissions } from '../types/system';

export function login(username: string, password: string) {
  return apiRequest<LoginResponse>('/system/auth/login', {
    auth: 'none',
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export function logout() {
  return apiRequest<Record<string, never>>('/system/auth/logout', { method: 'POST' });
}

export function currentUser(signal?: AbortSignal) {
  return apiRequest<CurrentUserResponse>('/system/auth/me', { signal });
}

export function listUsers() {
  return apiRequest<{ list: SystemUser[]; total: number }>('/system/users');
}

export function updateUserStatus(id: number, status: string) {
  return apiRequest<SystemUser>(`/system/users/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  });
}

export function updateUserExternal(id: number, input: { expiresOn: string | null; sponsorId: number | null }) {
  return apiRequest<SystemUser>(`/system/users/${id}/external`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}

export function listRoles() {
  return apiRequest<{ list: SystemRoleWithPermissions[]; total: number }>('/system/roles');
}

export interface RoleInput {
  code?: string;
  name: string;
  description: string;
  approvable?: boolean;
}

export function createRole(input: RoleInput) {
  return apiRequest<SystemRole>('/system/roles', { method: 'POST', body: JSON.stringify(input) });
}

export function updateRole(id: number, input: RoleInput) {
  return apiRequest<SystemRole>(`/system/roles/${id}`, { method: 'PUT', body: JSON.stringify(input) });
}

export function deleteRole(id: number) {
  return apiRequest<Record<string, never>>(`/system/roles/${id}`, { method: 'DELETE' });
}

export function permissionTree() {
  return apiRequest<{ list: SystemPermission[] }>('/system/permissions/tree');
}

export function saveRolePermissions(id: number, permissionCodes: string[]) {
  return apiRequest<Record<string, never>>(`/system/roles/${id}/permissions`, {
    method: 'PUT',
    body: JSON.stringify({ permissionCodes }),
  });
}

export function getLDAPConfig() {
  return apiRequest<LDAPConfig>('/system/ldap/config');
}

export function saveLDAPConfig(config: LDAPConfig) {
  return apiRequest<LDAPConfig>('/system/ldap/config', {
    method: 'PUT',
    body: JSON.stringify(config),
  });
}

export function testLDAPConfig(config: LDAPConfig & { useSavedBindPassword?: boolean }) {
  return apiRequest<{ success: boolean; message: string }>('/system/ldap/test', {
    method: 'POST',
    body: JSON.stringify(config),
  });
}

export function previewLDAPSync() {
  return apiRequest<LDAPSyncPreview>('/system/ldap/sync/preview', { method: 'POST' });
}

export function executeLDAPSync(previewId: string, choices: LDAPSyncChoices) {
  return apiRequest<LDAPSyncResult>('/system/ldap/sync/execute', {
    method: 'POST',
    body: JSON.stringify({ previewId, ...choices }),
  });
}

function auditQuery(filters: AuditLogFilter, paginated: boolean) {
  const params = new URLSearchParams();
  const keys: (keyof AuditLogFilter)[] = ['actorUserId', 'action', 'resourceType', 'result', 'startTime', 'endTime', 'keyword'];
  if (paginated) keys.push('page', 'pageSize');
  for (const key of keys) {
    const value = filters[key];
    if (value !== undefined && value !== '') params.set(key, String(value));
  }
  const query = params.toString();
  return query ? `?${query}` : '';
}

function validAuditPage(value: unknown): value is AuditLogPage {
  if (typeof value !== 'object' || value === null) return false;
  const data = value as Record<string, unknown>;
  if (!Array.isArray(data.list) || !Number.isInteger(data.total) || (data.total as number) < 0
    || !Number.isInteger(data.page) || (data.page as number) < 1
    || !Number.isInteger(data.pageSize) || (data.pageSize as number) < 1 || (data.pageSize as number) > 200) return false;
  return data.list.every(row => typeof row === 'object' && row !== null
    && Number.isInteger(row.id) && row.id >= 0 && Number.isInteger(row.actorUserId) && row.actorUserId >= 0
    && ['actorUsername', 'action', 'resourceType', 'resourceId', 'result', 'ip', 'userAgent', 'detail', 'createdAt']
      .every(key => typeof row[key] === 'string'));
}

export async function listAuditLogs(filters: AuditLogFilter = {}, signal?: AbortSignal) {
  const data = await apiRequest<unknown>(`/system/audit-logs${auditQuery(filters, true)}`, { signal });
  if (!validAuditPage(data)) throw new ApiError('Invalid audit page data', 200, 'protocol');
  return data;
}

export function exportAuditLogs(filters: AuditLogFilter = {}, signal?: AbortSignal) {
  return apiDownloadCSV(`/system/audit-logs/export${auditQuery(filters, false)}`, { signal });
}
