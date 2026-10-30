export interface SystemUser {
  id: number;
  uuid: string;
  username: string;
  displayName: string;
  email: string;
  source: 'local' | 'ldap' | string;
  status: 'enabled' | 'disabled' | string;
  roles?: string[];
}

export interface CurrentUserResponse {
  user: SystemUser;
  roles: string[];
  menuPermissions: string[];
}

export interface LoginResponse extends CurrentUserResponse {
  token: string;
}

export interface SystemRole {
  id: number;
  code: string;
  name: string;
  description: string;
  status: string;
}

export interface SystemPermission {
  id?: number;
  code: string;
  name: string;
  type: 'menu' | 'api' | string;
  method?: string;
  path?: string;
  parentCode?: string;
  sort?: number;
}

export interface LDAPConfig {
  enabled: boolean;
  url: string;
  baseDn: string;
  bindDn: string;
  bindPassword?: string;
  hasBindPassword: boolean;
  userFilter: string;
  usernameAttr: string;
  displayNameAttr: string;
  emailAttr: string;
}

export interface AuditLog {
  id: number;
  actorUserId: number;
  actorUsername: string;
  action: string;
  resourceType: string;
  resourceId: string;
  result: string;
  ip: string;
  userAgent: string;
  detail: string;
  createdAt: string;
}

export interface AuditLogFilter {
  actorUserId?: number;
  action?: string;
  resourceType?: string;
  result?: string;
  startTime?: string;
  endTime?: string;
  /** Matches the recorded detail text, not usernames or all columns. */
  keyword?: string;
  page?: number;
  pageSize?: number;
}

export interface AuditLogPage {
  list: AuditLog[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ListResponse<T> {
  items: T[];
  total: number;
}
