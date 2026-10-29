import { apiDownloadCSV, apiRequest } from './client';
import type {
  AuditLogFilter,
  AuditLogPage,
  CurrentUserResponse,
  LDAPConfig,
  ListResponse,
  LoginResponse,
  SystemPermission,
  SystemRole,
  SystemUser,
} from '../types/system';

export function login(username: string, password: string) {
  return apiRequest<LoginResponse>('/system/auth/login', {
    auth: 'none',
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export function logout() {
  return apiRequest<{ message: string }>('/system/auth/logout', { method: 'POST' });
}

export function currentUser(signal?: AbortSignal) {
  return apiRequest<CurrentUserResponse>('/system/auth/me', { signal });
}

export function listUsers() {
  return apiRequest<ListResponse<SystemUser>>('/system/users');
}

export function updateUserStatus(id: number, status: string) {
  return apiRequest<SystemUser>(`/system/users/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  });
}

export function listRoles() {
  return apiRequest<ListResponse<SystemRole>>('/system/roles');
}

export function permissionTree() {
  return apiRequest<ListResponse<SystemPermission>>('/system/permissions/tree');
}

export function saveRolePermissions(id: number, permissionCodes: string[]) {
  return apiRequest<{ message: string }>(`/system/roles/${id}/permissions`, {
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

export function listAuditLogs(filters: AuditLogFilter = {}, signal?: AbortSignal) {
  return apiRequest<AuditLogPage>(`/system/audit-logs${auditQuery(filters, true)}`, { signal });
}

export function exportAuditLogs(filters: AuditLogFilter = {}, signal?: AbortSignal) {
  return apiDownloadCSV(`/system/audit-logs/export${auditQuery(filters, false)}`, { signal });
}
