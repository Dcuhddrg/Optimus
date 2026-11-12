export interface SystemUser {
  id: number;
  uuid: string;
  username: string;
  displayName: string;
  email: string;
  // LDAP value as is; display only (design 3.3 D4).
  department?: string;
  // Bound to a directory identity; a legacy LDAP account without it awaits
  // admin-confirmed binding and cannot sign in.
  bound?: boolean;
  source: 'local' | 'ldap' | string;
  status: 'enabled' | 'disabled' | string;
  roles?: string[];
  // Design 3.6: from the directory's employee type, read-only.
  personType?: 'employee' | 'external' | string;
  // External accounts only, set by an administrator. expiresOn is the last
  // valid day (YYYY-MM-DD). Without both the account has no permissions.
  expiresOn?: string | null;
  sponsorId?: number | null;
  // User list only: the sponsor's username, whether the sponsor is no longer
  // an enabled employee, and whether the account is past its expiry.
  sponsor?: string;
  sponsorInvalid?: boolean;
  expired?: boolean;
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
  // Can be chosen as an approval level (design 3.5).
  approvable: boolean;
}

export interface SystemRoleWithPermissions extends SystemRole {
  permissionCodes: string[];
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
  // One of IDENTITY_ATTRIBUTES; fixed while boundUsers > 0.
  idAttr: string;
  // Empty: departments are not synced.
  departmentAttr: string;
  // Person type attribute; entries with a value in externalTypeValues are
  // external (design 3.6). An empty list means nobody is external.
  employeeTypeAttr: string;
  externalTypeValues: string[];
  boundUsers?: number;
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

// LDAP full sync (design 3.3 D6).
export type LDAPSyncCategory =
  | 'create'
  | 'update'
  | 'autoDisable'
  | 'unconfirmable'
  | 'outOfScope'
  | 'pendingBinding'
  | 'conflict'
  | 'disabledInfo'
  | 'unidentifiable';

export interface LDAPSyncAccount {
  id: number;
  username: string;
  displayName: string;
  email: string;
  department: string;
  externalId: string;
  status: string;
  createdAt: string;
  roles: string[] | null;
}

export interface LDAPSyncEntry {
  dn: string;
  username: string;
  displayName: string;
  email: string;
  department: string;
  externalId: string;
  createdAt: string | null;
}

export interface LDAPSyncItem {
  category: LDAPSyncCategory;
  reason?: string;
  account?: LDAPSyncAccount;
  entry?: LDAPSyncEntry;
  changes?: { field: string; before: string; after: string }[];
  // Pending binding whose directory entry is newer than the account (or of
  // unknown age): likely a reused uid, confirm one by one.
  suspicious?: boolean;
  // Set when an update turns an employee into an external person: these
  // roles stop applying and these projects need another owner.
  restrictedRoles?: string[];
  ownedProjects?: string[];
}

export interface LDAPSyncPreview {
  previewId: string;
  expiresAt: string;
  disablesBlocked: boolean;
  configChanged: boolean;
  items: LDAPSyncItem[];
}

export interface LDAPSyncChoices {
  disable: number[];
  bind: number[];
  separate: number[];
}

export interface LDAPSyncOutcome {
  action: string;
  accountId?: number;
  username: string;
  outcome: 'done' | 'skipped' | 'error' | 'unprocessed';
  reason?: string;
}

export interface LDAPSyncResult {
  status: 'completed' | 'partial';
  outcomes: LDAPSyncOutcome[];
}
