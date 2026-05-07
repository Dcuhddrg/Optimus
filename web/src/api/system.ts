import { apiRequest } from './client';
import type {
  AuditLog,
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
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export function logout() {
  return apiRequest<{ message: string }>('/system/auth/logout', { method: 'POST' });
}

export function currentUser() {
  return apiRequest<CurrentUserResponse>('/system/auth/me');
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

export function listAuditLogs() {
  return apiRequest<ListResponse<AuditLog>>('/system/audit-logs');
}
